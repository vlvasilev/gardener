// Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package networkpolicies

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/gardener/gardener/pkg/apis/garden/v1beta1"
	"github.com/gardener/gardener/pkg/client/kubernetes"
	"github.com/gardener/gardener/pkg/logger"
	. "github.com/gardener/gardener/test/integration/framework"
	. "github.com/gardener/gardener/test/integration/shoots"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	kubeconfig     = flag.String("kubeconfig", "", "the path to the kubeconfig  of the garden cluster that will be used for integration tests")
	shootName      = flag.String("shootName", "", "the name of the shoot we want to test")
	shootNamespace = flag.String("shootNamespace", "", "the namespace name that the shoot resides in")
	logLevel       = flag.String("verbose", "", "verbosity level, when set, logging level will be DEBUG")
	cleanup        = flag.Bool("cleanup", false, "deletes the newly created / existing test shoot after the test suite is done")
)

const (
	InitializationTimeout = 600 * time.Second
	FinalizationTimeout   = 1800 * time.Second
	APIServer             = "kube-apiserver"
)

func validateFlags() {

	if !StringSet(*kubeconfig) {
		Fail("you need to specify the correct path for the kubeconfig")
	}

	if !FileExists(*kubeconfig) {
		Fail("kubeconfig path does not exist")
	}
}

var _ = Describe("Network Policy Testing", func() {
	var (
		shootGardenerTest   *ShootGardenerTest
		shootTestOperations *GardenerTestOperation
		cloudProvider       v1beta1.CloudProvider
		shootAppTestLogger  *logrus.Logger
		namespaceName       string
	)

	CBeforeSuite(func(ctx context.Context) {
		// validate flags
		validateFlags()
		shootAppTestLogger = logger.AddWriter(logger.NewLogger(*logLevel), GinkgoWriter)

		if StringSet(*shootName) {
			var err error
			//make GardenClient from kubeconfig file
			shootGardenerTest, err = NewShootGardenerTest(*kubeconfig, nil, shootAppTestLogger)
			Expect(err).NotTo(HaveOccurred())
			//make shoot template with name and namespace
			shoot := &v1beta1.Shoot{ObjectMeta: metav1.ObjectMeta{Namespace: *shootNamespace, Name: *shootName}}
			//from the GardenClient extract the shoot seed and garden object as their clients in one struct -> shootTestOperations
			shootTestOperations, err = NewGardenTestOperation(ctx, shootGardenerTest.GardenClient, shootAppTestLogger, shoot)
			Expect(err).NotTo(HaveOccurred())
		}

		var err error
		//get the cloud provider object
		cloudProvider, err = shootTestOperations.GetCloudProvider()
		Expect(err).NotTo(HaveOccurred())
		//deploy namespace named "gardener-e2e-network-policies-"
		// with lable "gardener-e2e-test": "networkpolicies"
		ns, err := shootTestOperations.SeedClient.CreateNamespace(
			&corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "gardener-e2e-network-policies-",
					Labels: map[string]string{
						"gardener-e2e-test": "networkpolicies",
					},
				},
			}, true)

		Expect(err).NotTo(HaveOccurred())

		namespaceName = ns.GetName()
		// make busybox template
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "busybox",
				Namespace: namespaceName,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					corev1.Container{
						Args:      []string{"sh"},
						Image:     "busybox",
						Name:      "busybox",
						Stdin:     true,
						StdinOnce: true,
						TTY:       false,
					},
				},
			},
		}
		// deploy the pod busybox
		err = shootTestOperations.SeedClient.Client().Create(ctx, pod)
		Expect(err).NotTo(HaveOccurred())

	}, InitializationTimeout)

	CAfterSuite(func(ctx context.Context) {
		namespaces := &corev1.NamespaceList{}
		selector := &client.ListOptions{
			LabelSelector: labels.SelectorFromSet(labels.Set(map[string]string{
				"gardener-e2e-test": "networkpolicies",
			})),
		}
		//load all namepsace objects with label ("gardener-e2e-test": "networkpolicies") into namespaces variable
		err := shootTestOperations.SeedClient.Client().List(ctx, selector, namespaces)
		Expect(err).NotTo(HaveOccurred())
		//delete each of these namespaces
		for _, ns := range namespaces.Items {
			err = shootTestOperations.SeedClient.Client().Delete(ctx, &ns)
			if err != nil && !errors.IsConflict(err) {
				Expect(err).NotTo(HaveOccurred())
			}
		}

	}, FinalizationTimeout)

	Context("Components are selected by correct policies", func() {
		const (
			timeout = 10 * time.Second
		)
		var (
			assertMatchAllPolicies = func(labels.Selector, expectedPolicies sets.String) func(ctx context.Context) {
				return func(ctx context.Context) {

					matched := sets.NewString()
					var podLabelSet labels.Set

					By(fmt.Sprintf("Getting first running pod with selectors %q in namespace %q", podGetSelector, shootTestOperations.ShootSeedNamespace()))
					pod, err := shootTestOperations.GetFirstRunningPodWithLabels(ctx, podGetSelector, shootTestOperations.ShootSeedNamespace(), shootTestOperations.SeedClient)
					podLabelSet = pod.GetLabels()
					Expect(err).NotTo(HaveOccurred())

					By(fmt.Sprintf("Getting all network policies in namespace %q", shootTestOperations.ShootSeedNamespace()))
					list := &networkingv1.NetworkPolicyList{}
					err = shootTestOperations.SeedClient.Client().List(ctx, &client.ListOptions{Namespace: shootTestOperations.ShootSeedNamespace()}, list)
					Expect(err).ToNot(HaveOccurred())

					for _, netPol := range list.Items {
						netPolSelector, err := metav1.LabelSelectorAsSelector(&netPol.Spec.PodSelector)
						Expect(err).NotTo(HaveOccurred())

						if netPolSelector.Matches(podLabelSet) {
							matched.Insert(netPol.GetName())
						}
					}

					Expect(matched.List()).Should(ConsistOf(expectedPolicies.List()))
				}
			}
		)

		CIt("should be matched by deny-all", assertMatchAllPolicies(KubeAPIServerSelector, sets.NewString("deny-all")), timeout)

	})

	Context("Old Deprecated policies are removed", func() {

		const (
			deprecatedKubeAPIServerPolicy = "kube-apiserver-default"
			deprecatedMetadataAppPolicy   = "cloud-metadata-service-deny-blacklist-app"
			deprecatedMetadataRolePolicy  = "cloud-metadata-service-deny-blacklist-role"
			timeout                       = 10 * time.Second
		)
		var (
			assertPolicyIsGone = func(policyName string) func(ctx context.Context) {
				return func(ctx context.Context) {
					By(fmt.Sprintf("Getting network policy %q in namespace %q", policyName, shootTestOperations.ShootSeedNamespace()))
					getErr := shootTestOperations.SeedClient.Client().Get(ctx, types.NamespacedName{Name: policyName, Namespace: shootTestOperations.ShootSeedNamespace()}, &networkingv1.NetworkPolicy{})
					Expect(getErr).To(HaveOccurred())
					By("error is NotFound")
					Expect(errors.IsNotFound(getErr)).To(BeTrue())
				}
			}
		)

		CIt(deprecatedKubeAPIServerPolicy, assertPolicyIsGone(deprecatedKubeAPIServerPolicy), timeout)
		CIt(deprecatedMetadataAppPolicy, assertPolicyIsGone(deprecatedMetadataAppPolicy), timeout)
		CIt(deprecatedMetadataRolePolicy, assertPolicyIsGone(deprecatedMetadataRolePolicy), timeout)

	})

	Context("Block Ingress for other namespaces", func() {
		var (
			NetworkPolicyTimeout = 30 * time.Second

			assertConnectivity = func(selector labels.Selector, port string) func(ctx context.Context) {
				return func(ctx context.Context) {
					By("Checking for source Pod is running")
					err := shootTestOperations.WaitUntilPodIsRunning(ctx, "busybox", namespaceName, shootTestOperations.SeedClient)
					ExpectWithOffset(1, err).NotTo(HaveOccurred())

					By("Checking that target Pod is running")
					pod, err := shootTestOperations.GetFirstRunningPodWithLabels(ctx, selector, shootTestOperations.ShootSeedNamespace(), shootTestOperations.SeedClient)
					ExpectWithOffset(1, err).NotTo(HaveOccurred())

					By("Check for Pod IP")
					ip := pod.Status.PodIP
					ExpectWithOffset(1, ip).NotTo(BeEmpty())

					By("Executing connectivity command")
					r, err := kubernetes.NewPodExecutor(shootTestOperations.SeedClient.RESTConfig()).
						Execute(ctx, namespaceName, "busybox", "busybox", fmt.Sprintf("nc -v -z -w 2 %s %s", ip, port))
					ExpectWithOffset(1, err).To(HaveOccurred())
					bytes, err := ioutil.ReadAll(r)
					ExpectWithOffset(1, err).NotTo(HaveOccurred())

					By("Connection message is timed out")
					ExpectWithOffset(1, string(bytes)).To(ContainSubstring(fmt.Sprintf("nc: %s (%s:%s): Connection timed out", ip, ip, port)))
				}
			}
		)
		CBeforeEach(func(ctx context.Context) {

		}, NetworkPolicyTimeout)

		CIt("etcd-main", assertConnectivity(EtcdMainSelector, "2379"), NetworkPolicyTimeout)
		CIt("to etcd-events", assertConnectivity(EtcdEventsSelector, "2379"), NetworkPolicyTimeout)
		CIt("to cloud-controller-manager", assertConnectivity(CloudControllerManagerSelector, "10253"), NetworkPolicyTimeout)
		CIt("to elasticsearch-logging", assertConnectivity(ElasticSearchSelector, "9200"), NetworkPolicyTimeout)
		CIt("to grafana", assertConnectivity(GrafanaSelector, "3000"), NetworkPolicyTimeout)
		CIt("to kibana-logging", assertConnectivity(KibanaSelector, "5601"), NetworkPolicyTimeout)
		CIt("to kube-controller-manager", assertConnectivity(KubeControllerManagerSelector, "10252"), NetworkPolicyTimeout)
		CIt("to kube-scheduler", assertConnectivity(KubeSchedulerSelector, "10251"), NetworkPolicyTimeout)
		CIt("to kube-state-metrics-shoot", assertConnectivity(KubeStateMetricsShootSelector, "8080"), NetworkPolicyTimeout)
		CIt("to kube-state-metrics-seed", assertConnectivity(KubeStateMetricsSeedSelector, "8080"), NetworkPolicyTimeout)
		CIt("to machine-controller-manager", assertConnectivity(MachineControllerManagerSelector, "10258"), NetworkPolicyTimeout)
		CIt("to prometheus", assertConnectivity(PrometheusSelector, "9090"), NetworkPolicyTimeout)

	})

	Context("Network Policy Testing", func() {
		var (
			NetworkPolicyTimeout = 1 * time.Minute
			ExecNCOnAPIServer    = func(ctx context.Context, host, port string) error {
				_, err := shootTestOperations.PodExecByLabel(ctx, KubeAPIServerSelector, APIServer,
					fmt.Sprintf("apt-get update && apt-get -y install netcat && nc -z -w5 %s %s", host, port), shootTestOperations.ShootSeedNamespace(), shootTestOperations.SeedClient)

				return err
			}

			ItShouldAllowTrafficTo = func(name, host, port string) {
				CIt(fmt.Sprintf("%s should allow connections", name), func(ctx context.Context) {
					Expect(ExecNCOnAPIServer(ctx, host, port)).NotTo(HaveOccurred())
				}, NetworkPolicyTimeout)
			}

			ItShouldBlockTrafficTo = func(name, host, port string) {
				CIt(fmt.Sprintf("%s should allow connections", name), func(ctx context.Context) {
					Expect(ExecNCOnAPIServer(ctx, host, port)).To(HaveOccurred())
				}, NetworkPolicyTimeout)
			}
		)

		ItShouldAllowTrafficTo("seed apiserver/external connection", "kubernetes.default", "443")
		ItShouldAllowTrafficTo("shoot etcd-main", "etcd-main-client", "2379")
		ItShouldAllowTrafficTo("shoot etcd-events", "etcd-events-client", "2379")

		CIt("should allow traffic to the shoot pod range", func(ctx context.Context) {
			dashboardIP, err := shootTestOperations.GetDashboardPodIP(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(ExecNCOnAPIServer(ctx, dashboardIP, "8443")).NotTo(HaveOccurred())
		}, NetworkPolicyTimeout)

		CIt("should allow traffic to the shoot node range", func(ctx context.Context) {
			nodeIP, err := shootTestOperations.GetFirstNodeInternalIP(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(ExecNCOnAPIServer(ctx, nodeIP, "10250")).NotTo(HaveOccurred())
		}, NetworkPolicyTimeout)

		ItShouldBlockTrafficTo("seed kubernetes dashboard", "kubernetes-dashboard.kube-system", "443")
		ItShouldBlockTrafficTo("shoot grafana", "grafana", "3000")
		ItShouldBlockTrafficTo("shoot kube-controller-manager", "kube-controller-manager", "10252")
		ItShouldBlockTrafficTo("shoot cloud-controller-manager", "cloud-controller-manager", "10253")
		ItShouldBlockTrafficTo("shoot machine-controller-manager", "machine-controller-manager", "10258")

		CIt("should block traffic to the metadataservice", func(ctx context.Context) {
			if cloudProvider == v1beta1.CloudProviderAlicloud {
				Expect(ExecNCOnAPIServer(ctx, "100.100.100.200", "80")).To(HaveOccurred())
			} else {
				Expect(ExecNCOnAPIServer(ctx, "169.254.169.254", "80")).To(HaveOccurred())
			}
		}, NetworkPolicyTimeout)
	})
})
