# Logging and Monitoring configuration

By default, Gardener deploys a central Prometheus, AlertManager, and Grafana instance into the `garden` namespace of all seed clusters.
Additionally, as part of the shoot reconciliation flow, it deploys a shoot-specific Prometheus, Grafana (and, if configured, an AlertManager) into the shoot namespace next to the other control plane components.

Configurable by the `Logging` feature gate in the `gardenlet` configuration it might also deploy a central fluent-bit and Loki Statefulset into the `garden` namespace of all seed clusters.
Additionally, as part of the shoot reconciliation flow, it might deploy a shoot-specific Loki into the shoot namespace next to the other control plane components.

## Monitoring

The central Prometheus instance in the `garden` namespace fetches metrics and data from all seed cluster nodes and all seed cluster pods.
It uses the federation concept to allow the shoot-specific instances to scrape exactly the metrics for the pods of the control plane they are responsible for.
This allows to only scrape the metrics for the nodes/pods once for the whole cluster, and to distribute them afterwards.

Extension controllers might deploy components as part of their reconciliation next to the shoot's control plane.
Examples for this would be a cloud-controller-manager or CSI controller deployments.
In some cases, the extensions want to submit scrape configuration, alerts, and/or dashboards for these components such that their metrics can be scraped by Gardener's Prometheus deployment(s), and later be visible in the Grafana dashboards.

### What's the approach to submit scrape configuration, alerts, and/or dashboards?

Before deploying the shoot-specific Prometheus instance Gardener will read all `ConfigMap`s in the shoot namespacewhich are labeled with `extensions.gardener.cloud/configuration=monitoring`.
Such `ConfigMap`s may contain four fields in their `data`:

* `scrape_config`: This field contains Prometheus scrape configuration for the component(s) and metrics that shall be scraped.
* `alerting_rules`: This field contains AlertManager rules for alerts that shall be raised.
* `dashboard_operators`: This field contains a Grafana dashboard in JSON that is only relevant for Gardener operators.
* `dashboard_users`: This field contains a Grafana dashboard in JSON that is only relevant for Gardener users (shoot owners).

**Example:** The `ControlPlane` controller might deploy a `cloud-controller-manager` into the shoot namespace, and it wants to submit some monitoring configuration.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: extension-controlplane-monitoring-ccm
  namespace: shoot--project--name
  labels:
    extensions.gardener.cloud/configuration: monitoring
data:
  scrape_config: |
    scrape_configs:
    - job_name: cloud-controller-manager
      scheme: https
      tls_config:
        insecure_skip_verify: true
        cert_file: /etc/prometheus/seed/prometheus.crt
        key_file: /etc/prometheus/seed/prometheus.key
      honor_labels: false
      kubernetes_sd_configs:
      - role: endpoints
        namespaces:
          names: [shoot--project--name]
      relabel_configs:
      - source_labels:
        - __meta_kubernetes_service_name
        - __meta_kubernetes_endpoint_port_name
        action: keep
        regex: cloud-controller-manager;metrics
      # common metrics
      - action: labelmap
        regex: __meta_kubernetes_service_label_(.+)
      - source_labels: [ __meta_kubernetes_pod_name ]
        target_label: pod
      metric_relabel_configs:
      - process_max_fds
      - process_open_fds

  alerting_rules:
    groups:
    - name: cloud-controller-manager.rules
      rules:
      - alert: CloudControllerManagerDown
        expr: absent(up{job="cloud-controller-manager"} == 1)
        for: 15m
        labels:
          service: cloud-controller-manager
          severity: critical
          type: seed
          visibility: all
        annotations:
          description: All infrastructure specific operations cannot be completed (e.g. creating load balancers or persistent volumes).
          summary: Cloud controller manager is down.

  dashboard_operators:
    <some-json-describing-a-grafana-dashboard-for-operators>

  dashboard_users:
    <some-json-describing-a-grafana-dashboard-for-users>
```

## Logging

The central fluent-bit instances in the `garden` namespace are parsing the logs from all containers in the seed cluster.
The shoot-specific instances extract only the logs for the pods of the control plane they are responsible for.
This allows to fetch the logs for the pods once for the whole cluster, and to distribute them afterwards.

Extension controllers could have custom filters and parsers defined for parsing their log messages correctly and for better visualization in the Grafana dashboards. 
There is a configmap under /charts directory called `logging-config` where their filters and parsers are. The fluent-bit reads this type of config files. 
You could find how to integrate new ones under the next section.

**When a new extension is produced, it would be advisable to add `logging-config` configmap with the extension-specific filters and parsers to integrate with the logging stack. Otherwise the log messages from the extension will not be parsed correctly**

:warning: As there is only the central fluent-bit daemonset and not a shoot-specific deployment like in the case of monitoring, see [above](#monitoring), the logging parse configuration must be only provided once and **not per shoot namespace**.
Also, as fluent-bit parses the logs based on the container name you should make sure that the container names inside your provider-specific pods are prefixed with your extension name.

### Submitting logging parse configuration?

To submit a new logging parse configuration, create a `ConfigMap` labeled `extensions.gardener.cloud/configuration=logging` with a single data field - `filter-kubernetes.conf`, containing the fluent-bit configuration for parsing container logs. Before deploying the central fluent-bit instances into the `garden` namespace, Gardener will read all `ConfigMap`s in `garden` with that label and apply the configuration encoded in them.
Such `ConfigMap`s may contain a single field in their `data`:

**Example:** The `Worker` controller might deploy a `machine-controller-manager` into the shoot namespace, and it wants to submit some logging parse configuration.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: extension-controlplane-logging-mcm
  namespace: garden
  labels:
    extensions.gardener.cloud/configuration: logging
data:
  filter-kubernetes.conf: |
    [FILTER]
        Name                parser
        Match               kubernetes.machine-controller-manager*openstack-machine-controller-manager*
        Key_Name            log
        Parser              kubeapiserverParser
        Reserve_Data        True
```

:information: It's a good idea to put the logging configuration into the Helm chart that also deploys the extension controller while the monitoring configuration can be part of the Helm chart/deployment routine that deploys the provider-specific component into the shoot namespace.

## References and additional resources

* [GitHub issue describing the concept](https://github.com/gardener/gardener/issues/1351)
* [Exemplary implementation (monitoring) for the GCP provider](https://github.com/gardener/gardener-extension-provider-gcp/blob/master/charts/internal/seed-controlplane/charts/cloud-controller-manager/templates/configmap-monitoring.yaml)
* [Exemplary implementation (logging) for the OpenStack provider](https://github.com/gardener/gardener-extension-provider-openstack/blob/master/charts/gardener-extension-provider-openstack/templates/configmap-logging.yaml)
