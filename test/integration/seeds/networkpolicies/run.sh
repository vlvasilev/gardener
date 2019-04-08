#!/bin/sh

go test -kubeconfig $HOME/.kube/config \
        -shootName "local" \
        -shootNamespace "garden-dev" \
        -ginkgo.v \
        -ginkgo.progress
        # -ginkgo.focus="etcd"
