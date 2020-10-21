apiVersion: v1
data:
  add_tag_to_record.lua: |-
    function add_tag_to_record(tag, timestamp, record)
      record["tag"] = tag
      return 1, timestamp, record
    end
  filter-kubernetes.conf: |-
    # Systemd Filters
    [FILTER]
        Name record_modifier
        Match journald.docker
        Record hostname ${NODE_NAME}
        Record unit docker

    [FILTER]
        Name record_modifier
        Match journald.containerd
        Record hostname ${NODE_NAME}
        Record unit containerd

    [FILTER]
        Name record_modifier
        Match journald.kubelet
        Record hostname ${NODE_NAME}
        Record unit kubelet

    [FILTER]
        Name record_modifier
        Match journald.cloud-config-downloader*
        Record hostname ${NODE_NAME}
        Record unit cloud-config-downloader

    [FILTER]
        Name record_modifier
        Match journald.docker-monitor
        Record hostname ${NODE_NAME}
        Record unit docker-monitor

    [FILTER]
        Name record_modifier
        Match journald.containerd-monitor
        Record hostname ${NODE_NAME}
        Record unit containerd-monitor

    [FILTER]
        Name record_modifier
        Match journald.kubelet-monitor
        Record hostname ${NODE_NAME}
        Record unit kubelet-monitor

    # Shoot controlplane filters
    [FILTER]
        Name                parser
        Match               kubernetes.*kube-apiserver*kube-apiserver*
        Key_Name            log
        Parser              kubeapiserverParser
        Reserve_Data        True

    [FILTER]
        Name                parser
        Match               kubernetes.*kube-apiserver*vpn-seed*
        Key_Name            log
        Parser              vpnshootParser
        Reserve_Data        True

    [FILTER]
        Name                parser
        Match               kubernetes.*etcd*etcd*
        Key_Name            log
        Parser              etcdeventsParser
        Reserve_Data        True

    [FILTER]
        Name                parser
        Match               kubernetes.*etcd*backup-restore*
        Key_Name            log
        Parser              gsacParser
        Reserve_Data        True

    [FILTER]
        Name                parser
        Match               kubernetes.*kube-state-metrics*kube-state-metrics*
        Key_Name            log
        Parser              kubeapiserverParser
        Reserve_Data        True

    [FILTER]
        Name                parser
        Match               kubernetes.*addons-kubernetes-dashboard*kubernetes-dashboard*
        Key_Name            log
        Parser              kubernetesdashboardParser
        Reserve_Data        True

    # System components filters
    [FILTER]
        Name                parser
        Match               kubernetes.*kube-proxy*kube-proxy*
        Key_Name            log
        Parser              kubeapiserverParser
        Reserve_Data        True

    [FILTER]
        Name                parser
        Match               kubernetes.*addons-nginx-ingress-controller*nginx-ingress-controller*
        Key_Name            log
        Parser              kubeapiserverParser
        Reserve_Data        True

    [FILTER]
        Name                parser
        Match               kubernetes.*metrics-server*metrics-server*
        Key_Name            log
        Parser              kubeapiserverParser
        Reserve_Data        True

    [FILTER]
        Name                parser
        Match               kubernetes.*vpn-shoot*vpn-shoot*
        Key_Name            log
        Parser              vpnshootParser
        Reserve_Data        True

    [FILTER]
        Name                parser
        Match               kubernetes.*node-exporter*node-exporter*
        Key_Name            log
        Parser              nodeexporterParser
        Reserve_Data        True

    [FILTER]
        Name                parser
        Match               kubernetes.*node-problem-detector*node-problem-detector*
        Key_Name            log
        Parser              kubeapiserverParser
        Reserve_Data        True

    [FILTER]
        Name                parser
        Match               kubernetes.*coredns*coredns*
        Key_Name            log
        Parser              corednsParser
        Parser              kubeapiserverParser
        Reserve_Data        True

    # Garden filters
    [FILTER]
        Name                parser
        Match               kubernetes.*alertmanager*alertmanager*
        Key_Name            log
        Parser              alertmanagerParser
        Reserve_Data        True

    [FILTER]
        Name                parser
        Match               kubernetes.*gardener-seed-admission-controller*gardener-seed-admission-controller*
        Key_Name            log
        Parser              gsacParser
        Reserve_Data        True

    [FILTER]
        Name                parser
        Match               kubernetes.*prometheus*prometheus*
        Key_Name            log
        Parser              alertmanagerParser
        Reserve_Data        True

    [FILTER]
        Name                parser
        Match               kubernetes.*grafana*grafana*
        Key_Name            log
        Parser              grafanaParser
        Reserve_Data        True

    [FILTER]
        Name                parser
        Match               kubernetes.*vpa-*
        Key_Name            log
        Parser              kubeapiserverParser
        Reserve_Data        True

    [FILTER]
        Name                parser
        Match               kubernetes.*hvpa-controller*hvpa-controller*
        Key_Name            log
        Parser              kubeapiserverParser
        Reserve_Data        True

    [FILTER]
        Name                parser
        Match               kubernetes.*dependency-watchdog*dependency-watchdog*
        Key_Name            log
        Parser              kubeapiserverParser
        Reserve_Data        True

    [FILTER]
        Name                parser
        Match               kubernetes.*gardener-resource-manager*gardener-resource-manager*
        Key_Name            log
        Parser              gardenerResourceManagerParser
        Reserve_Data        True

    [FILTER]
        Name                modify
        Match               kubernetes.*gardener-resource-manager*gardener-resource-manager*
        Rename              level  severity
        Rename              msg    log
        Rename              logger source

    # Extension filters
    [FILTER]
        Name                parser
        Match               kubernetes.*gardener-extension*
        Key_Name            log
        Parser              extensionsParser
        Reserve_Data        True

    [FILTER]
        Name                modify
        Match               kubernetes.*gardener-extension*
        Rename              level  severity
        Rename              msg    log
        Rename              logger source

    # Extensions
    [FILTER]
        Name                parser
        Match               kubernetes.*cluster-autoscaler*cluster-autoscaler*
        Key_Name            log
        Parser              clusterAutoscalerParser
        Reserve_Data        True

    [FILTER]
        Name                parser
        Match               kubernetes.*kube-scheduler*kube-scheduler*
        Key_Name            log
        Parser              kubeSchedulerParser
        Reserve_Data        True

    [FILTER]
        Name                parser
        Match               kubernetes.*kube-controller-manager*kube-controller-manager*
        Key_Name            log
        Parser              kubeControllerManagerParser
        Reserve_Data        True

    [FILTER]
        Name          rewrite_tag
        Match         kubernetes.*
        Rule          $kubernetes['pod_name'] ^(kube-apiserver-?.+|cluster-autoscaler-?.+|kube-scheduler-?.+|kube-controller-manager-?.+)$ user-exposed.$TAG true
        Emitter_Name  re_emitted

    [FILTER]
        Name                parser
        Match               kubernetes.*calico-node*calico-node*
        Key_Name            log
        Parser              caliconodeParser
        Reserve_Data        True

    [FILTER]
        Name                parser
        Match               kubernetes.*cloud-controller-manager*aws-cloud-controller-manager*
        Key_Name            log
        Parser              kubeapiserverParser
        Reserve_Data        True

    [FILTER]
        Name                parser
        Match               kubernetes.*machine-controller-manager*aws-machine-controller-manager*
        Key_Name            log
        Parser              kubeapiserverParser
        Reserve_Data        True

    [FILTER]
        Name                parser
        Match               kubernetes.*lb-readvertiser*aws-lb-readvertiser*
        Key_Name            log
        Parser              lbReadvertiserParser
        Reserve_Data        True

    [FILTER]
        Name                parser
        Match               kubernetes.cloud-controller-manager*packet-cloud-controller-manager*
        Key_Name            log
        Parser              kubeapiserverParser
        Reserve_Data        True

    [FILTER]
        Name                parser
        Match               kubernetes.machine-controller-manager*packet-machine-controller-manager*
        Key_Name            log
        Parser              kubeapiserverParser
        Reserve_Data        True

    [FILTER]
        Name record_modifier
        Match *packet-cloud-controller-manager*
        Record type user


    # Scripts
    [FILTER]
        Name                lua
        Match               kubernetes.*
        script              modify_severity.lua
        call                cb_modify

    [FILTER]
        Name                lua
        Match               kubernetes.*
        script              add_tag_to_record.lua
        call                add_tag_to_record
  fluent-bit.conf: |-
    # Service section

    [SERVICE]
        Flush           30
        Daemon          Off
        Log_Level       info
        Parsers_File    parsers.conf
        HTTP_Server     On
        HTTP_Listen     0.0.0.0
        HTTP_PORT       2020


    @INCLUDE input.conf
    @INCLUDE filter-kubernetes.conf
    @INCLUDE output.conf
  input.conf: |-
    # Input section

    [INPUT]
        Name              tail
        Tag               kubernetes.*
        Path              /var/log/containers/*.log
        Exclude_Path      *_garden_fluent-bit-*.log,*_garden_loki-*.log
        Parser            docker
        DB                /var/log/flb_kube.db
        read_from_head    true
        DB.sync           full
        DB.locking        true
        Skip_Long_Lines   On
        Mem_Buf_Limit     30MB
        Refresh_Interval  10
        Ignore_Older      1800s

    [INPUT]
        Name            systemd
        Tag             journald.docker
        Path            /var/log/journal/
        Read_From_Tail  True
        Systemd_Filter  _SYSTEMD_UNIT=docker.service

    [INPUT]
        Name            systemd
        Tag             journald.kubelet
        Path            /var/log/journal/
        Read_From_Tail  True
        Systemd_Filter  _SYSTEMD_UNIT=kubelet.service

    [INPUT]
        Name            systemd
        Tag             journald.containerd
        Path            /var/log/journal/
        Read_From_Tail  True
        Systemd_Filter  _SYSTEMD_UNIT=containerd.service

    [INPUT]
        Name            systemd
        Tag             journald.cloud-config-downloader
        Path            /var/log/journal/
        Read_From_Tail  True
        Systemd_Filter  _SYSTEMD_UNIT=cloud-config-downloader.service

    [INPUT]
        Name            systemd
        Tag             journald.docker-monitor
        Path            /var/log/journal/
        Read_From_Tail  True
        Systemd_Filter  _SYSTEMD_UNIT=docker-monitor.service

    [INPUT]
        Name            systemd
        Tag             journald.containerd-monitor
        Path            /var/log/journal/
        Read_From_Tail  True
        Systemd_Filter  _SYSTEMD_UNIT=containerd-monitor.service

    [INPUT]
        Name            systemd
        Tag             journald.kubelet-monitor
        Path            /var/log/journal/
        Read_From_Tail  True
        Systemd_Filter  _SYSTEMD_UNIT=kubelet-monitor.service
  kubernetes_label_map.json: |-
    {
      "kubernetes": {"container_name":"container_name","docker_id":"docker_id","namespace_name":"namespace_name","pod_name":"pod_name"} ,
      "severity": "severity"
    }
  modify_severity.lua: |-
    function cb_modify(tag, timestamp, record)
      local unified_severity = cb_modify_unify_severity(record)

      if not unified_severity then
        return 0, 0, 0
      end

      return 1, timestamp, record
    end

    function cb_modify_unify_severity(record)
      local modified = false
      local severity = record["severity"]
      if severity == nil or severity == "" then
        return modified
      end

      severity = trim(severity):upper()

      if severity == "I" or severity == "INF" or severity == "INFO" then
        record["severity"] = "INFO"
        modified = true
      elseif severity == "W" or severity == "WRN" or severity == "WARN" or severity == "WARNING" then
        record["severity"] = "WARN"
        modified = true
      elseif severity == "E" or severity == "ERR" or severity == "ERROR" or severity == "EROR" then
        record["severity"] = "ERR"
        modified = true
      elseif severity == "D" or severity == "DBG" or severity == "DEBUG" then
        record["severity"] = "DBG"
        modified = true
      elseif severity == "N" or severity == "NOTICE" then
        record["severity"] = "NOTICE"
        modified = true
      elseif severity == "F" or severity == "FATAL" then
        record["severity"] = "FATAL"
        modified = true
      end

      return modified
    end

    function trim(s)
      return (s:gsub("^%s*(.-)%s*$", "%1"))
    end
  output.conf: |-
    # Output section

    [Output]
        Name gardenerloki
        Match kubernetes.*
        Url http://loki.garden.svc:3100/loki/api/v1/push
        LogLevel info
        BatchWait 40
        BatchSize 30720
        Labels {test="fluent-bit-go"}
        LineFormat json
        ReplaceOutOfOrderTS true
        DropSingleKey false
        AutoKubernetesLabels false
        LabelSelector gardener.cloud/role:shoot
        RemoveKeys kubernetes,stream,time,tag
        LabelMapPath /fluent-bit/etc/kubernetes_label_map.json
        DynamicHostPath {"kubernetes": {"namespace_name": "namespace"}}
        DynamicHostPrefix http://loki.
        DynamicHostSuffix .svc:3100/loki/api/v1/push
        DynamicHostRegex shoot-
        MaxRetries 3
        Timeout 10
        MinBackoff 30
        Buffer true
        BufferType dque
        QueueDir  /fluent-bit/buffers/operator
        QueueSegmentSize 300
        QueueSync normal
        QueueName gardener-kubernetes-operator
        FallbackToTagWhenMetadataIsMissing true
        TagKey tag
        DropLogEntryWithoutK8sMetadata true
        TenantID operator

    [Output]
        Name gardenerloki
        Match user-exposed.kubernetes.*
        Url http://loki.garden.svc:3100/loki/api/v1/push
        LogLevel info
        BatchWait 40
        BatchSize 30720
        Labels {test="fluent-bit-go", lang="Golang"}
        LineFormat json
        ReplaceOutOfOrderTS true
        DropSingleKey false
        AutoKubernetesLabels true
        LabelSelector gardener.cloud/role:shoot
        RemoveKeys kubernetes,stream,type,time,tag
        LabelMapPath /fluent-bit/etc/kubernetes_label_map.json
        DynamicHostPath {"kubernetes": {"namespace_name": "namespace"}}
        DynamicHostPrefix http://loki.
        DynamicHostSuffix .svc:3100/loki/api/v1/push
        DynamicHostRegex shoot-
        MaxRetries 3
        Timeout 10
        MinBackoff 30
        Buffer true
        BufferType dque
        QueueDir  /fluent-bit/buffers/user
        QueueSegmentSize 300
        QueueSync normal
        QueueName gardener-kubernetes-user
        FallbackToTagWhenMetadataIsMissing true
        TagKey tag
        DropLogEntryWithoutK8sMetadata true
        TenantID user

    [Output]
        Name gardenerloki
        Match journald.*
        Url http://loki.garden.svc:3100/loki/api/v1/push
        LogLevel info
        BatchWait 60
        BatchSize 30720
        Labels {test="fluent-bit-go"}
        LineFormat json
        ReplaceOutOfOrderTS true
        DropSingleKey false
        RemoveKeys kubernetes,stream,hostname,unit
        LabelMapPath /fluent-bit/etc/systemd_label_map.json
        MaxRetries 3
        Timeout 10
        MinBackoff 30
        Buffer true
        BufferType dque
        QueueDir  /fluent-bit/buffers
        QueueSegmentSize 300
        QueueSync normal
        QueueName gardener-journald

    [OUTPUT]
        Name file
        Match *
        Path /log
    
    [OUTPUT]
        Name  counter
        Match *logger*

  parsers.conf: |-
    # Custom parsers
    [PARSER]
        Name        docker
        Format      json
        Time_Key    time
        Time_Format %Y-%m-%dT%H:%M:%S.%L
        Time_Keep   On
        # Command      |  Decoder | Field | Optional Action
        # =============|==================|=================
        Decode_Field_As   escaped    log

    [PARSER]
        Name        kubeapiserverParser
        Format      regex
        Regex       ^(?<severity>\w)(?<time>\d{4} [^\s]*)\s+(?<pid>\d+)\s+(?<source>[^ \]]+)\] (?<log>.*)$
        Time_Key    time
        Time_Format %m%d %H:%M:%S.%L

    [PARSER]
        Name        etcdeventsParser
        Format      regex
        Regex       ^(?<time>\d{4}-\d{2}-\d{2}\s+[^ ]*)\s+(?<severity>\w+)\s+\|\s+(?<source>[^ :]*):\s+(?<log>.*)
        Time_Key    time
        Time_Format %Y-%m-%d %H:%M:%S.%L

    [PARSER]
        Name        alertmanagerParser
        Format      regex
        Regex       ^level=(?<severity>\w+)\s+ts=(?<time>\d{4}-\d{2}-\d{2}[Tt].*[zZ])\s+caller=(?<source>[^\s]*+)\s+(?<log>.*)
        Time_Key    time
        Time_Format %Y-%m-%dT%H:%M:%S.%L

    [PARSER]
        Name        corednsParser
        Format      regex
        Regex       ^(?<time>\d{4}-\d{2}-\d{2}[Tt].*[zZ])\s+\[(?<severity>\w*[^\]])\]\s+(?<log>.*)
        Time_Key    time
        Time_Format  %Y-%m-%dT%H:%M:%S.%L

    [PARSER]
        Name        vpnshootParser
        Format      regex
        Regex       ^(?<time>[^0-9]*\d{1,2}\s+[^\s]+\s+\d{4})\s+(?<log>.*)
        Time_Key    time
        Time_Format %a %b%t%d %H:%M:%S %Y

    [PARSER]
        Name        kubernetesdashboardParser
        Format      regex
        Regex       ^(?<time>\d{4}\/\d{2}\/\d{2}\s+[^\s]*)\s+(?<log>.*)
        Time_Key    time
        Time_Format %Y/%m/%d %H:%M:%S

    [PARSER]
        Name        nodeexporterParser
        Format      regex
        Regex       ^time="(?<time>\d{4}-\d{2}-\d{2}T[^"]*)"\s+level=(?<severity>\w+)\smsg="(?<log>.*)"\s+source="(?<source>.*)"
        Time_Key    time
        Time_Format %Y-%m-%dT%H:%M:%S.%L

    [PARSER]
        Name        gsacParser
        Format      regex
        Regex       ^time="(?<time>\d{4}-\d{2}-\d{2}T[^"]*)"\s+level=(?<severity>\w+)\smsg="(?<log>.*)"
        Time_Key    time
        Time_Format %Y-%m-%dT%H:%M:%S.%L

    [PARSER]
        Name        grafanaParser
        Format      regex
        Regex       ^t=(?<time>\d{4}-\d{2}-\d{2}T[^ ]*)\s+lvl=(?<severity>\w+)\smsg="(?<log>.*)"\s+logger=(?<source>.*)
        Time_Key    time
        Time_Format %Y-%m-%dT%H:%M:%S%z

    [PARSER]
        Name        gardenerResourceManagerParser
        Format      json
        Time_Key    ts
        Time_Format %Y-%m-%dT%H:%M:%S.%L

    [PARSER]
        Name        extensionsParser
        Format      json
        Time_Key    ts
        Time_Format %Y-%m-%dT%H:%M:%S.%L

    [PARSER]
        Name        clusterAutoscalerParser
        Format      regex
        Regex       ^(?<severity>\w)(?<time>\d{4} [^\s]*)\s+(?<pid>\d+)\s+(?<source>[^ \]]+)\] (?<log>.*)$
        Time_Key    time
        Time_Format %m%d %H:%M:%S.%L

    [PARSER]
        Name        kubeSchedulerParser
        Format      regex
        Regex       ^(?<severity>\w)(?<time>\d{4} [^\s]*)\s+(?<pid>\d+)\s+(?<source>[^ \]]+)\] (?<log>.*)$
        Time_Key    time
        Time_Format %m%d %H:%M:%S.%L

    [PARSER]
        Name        kubeControllerManagerParser
        Format      regex
        Regex       ^(?<severity>\w)(?<time>\d{4} [^\s]*)\s+(?<pid>\d+)\s+(?<source>[^ \]]+)\] (?<log>.*)$
        Time_Key    time
        Time_Format %m%d %H:%M:%S.%L

    [PARSER]
        Name        caliconodeParser
        Format      regex
        Regex       ^(?<time>\d{4}-\d{2}-\d{2}\s+[^ ]*)\s+\[(?<severity>\w*)\]\[(?<pid>\d+)\]\s+(?<source>[^:]*):\s+(?<log>.*)
        Time_Key    time
        Time_Format %Y-%m-%d %H:%M:%S.%L

    [PARSER]
        Name        lbReadvertiserParser
        Format      regex
        Regex       ^time="(?<time>\d{4}-\d{2}-\d{2}T[^"]*)"\s+level=(?<severity>\w+)\smsg="(?<log>.*)"
        Time_Key    time
        Time_Format %Y-%m-%dT%H:%M:%SZ
  plugin.conf: |-
    [PLUGINS]
        Path /fluent-bit/plugins/out_loki.so
  systemd_label_map.json: '{"hostname":"host_name","unit":"systemd_component"}'
kind: ConfigMap
metadata:
  labels:
    app: fluent-bit
    gardener.cloud/role: logging
    role: logging
  name: fluent-bit-config
  namespace: garden
