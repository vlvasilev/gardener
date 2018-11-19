{{- define "jvm.memory" -}}
{{- $r := $.resources.requests.memory -}}
{{- $base := $.jvmHeapBase -}}
{{ printf "%d%s" ( add $base ( mul ( div $.objectCount $r.weight ) 79 $r.weight ) ) "m" }}
{{- end -}}

{{- define "es.master" -}}
{{ printf "%d" ( add 1 (div $.elasticsearchMasterReplicas 2 ) )  }}
{{- end -}}

{{- define "es.master.pdb" -}}
{{ printf "%d" ( div $.elasticsearchMasterReplicas 2 )  }}
{{- end -}}