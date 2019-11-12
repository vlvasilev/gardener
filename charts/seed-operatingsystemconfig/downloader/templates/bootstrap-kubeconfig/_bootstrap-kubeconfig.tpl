{{- define "bootstrap-kubeconfig" -}}
---
apiVersion: v1
kind: Config
current-context: kubelet-bootstrap@default
clusters:
- cluster:
    certificate-authority: /var/lib/cloud-config-downloader/credentials/ca.crt
    server: {{ required ".Values.Server is required" .Values.server }}
  name: default
contexts:
- context:
    cluster: default
    user: kubelet-bootstrap
  name: kubelet-bootstrap@default
users:
- name: kubelet-bootstrap
  user:
    as-user-extra: {}
    token: "<<BOOTSTRAP_TOKEN>>"
{{- end -}}