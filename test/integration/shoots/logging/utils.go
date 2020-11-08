// Copyright 2019 Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file.
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
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/client/kubernetes"
	"github.com/gardener/gardener/pkg/utils/retry"
	"github.com/gardener/gardener/test/framework"

	"github.com/onsi/ginkgo"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Checks whether required logging resources are present.
// If not, probably the logging feature gate is not enabled.
func hasRequiredResources(ctx context.Context, k8sSeedClient kubernetes.Interface) (bool, error) {
	fluentBit := &appsv1.DaemonSet{}
	if err := k8sSeedClient.DirectClient().Get(ctx, client.ObjectKey{Namespace: garden, Name: fluentBitName}, fluentBit); err != nil {
		return false, err
	}

	loki := &appsv1.StatefulSet{}
	if err := k8sSeedClient.DirectClient().Get(ctx, client.ObjectKey{Namespace: garden, Name: lokiName}, loki); err != nil {
		return false, err
	}

	return true, nil
}

func checkRequiredResources(ctx context.Context, k8sSeedClient kubernetes.Interface) {
	isLoggingEnabled, err := hasRequiredResources(ctx, k8sSeedClient)
	if !isLoggingEnabled {
		message := fmt.Sprintf("Error occurred checking for required logging resources in the seed %s namespace. Ensure that the logging feature gate is enabled: %s", garden, err.Error())
		ginkgo.Fail(message)
	}
}

// WaitUntilLokiReceivesLogs waits until the loki instance in <lokiNamespace> receives <expected> logs from <key>, <value>
func WaitUntilLokiReceivesLogs(ctx context.Context, interval time.Duration, f *framework.ShootFramework, lokiNamespace, key, value string, expected int, client kubernetes.Interface) error {
	return retry.Until(ctx, interval, func(ctx context.Context) (done bool, err error) {
		search, err := f.GetLokiLogs(ctx, lokiNamespace, key, value, client)
		if err != nil {
			return retry.SevereError(err)
		}
		var actual int
		for _, result := range search.Data.Result {
			currentStr, ok := result.Value[1].(string)
			if !ok {
				return retry.SevereError(fmt.Errorf("Data.Result.Value[1] is not a string for %s=%s", key, value))
			}
			current, err := strconv.Atoi(currentStr)
			if err != nil {
				return retry.SevereError(fmt.Errorf("Data.Result.Value[1] string is not parsable to intiger for %s=%s", key, value))
			}
			actual += current
		}

		if expected > actual {
			f.Logger.Infof("Waiting to receive %d logs, currently received %d", expected, actual)
			return retry.MinorError(fmt.Errorf("received only %d/%d logs", actual, expected))
		} else if expected < actual {
			return retry.SevereError(fmt.Errorf("expected to receive %d logs but was %d", expected, actual))
		}

		f.Logger.Infof("Received all of %d logs", actual)
		return retry.Ok()
	})
}

func encode(obj runtime.Object) []byte {
	data, _ := json.Marshal(obj)
	return data
}

func create(ctx context.Context, c client.Client, obj runtime.Object) error {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return err
	}
	accessor.SetResourceVersion("")
	err = c.Create(ctx, obj)
	if apierrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

func getShootNamesapce(number int) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s%v", simulatesShootNamespacePrefix, number),
		},
	}
}

func getCluster(number int) *extensionsv1alpha1.Cluster {
	shoot := &gardencorev1beta1.Shoot{
		Spec: gardencorev1beta1.ShootSpec{
			Hibernation: &gardencorev1beta1.Hibernation{
				Enabled: pointer.BoolPtr(false),
			},
			Purpose: (*gardencorev1beta1.ShootPurpose)(pointer.StringPtr("evaluation")),
		},
	}

	return &extensionsv1alpha1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Cluster",
			APIVersion: "extensions.gardener.cloud/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s%v", simulatesShootNamespacePrefix, number),
		},
		Spec: extensionsv1alpha1.ClusterSpec{
			Shoot: runtime.RawExtension{
				Raw: encode(shoot),
			},
			CloudProfile: runtime.RawExtension{
				Raw: encode(&gardencorev1beta1.CloudProfile{}),
			},
			Seed: runtime.RawExtension{
				Raw: encode(&gardencorev1beta1.Seed{}),
			},
		},
	}
}

func getLokiShootService(number int) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      lokiName,
			Namespace: fmt.Sprintf("%s%v", simulatesShootNamespacePrefix, number),
		},
		Spec: corev1.ServiceSpec{
			Type:         corev1.ServiceType(corev1.ServiceTypeExternalName),
			ExternalName: "loki.garden.svc.cluster.local",
			Ports: []corev1.ServicePort{
				{Port: 80},
			},
		},
	}
}
