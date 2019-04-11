// Copyright (c) 2018 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

package gcpbotanist

import (
	"fmt"
	"net"
	"path"

	"github.com/gardener/gardener/pkg/operation/common"
	"github.com/gardener/gardener/pkg/operation/terraformer"
)

const cloudProviderConfigTemplate = `
[Global]
project-id=%q
network-name=%q
%v
multizone=true
local-zone=%q
token-url=nil
node-tags=%q
`

// GenerateCloudProviderConfig generates the GCE cloud provider config.
// See this for more details:
// https://github.com/kubernetes/kubernetes/blob/master/pkg/cloudprovider/providers/gce/gce.go
func (b *GCPBotanist) GenerateCloudProviderConfig() (string, error) {
	networkName := b.VPCName
	if networkName == "" {
		networkName = b.Shoot.SeedNamespace
	}

	var (
		subnetID       = "subnet_internal"
		subNetworkName = ""
	)
	tf, err := b.NewShootTerraformer(common.TerraformerPurposeInfra)
	if err != nil {
		return "", err
	}
	stateVariables, err := tf.GetStateOutputVariables(subnetID)
	if err != nil {
		if !terraformer.IsVariablesNotFoundError(err) {
			return "", err
		}
		b.Logger.Debugf("Skipping explicit GCP subnet creation for internal loadbalancers because subnet_internal variable has not been found in the Terraform state.")
	} else {
		subNetworkName = fmt.Sprintf("subnetwork-name=%q", stateVariables[subnetID])
	}
	return fmt.Sprintf(
		cloudProviderConfigTemplate,
		b.Project,
		networkName,
		subNetworkName,
		b.Shoot.Info.Spec.Cloud.GCP.Zones[0],
		b.Shoot.SeedNamespace,
	), nil
}

// RefreshCloudProviderConfig refreshes the cloud provider credentials in the existing cloud
// provider config.
// Not needed on GCP (cloud provider config does not contain the credentials), hence, the
// original is returned back.
func (b *GCPBotanist) RefreshCloudProviderConfig(currentConfig map[string]string) map[string]string {
	return currentConfig
}

// GenerateKubeAPIServerServiceConfig generates the cloud provider specific values which are required to render the
// Service manifest of the kube-apiserver-service properly.
func (b *GCPBotanist) GenerateKubeAPIServerServiceConfig() (map[string]interface{}, error) {
	return nil, nil
}

// GenerateKubeAPIServerExposeConfig defines the cloud provider specific values which configure how the kube-apiserver
// is exposed to the public.
func (b *GCPBotanist) GenerateKubeAPIServerExposeConfig() (map[string]interface{}, error) {
	return map[string]interface{}{
		"advertiseAddress": b.APIServerAddress,
	}, nil
}

// GenerateKubeAPIServerConfig generates the cloud provider specific values which are required to render the
// Deployment manifest of the kube-apiserver properly.
func (b *GCPBotanist) GenerateKubeAPIServerConfig() (map[string]interface{}, error) {
	return map[string]interface{}{
		"environment": getGCPCredentialsEnvironment(),
	}, nil
}

// GenerateCloudControllerManagerConfig generates the cloud provider specific values which are required to
// render the Deployment manifest of the cloud-controller-manager properly.
func (b *GCPBotanist) GenerateCloudControllerManagerConfig() (map[string]interface{}, string, error) {
	return map[string]interface{}{
		"environment": getGCPCredentialsEnvironment(),
	}, common.CloudControllerManagerDeploymentName, nil
}

// GenerateCSIConfig generates the configuration for CSI charts
func (b *GCPBotanist) GenerateCSIConfig() (map[string]interface{}, error) {
	return nil, nil
}

// MetadataServiceAddress returns GCP's MetadataService address
func (b *GCPBotanist) MetadataServiceAddress() (*net.IPNet, error) {
	_, cidr, err := net.ParseCIDR("169.254.169.254/32")
	return cidr, err
}

// GenerateKubeControllerManagerConfig generates the cloud provider specific values which are required to
// render the Deployment manifest of the kube-controller-manager properly.
func (b *GCPBotanist) GenerateKubeControllerManagerConfig() (map[string]interface{}, error) {
	return map[string]interface{}{
		"environment": getGCPCredentialsEnvironment(),
	}, nil
}

// maps are mutable, so it's safer to create a new instance
func getGCPCredentialsEnvironment() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":  "GOOGLE_APPLICATION_CREDENTIALS",
			"value": fmt.Sprintf("/srv/cloudprovider/%s", ServiceAccountJSON),
		},
	}
}

// GenerateKubeSchedulerConfig generates the cloud provider specific values which are required to render the
// Deployment manifest of the kube-scheduler properly.
func (b *GCPBotanist) GenerateKubeSchedulerConfig() (map[string]interface{}, error) {
	return nil, nil
}

// GenerateEtcdBackupConfig returns the etcd backup configuration for the etcd Helm chart.
func (b *GCPBotanist) GenerateEtcdBackupConfig() (map[string][]byte, map[string]interface{}, error) {
	var (
		mountPath  = "/root/.gcp/"
		bucketName = "bucketName"
	)
	tf, err := b.NewBackupInfrastructureTerraformer()
	if err != nil {
		return nil, nil, err
	}
	stateVariables, err := tf.GetStateOutputVariables(bucketName)
	if err != nil {
		return nil, nil, err
	}

	secretData := map[string][]byte{
		ServiceAccountJSON: []byte(b.MinifiedServiceAccount),
	}

	backupConfigData := map[string]interface{}{
		"schedule":         b.Operation.ShootBackup.Schedule,
		"storageProvider":  "GCS",
		"storageContainer": stateVariables[bucketName],
		"env": []map[string]interface{}{
			{
				"name":  "GOOGLE_APPLICATION_CREDENTIALS",
				"value": path.Join(mountPath, ServiceAccountJSON),
			},
		},
		"volumeMounts": []map[string]interface{}{
			{
				"mountPath": mountPath,
				"name":      common.BackupSecretName,
			},
		},
	}

	return secretData, backupConfigData, nil
}

// DeployCloudSpecificControlPlane does currently nothing for GCP.
func (b *GCPBotanist) DeployCloudSpecificControlPlane() error {
	return nil
}
