# Logging stack

### Why do we need logging stack?
We need logging stack because currently Kubernetes use the underlying container runtime logging which does not persist log for not existing containers. This makes difficult to investigate issues ocurred in long ago destroyed containers.
Our logging stack is created to keep logs for 14 days. Also there are pretty UI which can be used for easier log filtering.


### Architecture:
* Fluent-bit Deamonset which works like a log collector. Also we use custom Golang plugin which spreads log messages to their Loki instances
* One Loki Statefulset in the `garden` namespace which contains logs for the seed cluster and one per shoot namespace which contains logs for shoot's controlplane
* One Grafana Deployment in `garden` namespace and two Deployments per shoot namespace(one exposed to the end users and one for the operators). Grafana is the UI component used in the logging stack.

### How to access the logs
We need to authenticate first in front of the Grafana ingress (Can be found in the Garden Dashboard in `Monitoring and Logging` section)
After that there are two options for monitoring logs.
* The first one is to log in in the admin panel (bottom left corner). The default username and password are `admin`, `admin`. Now we can choose `Explore` menu (Left side of the screen) and after that we can create custom filters based on the log labels supported in Loki and their values. For example: 
```{pod_name='prometheus-0'}```
or with regex:
```{pod_name=~'prometheus.+'}```

* The other option is to go to the `Dashboards` panel. As mentioned before, we have a custom dashboards for pod logs. There we have one selector field for `pod_name` and one search field where we can search for particular string in the log messages. The following dashboards can be used for logs:

  * Garden Grafana
    * Pod Logs
    * Extensions
    * Systemd Logs
  * User Grafana
    * Kubernetes Control Plane Status
  * Operator Grafana 
    * Kubernetes Pods
    * Kubernetes Control Plane Status

### Configuration
#### Fluent-bit

We can modify the fluent-bit configuration from `charts/seed-bootstrap/charts/fluent-bit/templates/fluent-bit-configmap.yaml`
We can see that there are five different specifications
* SERVICE: Where is defined the server specifications
* INPUT: Where is defined the input stream of the logs
* OUTPUT: Where is defined the output source (Loki for example)
* FILTER: Where we filter logs with specific key
* PARSER: Which is used by filters for parsing the log message

All of the components have custom filters and parsers defined for parsing their log messages correctly. Currently in the extensions we create configmap under /charts directory called `logging-config` where their filters and parsers are. The fluent-bit reads this type of config files.

***When we create a new extension we need to create such a configmap and to put there filters and parsers for it. Otherwise the log messages from the extension will not be parsed correctly***

#### Loki
We can modify the Loki configuration from `charts/seed-bootstrap/charts/loki/templates/loki-configmap.yaml`

The main sections we have to know are:

* Index configuration: Currently we have the following Index configuration:
```
    schema_config:
      configs:
      - from: 2018-04-15
        store: boltdb
        object_store: filesystem
        schema: v11
        index:
          prefix: index_
          period: 24h
```
  * `from`: is the date from which we start collecting logs. Currently we use 2018.04.15 because it is in the past.
  * `store`: The DB used for storing the index.
  * `object_store`: Where the data is stored
  * `schema`: Schema version which should use (v11 is currently recommended)
  * `index.prefix`: The prefix for the Index.
  * `index.period`: The period for updating the indices

***If we want to create a new index config we have to add a new config block. `from` field should start from the current day + previous `index.period` and should not overlap with the current index. The `prefix` also should be different***
```
    schema_config:
      configs:
      - from: 2018-04-15
        store: boltdb
        object_store: filesystem
        schema: v11
        index:
          prefix: index_
          period: 24h
      - from: 2020-06-18
        store: boltdb
        object_store: filesystem
        schema: v11
        index:
          prefix: index_new_
          period: 24h
```

* chunk_store_config Configuration
```
    chunk_store_config: 
      max_look_back_period: 336h
```
***`chunk_store_config.max_look_back_period` should be the same as the `retention_period`***

* table_manager Configuration
```
    table_manager:
      retention_deletes_enabled: true
      retention_period: 336h
```
`table_manager.retention_period` is the living time for each log message. Loki will keep messages for sure for (`table_manager.retention_period` - `index.period`) time due to specification in the Loki implementation.

#### Grafana
We can modify the Grafana configuration from `charts/seed-bootstrap/charts/templates/grafana/grafana-datasources-configmap.yaml` and 
`charts/seed-monitoring/charts/grafana/tempates/grafana-datasources-configmap.yaml`

This is the Loki configuration we are currently use:

```
    - name: loki
      type: loki
      access: proxy
      url: http://loki.{{ .Release.Namespace }}.svc:3100
      jsonData:
        maxLines: 5000
```

* `name`: is the name of the datasource
* `type`: is the type of the datasource
* `access`: should be set to proxy
* `url`: Loki's url
* `svc`: Loki's port
* `jsonData.maxLines`: The limit of the log messages which Grafana will show to the users.

***Decrease this value if the browser works slowly!***

