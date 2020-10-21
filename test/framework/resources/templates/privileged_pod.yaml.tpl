apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: privileged
  namespace: garden
spec:
  selector:                   
    matchLabels:                          
      app: privileged           
      role: logging
  template:
    metadata:
      labels:
        app: privileged           
        role: logging
    spec:
      tolerations:
      - effect: NoSchedule
        key: node-role.kubernetes.io/master
      - effect: NoExecute
        key: pool.worker.gardener.cloud/dedicated-for
        operator: Equal
        value: etcd      
      containers:
      - name: busybox
        image: busybox
        securityContext:
          privileged: true
        args:
        - /bin/sh
        - -c
        - |-
            #rm -f /var/log/flb_kube.db /var/log/flb_kube.db-shm /var/log/flb_kube.db-wal

            # Sleep forever to prevent restarts
            while true; do
                sleep 3600;
            done
        volumeMounts:
        - name: varlog
          mountPath: /var/log
      volumes:
      - name: varlog
        hostPath:
          path: /var/log