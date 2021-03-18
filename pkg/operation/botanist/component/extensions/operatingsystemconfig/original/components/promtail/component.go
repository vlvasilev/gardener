// Copyright (c) 2021 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

package promtail

import (
	"github.com/gardener/gardener/charts"
	gardencorev1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/operation/botanist/component/extensions/operatingsystemconfig/original/components"
	"github.com/gardener/gardener/pkg/operation/botanist/component/extensions/operatingsystemconfig/original/components/docker"
	"github.com/gardener/gardener/pkg/utils/imagevector"
	"k8s.io/utils/pointer"
)

const (
	// UnitName is the name of the promtail service.
	UnitName = gardencorev1beta1constants.OperatingSystemConfigUnitNamePromtailService
	// PathPromtailDirectory is the path for the promtail's directory.
	PathPromtailDirectory = "/var/lib/promtail"
	// PathPromtailBinary is the path for the promtail binary.
	PathPromtailBinary = "/opt/bin"
	// PathPromtailAuthToken is the path for the promtail authentication token,
	// which is use to auth agains the Loki sidecar proxy.
	PathPromtailAuthToken = PathPromtailDirectory + "/auth-token"
	// PathPromtailConfig is the path for the promtail's configuration file
	PathPromtailConfig = gardencorev1beta1constants.OperatingSystemConfigFilePathPromtailConfig
	// PathPromtailCACert is the path for the loki-tls certificate authority.
	PathPromtailCACert = PathPromtailDirectory + "/ca.crt"
	// PromtailServerPort is the promtail listening port
	PromtailServerPort = 3001
	// PromtailPositionFile is the path for storing the scraped file offsets
	PromtailPositionFile = "/run/promtail/positions.yaml"
)

type component struct{}

// New returns a new promtail component.
func New() *component {
	return &component{}
}

func (component) Name() string {
	return "promtail"
}

func execStartPreCopyBinaryFromContainer(binaryName string, image *imagevector.Image) string {
	return docker.PathBinary + ` run --rm -v /opt/bin:/opt/bin:rw --entrypoint /bin/sh ` + image.String() + ` -c "cp /usr/bin/` + binaryName + ` /opt/bin"`
}

func (component) Config(ctx components.Context) ([]extensionsv1alpha1.Unit, []extensionsv1alpha1.File, error) {
	promtailAuthTokenFile := getPromtailAuthTokenFile(ctx)
	if promtailAuthTokenFile == nil {
		return nil, nil, nil
	}

	promtailConfigFiled, err := getPromtailConfigurationFile(ctx)
	if err != nil {
		return nil, nil, err
	}

	promtailCAFile := getPromtailCAFile(ctx)

	return []extensionsv1alpha1.Unit{
			{
				Name:    UnitName,
				Command: pointer.StringPtr("start"),
				Enable:  pointer.BoolPtr(true),
				Content: pointer.StringPtr(`[Unit]
Description=promtail daemon
Documentation=https://grafana.com/docs/loki/latest/clients/promtail/
[Install]
WantedBy=multi-user.target
[Service]
Restart=always
RestartSec=5
EnvironmentFile=/etc/environment
ExecStartPre=` + execStartPreCopyBinaryFromContainer("promtail", ctx.Images[charts.PromtailImageName]) + `
ExecStart=` + PathPromtailBinary + `/promtail -config.file=` + PathPromtailConfig),
			}},
		[]extensionsv1alpha1.File{
			*promtailConfigFiled,
			*promtailAuthTokenFile,
			*promtailCAFile,
		}, nil
}
