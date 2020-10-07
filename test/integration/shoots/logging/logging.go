// Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

package logging

import (
	"context"
	"fmt"
	"time"

	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"
	"github.com/gardener/gardener/test/framework"
	"github.com/gardener/gardener/test/framework/resources/templates"

	"github.com/onsi/ginkgo"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"k8s.io/utils/pointer"
)

const (
	logsCount                 = 5000
	numberOfSimulatedClusters = 100

	initializationTimeout          = 5 * time.Minute
	getLogsFromLokiTimeout         = 15 * time.Minute
	loggerDeploymentCleanupTimeout = 5 * time.Minute

	fluentBitName                 = "fluent-bit"
	lokiName                      = "loki"
	garden                        = "garden"
	logger                        = "logger-.*"
	fluentBitConfigMapName        = "fluent-bit-config"
	fluentBitClusterRoleName      = "fluent-bit-read"
	simulatesShootNamespacePrefix = "shoot--test--"
	lokiConfigMapName             = "loki-config"
)

var _ = ginkgo.Describe("Seed logging testing", func() {

	f := framework.NewShootFramework(nil)
	gardenNamespace := &corev1.Namespace{}
	fluentBit := &appsv1.DaemonSet{}
	fluentBitConfMap := &corev1.ConfigMap{}
	fluentBitService := &corev1.Service{}
	fluentBitClusterRole := &rbacv1.ClusterRole{}
	fluentBitClusterRoleBinding := &rbacv1.ClusterRoleBinding{}
	fluentBitServiceAccount := &corev1.ServiceAccount{}
	clusterCRD := &apiextensionsv1.CustomResourceDefinition{}
	lokiSts := &appsv1.StatefulSet{}
	lokiServiceAccount := &corev1.ServiceAccount{}
	lokiService := &corev1.Service{}
	lokiConfMap := &corev1.ConfigMap{}

	framework.CBeforeEach(func(ctx context.Context) {
		checkRequiredResources(ctx, f.SeedClient)
		//Get the Fluent-Bit DaemonSet from the seed
		err := f.SeedClient.Client().Get(ctx, types.NamespacedName{Namespace: v1beta1constants.GardenNamespace, Name: fluentBitName}, fluentBit)
		framework.ExpectNoError(err)
		//Get the Fluent-Bit ConfigMap from the seed
		err = f.SeedClient.Client().Get(ctx, types.NamespacedName{Namespace: v1beta1constants.GardenNamespace, Name: fluentBitConfigMapName}, fluentBitConfMap)
		framework.ExpectNoError(err)
		//Get the Fluent-Bit Service from the seed
		err = f.SeedClient.Client().Get(ctx, types.NamespacedName{Namespace: v1beta1constants.GardenNamespace, Name: fluentBitName}, fluentBitService)
		framework.ExpectNoError(err)
		//Get the Fluent-Bit ClusterRole from the seed
		err = f.SeedClient.Client().Get(ctx, types.NamespacedName{Namespace: v1beta1constants.GardenNamespace, Name: fluentBitClusterRoleName}, fluentBitClusterRole)
		framework.ExpectNoError(err)
		//Get the Fluent-Bit Rolebinding from the seed
		err = f.SeedClient.Client().Get(ctx, types.NamespacedName{Namespace: v1beta1constants.GardenNamespace, Name: fluentBitClusterRoleName}, fluentBitClusterRoleBinding)
		framework.ExpectNoError(err)
		//Get the Fluent-Bit ServiceAccount from the seed
		err = f.SeedClient.Client().Get(ctx, types.NamespacedName{Namespace: v1beta1constants.GardenNamespace, Name: fluentBitName}, fluentBitServiceAccount)
		framework.ExpectNoError(err)
		//Get the cluster CRD from the seed
		err = f.SeedClient.Client().Get(ctx, types.NamespacedName{Namespace: "", Name: "clusters.extensions.gardener.cloud"}, clusterCRD)
		framework.ExpectNoError(err)
		//Get the Loki StS from the seed
		err = f.SeedClient.Client().Get(ctx, types.NamespacedName{Namespace: v1beta1constants.GardenNamespace, Name: lokiName}, lokiSts)
		framework.ExpectNoError(err)
		//Get the Loki ServiceAccount from the seed
		err = f.SeedClient.Client().Get(ctx, types.NamespacedName{Namespace: v1beta1constants.GardenNamespace, Name: lokiName}, lokiServiceAccount)
		framework.ExpectNoError(err)
		//Get the Loki Service from the seed
		err = f.SeedClient.Client().Get(ctx, types.NamespacedName{Namespace: v1beta1constants.GardenNamespace, Name: lokiName}, lokiService)
		framework.ExpectNoError(err)
		//Get the Loki ConfigMap from the seed
		err = f.SeedClient.Client().Get(ctx, types.NamespacedName{Namespace: v1beta1constants.GardenNamespace, Name: lokiConfigMapName}, lokiConfMap)
		framework.ExpectNoError(err)
	}, initializationTimeout)

	f.Beta().Serial().CIt("should get container logs from loki for all namespaces", func(ctx context.Context) {
		ginkgo.By("Deploy the garden Namespace")
		gardenNamespace = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: v1beta1constants.GardenNamespace,
			},
		}
		err := create(ctx, f.ShootClient.Client(), gardenNamespace)
		framework.ExpectNoError(err)

		ginkgo.By("Deploy the Loki StatefulSet")
		err = create(ctx, f.ShootClient.Client(), lokiServiceAccount)
		framework.ExpectNoError(err)
		err = create(ctx, f.ShootClient.Client(), lokiConfMap)
		framework.ExpectNoError(err)
		lokiService.Spec.ClusterIP = ""
		err = create(ctx, f.ShootClient.Client(), lokiService)
		framework.ExpectNoError(err)
		// Remove the Loki PVC as it is no needed for the test
		lokiSts.Spec.VolumeClaimTemplates = nil
		// Instead use an empty dir volume
		lokiDataVolumeSize := resource.MustParse("500Mi")
		lokiDataVolume := corev1.Volume{
			Name: "loki",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{
					SizeLimit: &lokiDataVolumeSize,
				},
			},
		}
		lokiSts.Spec.Template.Spec.Volumes = append(lokiSts.Spec.Template.Spec.Volumes, lokiDataVolume)
		for index, container := range lokiSts.Spec.Template.Spec.Containers {
			if container.Name == lokiName {
				r := corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("800m"),
						corev1.ResourceMemory: resource.MustParse("1.5Gi"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("900m"),
						corev1.ResourceMemory: resource.MustParse("2.5Gi"),
					},
				}
				lokiSts.Spec.Template.Spec.Containers[index].Resources = r
			}
		}
		err = create(ctx, f.ShootClient.Client(), lokiSts)
		framework.ExpectNoError(err)

		ginkgo.By("Wait until Loki StatefulSet is ready")
		err = f.WaitUntilStatefulSetIsRunning(ctx, lokiName, v1beta1constants.GardenNamespace, f.ShootClient)
		framework.ExpectNoError(err)

		ginkgo.By("Deploy the cluster CRD")
		clusterCRD.Spec.PreserveUnknownFields = false
		for version := range clusterCRD.Spec.Versions {
			clusterCRD.Spec.Versions[version].Schema.OpenAPIV3Schema.XPreserveUnknownFields = pointer.BoolPtr(true)
		}
		err = create(ctx, f.ShootClient.Client(), clusterCRD)
		framework.ExpectNoError(err)

		ginkgo.By("Deploy the fluent-bit RBAC")
		err = create(ctx, f.ShootClient.Client(), fluentBitServiceAccount)
		framework.ExpectNoError(err)
		err = create(ctx, f.ShootClient.Client(), fluentBitClusterRole)
		framework.ExpectNoError(err)
		err = create(ctx, f.ShootClient.Client(), fluentBitClusterRoleBinding)
		framework.ExpectNoError(err)

		ginkgo.By("Deploy the fluent-bit DaemonSet")
		err = create(ctx, f.ShootClient.Client(), fluentBitConfMap)
		framework.ExpectNoError(err)
		err = create(ctx, f.ShootClient.Client(), fluentBit)
		framework.ExpectNoError(err)

		ginkgo.By("Wait until fluent-bit DaemonSet is ready")
		err = f.WaitUntilDaemonSetIsRunning(ctx, f.ShootClient.Client(), fluentBitName, v1beta1constants.GardenNamespace)
		framework.ExpectNoError(err)

		ginkgo.By("Deploy the simulated cluster and shoot controlplane namespaces")
		for i := 0; i < numberOfSimulatedClusters; i++ {
			shootNamespace := getShootNamesapce(i)
			ginkgo.By(fmt.Sprintf("Deploy namespace %s", shootNamespace.Name))
			err := create(ctx, f.ShootClient.Client(), shootNamespace)
			framework.ExpectNoError(err)
			_, err = kutil.TryUpdateNamespace(ctx, f.ShootClient.Kubernetes(), retry.DefaultBackoff, gardenNamespace.ObjectMeta, func(ns *corev1.Namespace) (*corev1.Namespace, error) {
				kutil.SetMetaDataLabel(&ns.ObjectMeta, "gardener.cloud", "shoot")
				return ns, nil
			})
			framework.ExpectNoError(err)

			cluster := getCluster(i)
			ginkgo.By(fmt.Sprintf("Deploy cluster %s", cluster.Name))
			err = create(ctx, f.ShootClient.DirectClient(), cluster)
			framework.ExpectNoError(err)

			ginkgo.By(fmt.Sprintf("Deploy the loki service in namespace %s", shootNamespace.Name))
			lokiShootService := getLokiShootService(i)
			err = create(ctx, f.ShootClient.Client(), lokiShootService)
			framework.ExpectNoError(err)

			ginkgo.By(fmt.Sprintf("Deploy the logger application in namespace %s", shootNamespace.Name))
			loggerParams := struct {
				HelmDeployNamespace string
				LogsCount           int
			}{
				shootNamespace.Name,
				logsCount,
			}
			err = f.RenderAndDeployTemplate(ctx, f.ShootClient, templates.LoggerAppName, loggerParams)
			framework.ExpectNoError(err)
		}

		loggerLabels := labels.SelectorFromSet(labels.Set(map[string]string{
			"app": logger,
		}))
		for i := 0; i < numberOfSimulatedClusters; i++ {
			shootNamespace := fmt.Sprintf("%s%v", simulatesShootNamespacePrefix, i)
			ginkgo.By(fmt.Sprintf("Wait until logger application is ready in namespace %s", shootNamespace))
			err := f.WaitUntilDeploymentsWithLabelsIsReady(ctx, loggerLabels, shootNamespace, f.ShootClient)
			framework.ExpectNoError(err)
		}

		ginkgo.By("Verify loki received logger application logs for all namespaces")
		err = WaitUntilLokiReceivesLogs(ctx, 1*time.Minute, f, v1beta1constants.GardenNamespace, "pod_name", logger, logsCount*numberOfSimulatedClusters, f.ShootClient)
		framework.ExpectNoError(err)

	}, getLogsFromLokiTimeout, framework.WithCAfterTest(func(ctx context.Context) {
		ginkgo.By("Cleaning up logger app resources")
		for i := 0; i < numberOfSimulatedClusters; i++ {
			shootNamespace := getShootNamesapce(i)
			loggerDeploymentToDelete := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: shootNamespace.Name,
					Name:      logger,
				},
			}
			err := kutil.DeleteObject(ctx, f.ShootClient.Client(), loggerDeploymentToDelete)
			framework.ExpectNoError(err)

			cluster := getCluster(i)
			err = kutil.DeleteObject(ctx, f.ShootClient.Client(), cluster)
			framework.ExpectNoError(err)

			lokiShootService := getLokiShootService(i)
			err = kutil.DeleteObject(ctx, f.ShootClient.Client(), lokiShootService)
			framework.ExpectNoError(err)

			err = kutil.DeleteObject(ctx, f.ShootClient.Client(), shootNamespace)
			framework.ExpectNoError(err)
		}

		ginkgo.By("Cleaning up garden namespace")
		objectsToDelete := []runtime.Object{
			fluentBit,
			fluentBitConfMap,
			fluentBitService,
			fluentBitClusterRole,
			fluentBitClusterRoleBinding,
			fluentBitServiceAccount,
			gardenNamespace,
		}
		for _, object := range objectsToDelete {
			err := kutil.DeleteObject(ctx, f.ShootClient.Client(), object)
			framework.ExpectNoError(err)
		}
	}, loggerDeploymentCleanupTimeout))
})
