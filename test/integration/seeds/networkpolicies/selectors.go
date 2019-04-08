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
	"k8s.io/apimachinery/pkg/labels"
)

var (
	// EtcdMainSelector is selector for etcd main.
	EtcdMainSelector = labels.SelectorFromSet(labels.Set{
		"app":                     "etcd-statefulset",
		"garden.sapcloud.io/role": "controlplane",
		"role":                    "main",
	})

	// EtcdEventsSelector is selector for etcd events.
	EtcdEventsSelector = labels.SelectorFromSet(labels.Set{
		"app":                     "etcd-statefulset",
		"garden.sapcloud.io/role": "controlplane",
		"role":                    "events",
	})

	// CloudControllerManagerSelector is selector for cloud-controller-manager.
	CloudControllerManagerSelector = labels.SelectorFromSet(labels.Set{
		"app":                     "kubernetes",
		"garden.sapcloud.io/role": "controlplane",
		"role":                    "cloud-controller-manager",
	})

	// ElasticSearchSelector is selector for ElasticSearch.
	ElasticSearchSelector = labels.SelectorFromSet(labels.Set{
		"app":                     "elasticsearch-logging",
		"garden.sapcloud.io/role": "logging",
		"role":                    "logging",
	})

	// GrafanaSelector is selector for Grafana.
	GrafanaSelector = labels.SelectorFromSet(labels.Set{
		"component":               "grafana",
		"garden.sapcloud.io/role": "monitoring",
	})

	// KibanaSelector is selector for Kibana.
	KibanaSelector = labels.SelectorFromSet(labels.Set{
		"app":                     "kibana-logging",
		"garden.sapcloud.io/role": "logging",
		"role":                    "logging",
	})

	// KubeAPIServerSelector is selector for Kubernetes API Server.
	KubeAPIServerSelector = labels.SelectorFromSet(labels.Set{
		"app":  "kubernetes",
		"role": "apiserver",
	})

	// KubeControllerManagerSelector is selector for Kubernetes Controller Manager.
	KubeControllerManagerSelector = labels.SelectorFromSet(labels.Set{
		"app":                     "kubernetes",
		"garden.sapcloud.io/role": "controlplane",
		"role":                    "controller-manager",
	})

	// KubeSchedulerSelector is selector for Kubernetes Scheduler.
	KubeSchedulerSelector = labels.SelectorFromSet(labels.Set{
		"app":                     "kubernetes",
		"garden.sapcloud.io/role": "controlplane",
		"role":                    "scheduler",
	})

	// KubeStateMetricsShootSelector is selector for KubeStateMetrics for Shoot.
	KubeStateMetricsShootSelector = labels.SelectorFromSet(labels.Set{
		"component":               "kube-state-metrics",
		"garden.sapcloud.io/role": "monitoring",
		"type":                    "shoot",
	})

	// KubeStateMetricsSeedSelector is selector for KubeStateMetrics for Seed.
	KubeStateMetricsSeedSelector = labels.SelectorFromSet(labels.Set{
		"component":               "kube-state-metrics",
		"garden.sapcloud.io/role": "monitoring",
		"type":                    "seed",
	})

	// MachineControllerManagerSelector is selector for MachineControllerManager.
	MachineControllerManagerSelector = labels.SelectorFromSet(labels.Set{
		"app":                     "kubernetes",
		"garden.sapcloud.io/role": "controlplane",
		"role":                    "machine-controller-manager",
	})

	// PrometheusSelector is selector for Prometheus.
	PrometheusSelector = labels.SelectorFromSet(labels.Set{
		"app":                     "prometheus",
		"garden.sapcloud.io/role": "monitoring",
		"role":                    "monitoring",
	})
)
