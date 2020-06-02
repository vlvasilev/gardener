module github.com/gardener/gardener

go 1.14

require (
	github.com/Masterminds/semver v1.5.0
	github.com/ahmetb/gen-crd-api-reference-docs v0.1.5
	github.com/coreos/etcd v3.3.17+incompatible
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf
	github.com/gardener/etcd-druid v0.3.0
	github.com/gardener/external-dns-management v0.7.7
	github.com/gardener/gardener-resource-manager v0.10.0
	github.com/gardener/hvpa-controller v0.0.0-20191014062307-fad3bdf06a25
	github.com/gardener/machine-controller-manager v0.27.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.1.0
	github.com/go-openapi/spec v0.19.2
	github.com/gobuffalo/packr v1.30.1
	github.com/gogo/protobuf v1.3.1
	github.com/golang/mock v1.4.3
	github.com/googleapis/gnostic v0.3.1
	github.com/hashicorp/go-multierror v1.0.0
	github.com/huandu/xstrings v1.3.1
	github.com/json-iterator/go v1.1.9
	github.com/mholt/archiver v3.1.1+incompatible
	github.com/onsi/ginkgo v1.10.3
	github.com/onsi/gomega v1.7.1
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.5.0
	github.com/prometheus/common v0.9.1
	github.com/robfig/cron v1.2.0
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.6
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.6.1
	go.uber.org/zap v1.13.0
	golang.org/x/crypto v0.0.0-20200220183623-bac4c82f6975
	golang.org/x/lint v0.0.0-20200130185559-910be7a94367
	golang.org/x/tools v0.0.0-20200422205258-72e4a01eba43 // indirect
	gomodules.xyz/jsonpatch/v2 v2.0.1
	google.golang.org/grpc v1.26.0
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.17.0
	k8s.io/apiextensions-apiserver v0.17.0
	k8s.io/apimachinery v0.17.0
	k8s.io/apiserver v0.16.8
	k8s.io/autoscaler v0.0.0-20190805135949-100e91ba756e
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/cluster-bootstrap v0.16.8
	k8s.io/code-generator v0.16.8
	k8s.io/component-base v0.16.8
	k8s.io/gengo v0.0.0-20190826232639-a874a240740c
	k8s.io/helm v2.16.1+incompatible
	k8s.io/klog v1.0.0
	k8s.io/kube-aggregator v0.16.8
	k8s.io/kube-openapi v0.0.0-20191107075043-30be4d16710a
	k8s.io/kubelet v0.16.8
	k8s.io/metrics v0.16.8
	k8s.io/utils v0.0.0-20200327001022-6496210b90e8
	sigs.k8s.io/controller-runtime v0.4.0
	sigs.k8s.io/yaml v1.1.0
)

replace (
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.9.2
	k8s.io/api => k8s.io/api v0.16.8 // 1.16.8
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.16.8 // 1.16.8
	k8s.io/apimachinery => k8s.io/apimachinery v0.16.8 // 1.16.8
	k8s.io/apiserver => k8s.io/apiserver v0.16.8 // 1.16.8
	k8s.io/client-go => k8s.io/client-go v0.16.8 // 1.16.8
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.16.8 // 1.16.8
	k8s.io/code-generator => k8s.io/code-generator v0.16.8 // 1.16.8
	k8s.io/component-base => k8s.io/component-base v0.16.8 // 1.16.8
	k8s.io/helm => k8s.io/helm v2.13.1+incompatible
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.16.8 // 1.16.8
)
