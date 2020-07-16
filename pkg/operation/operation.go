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

package operation

import (
	"context"
	"crypto/x509"
	"fmt"
	"strings"

	gardencorev1alpha1 "github.com/gardener/gardener/pkg/apis/core/v1alpha1"
	gardencorev1alpha1helper "github.com/gardener/gardener/pkg/apis/core/v1alpha1/helper"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	gardencorev1beta1helper "github.com/gardener/gardener/pkg/apis/core/v1beta1/helper"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	gardencoreinformers "github.com/gardener/gardener/pkg/client/core/informers/externalversions/core/v1beta1"
	"github.com/gardener/gardener/pkg/client/kubernetes/clientmap"
	"github.com/gardener/gardener/pkg/client/kubernetes/clientmap/keys"
	"github.com/gardener/gardener/pkg/gardenlet/apis/config"
	"github.com/gardener/gardener/pkg/operation/common"
	"github.com/gardener/gardener/pkg/operation/etcdencryption"
	"github.com/gardener/gardener/pkg/operation/garden"
	"github.com/gardener/gardener/pkg/operation/seed"
	"github.com/gardener/gardener/pkg/operation/shoot"
	"github.com/gardener/gardener/pkg/utils"
	"github.com/gardener/gardener/pkg/utils/chart"
	"github.com/gardener/gardener/pkg/utils/flow"
	"github.com/gardener/gardener/pkg/utils/imagevector"
	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"
	"github.com/gardener/gardener/pkg/utils/secrets"

	prometheusapi "github.com/prometheus/client_golang/api"
	prometheusclient "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// NewBuilder returns a new Builder.
func NewBuilder() *Builder {
	return &Builder{
		configFunc: func() (*config.GardenletConfiguration, error) {
			return nil, fmt.Errorf("config is required but not set")
		},
		gardenFunc: func(map[string]*corev1.Secret) (*garden.Garden, error) {
			return nil, fmt.Errorf("garden object is required but not set")
		},
		gardenerInfoFunc: func() (*gardencorev1beta1.Gardener, error) {
			return nil, fmt.Errorf("gardener info is required but not set")
		},
		imageVectorFunc: func() (imagevector.ImageVector, error) {
			return nil, fmt.Errorf("image vector is required but not set")
		},
		loggerFunc: func() (*logrus.Entry, error) {
			return nil, fmt.Errorf("logger is required but not set")
		},
		secretsFunc: func() (map[string]*corev1.Secret, error) {
			return nil, fmt.Errorf("secrets map is required but not set")
		},
		seedFunc: func(context.Context, client.Client) (*seed.Seed, error) {
			return nil, fmt.Errorf("seed object is required but not set")
		},
		shootFunc: func(context.Context, client.Client, *garden.Garden, *seed.Seed) (*shoot.Shoot, error) {
			return nil, fmt.Errorf("shoot object is required but not set")
		},
		chartsRootPathFunc: func() string {
			return common.ChartPath
		},
	}
}

// WithConfig sets the configFunc attribute at the Builder.
func (b *Builder) WithConfig(cfg *config.GardenletConfiguration) *Builder {
	b.configFunc = func() (*config.GardenletConfiguration, error) { return cfg, nil }
	return b
}

// WithGarden sets the gardenFunc attribute at the Builder.
func (b *Builder) WithGarden(g *garden.Garden) *Builder {
	b.gardenFunc = func(_ map[string]*corev1.Secret) (*garden.Garden, error) { return g, nil }
	return b
}

// WithGardenFrom sets the gardenFunc attribute at the Builder which will build a new Garden object.
func (b *Builder) WithGardenFrom(k8sGardenCoreInformers gardencoreinformers.Interface, namespace string) *Builder {
	b.gardenFunc = func(secrets map[string]*corev1.Secret) (*garden.Garden, error) {
		return garden.
			NewBuilder().
			WithProjectFromLister(k8sGardenCoreInformers.Projects().Lister(), namespace).
			WithInternalDomainFromSecrets(secrets).
			WithDefaultDomainsFromSecrets(secrets).
			Build()
	}
	return b
}

// WithGardenerInfo sets the gardenerInfoFunc attribute at the Builder.
func (b *Builder) WithGardenerInfo(gardenerInfo *gardencorev1beta1.Gardener) *Builder {
	b.gardenerInfoFunc = func() (*gardencorev1beta1.Gardener, error) { return gardenerInfo, nil }
	return b
}

// WithImageVector sets the imageVectorFunc attribute at the Builder.
func (b *Builder) WithImageVector(imageVector imagevector.ImageVector) *Builder {
	b.imageVectorFunc = func() (imagevector.ImageVector, error) { return imageVector, nil }
	return b
}

// WithLogger sets the loggerFunc attribute at the Builder.
func (b *Builder) WithLogger(logger *logrus.Entry) *Builder {
	b.loggerFunc = func() (*logrus.Entry, error) { return logger, nil }
	return b
}

// WithSecrets sets the secretsFunc attribute at the Builder.
func (b *Builder) WithSecrets(secrets map[string]*corev1.Secret) *Builder {
	b.secretsFunc = func() (map[string]*corev1.Secret, error) { return secrets, nil }
	return b
}

// WithSeed sets the seedFunc attribute at the Builder.
func (b *Builder) WithSeed(s *seed.Seed) *Builder {
	b.seedFunc = func(_ context.Context, _ client.Client) (*seed.Seed, error) { return s, nil }
	return b
}

// WithSeedFrom sets the seedFunc attribute at the Builder which will build a new Seed object.
func (b *Builder) WithSeedFrom(k8sGardenCoreInformers gardencoreinformers.Interface, seedName string) *Builder {
	b.seedFunc = func(ctx context.Context, c client.Client) (*seed.Seed, error) {
		return seed.
			NewBuilder().
			WithSeedObjectFromLister(k8sGardenCoreInformers.Seeds().Lister(), seedName).
			WithSeedSecretFromClient(ctx, c).
			Build()
	}
	return b
}

// WithShoot sets the shootFunc attribute at the Builder.
func (b *Builder) WithShoot(s *shoot.Shoot) *Builder {
	b.shootFunc = func(_ context.Context, _ client.Client, _ *garden.Garden, _ *seed.Seed) (*shoot.Shoot, error) {
		return s, nil
	}
	return b
}

// WithChartsRootPath sets the ChartsRootPath attribute at the Builder.
// Mainly used for testing. Optional.
func (b *Builder) WithChartsRootPath(chartsRootPath string) *Builder {
	b.chartsRootPathFunc = func() string { return chartsRootPath }
	return b
}

// WithShootFrom sets the shootFunc attribute at the Builder which will build a new Shoot object.
func (b *Builder) WithShootFrom(k8sGardenCoreInformers gardencoreinformers.Interface, s *gardencorev1beta1.Shoot) *Builder {
	b.shootFunc = func(ctx context.Context, c client.Client, gardenObj *garden.Garden, seedObj *seed.Seed) (*shoot.Shoot, error) {
		return shoot.
			NewBuilder().
			WithShootObject(s).
			WithCloudProfileObjectFromLister(k8sGardenCoreInformers.CloudProfiles().Lister()).
			WithShootSecretFromSecretBindingLister(k8sGardenCoreInformers.SecretBindings().Lister()).
			WithProjectName(gardenObj.Project.Name).
			WithDisableDNS(!seedObj.Info.Spec.Settings.ShootDNS.Enabled).
			WithInternalDomain(gardenObj.InternalDomain).
			WithDefaultDomains(gardenObj.DefaultDomains).
			Build(ctx, c)
	}
	return b
}

// Build initializes a new Operation object.
func (b *Builder) Build(ctx context.Context, clientMap clientmap.ClientMap) (*Operation, error) {
	operation := &Operation{
		ClientMap: clientMap,
		CheckSums: make(map[string]string),
	}

	gardenClient, err := clientMap.GetClient(ctx, keys.ForGarden())
	if err != nil {
		return nil, fmt.Errorf("failed to get garden client: %w", err)
	}
	operation.K8sGardenClient = gardenClient

	config, err := b.configFunc()
	if err != nil {
		return nil, err
	}
	operation.Config = config

	secretsMap, err := b.secretsFunc()
	if err != nil {
		return nil, err
	}
	secrets := make(map[string]*corev1.Secret)
	for k, v := range secretsMap {
		secrets[k] = v
	}
	operation.Secrets = secrets

	garden, err := b.gardenFunc(secrets)
	if err != nil {
		return nil, err
	}
	operation.Garden = garden

	gardenerInfo, err := b.gardenerInfoFunc()
	if err != nil {
		return nil, err
	}
	operation.GardenerInfo = gardenerInfo

	imageVector, err := b.imageVectorFunc()
	if err != nil {
		return nil, err
	}
	operation.ImageVector = imageVector

	logger, err := b.loggerFunc()
	if err != nil {
		return nil, err
	}
	operation.Logger = logger

	seed, err := b.seedFunc(ctx, gardenClient.Client())
	if err != nil {
		return nil, err
	}
	operation.Seed = seed

	shoot, err := b.shootFunc(ctx, gardenClient.Client(), garden, seed)
	if err != nil {
		return nil, err
	}
	operation.Shoot = shoot

	shootedSeed, err := gardencorev1beta1helper.ReadShootedSeed(shoot.Info)
	if err != nil {
		logger.Warnf("Cannot use shoot %s/%s as shooted seed: %+v", shoot.Info.Namespace, shoot.Info.Name, err)
	} else {
		operation.ShootedSeed = shootedSeed
	}

	operation.ChartsRootPath = b.chartsRootPathFunc()

	return operation, nil
}

// InitializeSeedClients will use the Garden Kubernetes client to read the Seed Secret in the Garden
// cluster which contains a Kubeconfig that can be used to authenticate against the Seed cluster. With it,
// a Kubernetes client as well as a Chart renderer for the Seed cluster will be initialized and attached to
// the already existing Operation object.
func (o *Operation) InitializeSeedClients() error {
	if o.K8sSeedClient != nil {
		return nil
	}

	seedClient, err := o.ClientMap.GetClient(context.TODO(), keys.ForSeed(o.Seed.Info))
	if err != nil {
		return fmt.Errorf("failed to get seed client: %w", err)
	}
	o.K8sSeedClient = seedClient
	return nil
}

// InitializeShootClients will use the Seed Kubernetes client to read the gardener Secret in the Seed
// cluster which contains a Kubeconfig that can be used to authenticate against the Shoot cluster. With it,
// a Kubernetes client as well as a Chart renderer for the Shoot cluster will be initialized and attached to
// the already existing Operation object.
func (o *Operation) InitializeShootClients() error {
	if o.K8sShootClient != nil {
		return nil
	}

	if o.Shoot.HibernationEnabled {
		// Don't initialize clients for Shoots, that are currently hibernated and their API server is not running
		apiServerRunning, err := o.IsAPIServerRunning()
		if err != nil {
			return err
		}
		if !apiServerRunning {
			return nil
		}
	}

	shootClient, err := o.ClientMap.GetClient(context.TODO(), keys.ForShoot(o.Shoot.Info))
	if err != nil {
		return err
	}
	o.K8sShootClient = shootClient

	return nil
}

// IsAPIServerRunning checks if the API server of the Shoot currently running (not scaled-down/deleted).
func (o *Operation) IsAPIServerRunning() (bool, error) {
	deployment := &appsv1.Deployment{}
	// use direct client here to make sure, we're not reading from a stale cache, when checking if we should initialize a shoot client (e.g. from within the care controller)
	if err := o.K8sSeedClient.DirectClient().Get(context.TODO(), kutil.Key(o.Shoot.SeedNamespace, v1beta1constants.DeploymentNameKubeAPIServer), deployment); err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	if deployment.GetDeletionTimestamp() != nil {
		return false, nil
	}

	if deployment.Spec.Replicas == nil {
		return false, nil
	}
	return *deployment.Spec.Replicas > 0, nil
}

// InitializeMonitoringClient will read the Prometheus ingress auth and tls
// secrets from the Seed cluster, which are containing the cert to secure
// the connection and the credentials authenticate against the Shoot Prometheus.
// With those certs and credentials, a Prometheus client API will be created
// and attached to the existing Operation object.
func (o *Operation) InitializeMonitoringClient() error {
	if o.MonitoringClient != nil {
		return nil
	}

	// Read the CA.
	tlsSecret := &corev1.Secret{}
	if err := o.K8sSeedClient.Client().Get(context.TODO(), kutil.Key(o.Shoot.SeedNamespace, common.PrometheusTLS), tlsSecret); err != nil {
		return err
	}

	ca := x509.NewCertPool()
	ca.AppendCertsFromPEM(tlsSecret.Data[secrets.DataKeyCertificateCA])

	// Read the basic auth credentials.
	credentials := &corev1.Secret{}
	if err := o.K8sSeedClient.Client().Get(context.TODO(), kutil.Key(o.Shoot.SeedNamespace, "monitoring-ingress-credentials"), credentials); err != nil {
		return err
	}

	config := prometheusapi.Config{
		Address: fmt.Sprintf("https://%s", o.ComputeIngressHost("p")),
		RoundTripper: &prometheusRoundTripper{
			authHeader: fmt.Sprintf("Basic %s", utils.EncodeBase64([]byte(fmt.Sprintf("%s:%s", credentials.Data[secrets.DataKeyUserName], credentials.Data[secrets.DataKeyPassword])))),
			ca:         ca,
		},
	}
	client, err := prometheusapi.NewClient(config)
	if err != nil {
		return err
	}
	o.MonitoringClient = prometheusclient.NewAPI(client)
	return nil
}

// GetSecretKeysOfRole returns a list of keys which are present in the Garden Secrets map and which
// are prefixed with <kind>.
func (o *Operation) GetSecretKeysOfRole(kind string) []string {
	return common.GetSecretKeysWithPrefix(kind, o.Secrets)
}

func makeDescription(stats *flow.Stats) string {
	if stats.ProgressPercent() == 100 {
		return "Execution finished"
	}
	return strings.Join(stats.Running.StringList(), ", ")
}

// ReportShootProgress will update the last operation object in the Shoot manifest `status` section
// by the current progress of the Flow execution.
func (o *Operation) ReportShootProgress(ctx context.Context, stats *flow.Stats) {
	var (
		description    = makeDescription(stats)
		progress       = stats.ProgressPercent()
		lastUpdateTime = metav1.Now()
	)

	newShoot, err := kutil.TryUpdateShootStatus(o.K8sGardenClient.GardenCore(), retry.DefaultRetry, o.Shoot.Info.ObjectMeta,
		func(shoot *gardencorev1beta1.Shoot) (*gardencorev1beta1.Shoot, error) {
			if shoot.Status.LastOperation == nil {
				return nil, fmt.Errorf("last operation of Shoot %s/%s is unset", shoot.Namespace, shoot.Name)
			}
			if shoot.Status.LastOperation.LastUpdateTime.After(lastUpdateTime.Time) {
				return nil, fmt.Errorf("last operation of Shoot %s/%s was updated mid-air", shoot.Namespace, shoot.Name)
			}
			shoot.Status.LastOperation.Description = description
			shoot.Status.LastOperation.Progress = progress
			shoot.Status.LastOperation.LastUpdateTime = lastUpdateTime
			return shoot, nil
		})
	if err != nil {
		o.Logger.Errorf("Could not report shoot progress: %v", err)
		return
	}

	o.Shoot.Info = newShoot
}

// CleanShootTaskError removes the error with taskID from the Shoot's status.LastErrors array.
// If the status.LastErrors array is empty then status.LastError is also removed.
func (o *Operation) CleanShootTaskError(_ context.Context, taskID string) {
	newShoot, err := kutil.TryUpdateShootStatus(o.K8sGardenClient.GardenCore(), retry.DefaultRetry, o.Shoot.Info.ObjectMeta,
		func(shoot *gardencorev1beta1.Shoot) (*gardencorev1beta1.Shoot, error) {
			shoot.Status.LastErrors = gardencorev1beta1helper.DeleteLastErrorByTaskID(o.Shoot.Info.Status.LastErrors, taskID)
			return shoot, nil
		},
	)
	if err != nil {
		o.Logger.Errorf("Could not report shoot progress: %v", err)
		return
	}
	o.Shoot.Info = newShoot
}

// SeedVersion is a shorthand for the kubernetes version of the K8sSeedClient.
func (o *Operation) SeedVersion() string {
	return o.K8sSeedClient.Version()
}

// ShootVersion is a shorthand for the desired kubernetes version of the operation's shoot.
func (o *Operation) ShootVersion() string {
	return o.Shoot.Info.Spec.Kubernetes.Version
}

// InjectSeedSeedImages injects images that shall run on the Seed and target the Seed's Kubernetes version.
func (o *Operation) InjectSeedSeedImages(values map[string]interface{}, names ...string) (map[string]interface{}, error) {
	return chart.InjectImages(values, o.ImageVector, names, imagevector.RuntimeVersion(o.SeedVersion()), imagevector.TargetVersion(o.SeedVersion()))
}

// InjectSeedShootImages injects images that shall run on the Seed but target the Shoot's Kubernetes version.
func (o *Operation) InjectSeedShootImages(values map[string]interface{}, names ...string) (map[string]interface{}, error) {
	return chart.InjectImages(values, o.ImageVector, names, imagevector.RuntimeVersion(o.SeedVersion()), imagevector.TargetVersion(o.ShootVersion()))
}

// InjectShootShootImages injects images that shall run on the Shoot and target the Shoot's Kubernetes version.
func (o *Operation) InjectShootShootImages(values map[string]interface{}, names ...string) (map[string]interface{}, error) {
	return chart.InjectImages(values, o.ImageVector, names, imagevector.RuntimeVersion(o.ShootVersion()), imagevector.TargetVersion(o.ShootVersion()))
}

// EnsureShootStateExists creates the ShootState resource for the corresponding shoot and sets its ownerReferences to the Shoot.
func (o *Operation) EnsureShootStateExists(ctx context.Context) error {
	shootState := &gardencorev1alpha1.ShootState{
		ObjectMeta: metav1.ObjectMeta{
			Name:      o.Shoot.Info.Name,
			Namespace: o.Shoot.Info.Namespace,
		},
	}
	ownerReference := metav1.NewControllerRef(o.Shoot.Info, gardencorev1beta1.SchemeGroupVersion.WithKind("Shoot"))
	blockOwnerDeletion := false
	ownerReference.BlockOwnerDeletion = &blockOwnerDeletion

	_, err := controllerutil.CreateOrUpdate(ctx, o.K8sGardenClient.Client(), shootState, func() error {
		shootState.OwnerReferences = []metav1.OwnerReference{*ownerReference}
		return nil
	})
	if err != nil {
		return err
	}

	o.ShootState = shootState
	gardenerResourceList := gardencorev1alpha1helper.GardenerResourceDataList(shootState.Spec.Gardener)
	o.Shoot.ETCDEncryption, err = etcdencryption.GetEncryptionConfig(gardenerResourceList)
	return err
}

// DeleteClusterResourceFromSeed deletes the `Cluster` extension resource for the shoot in the seed cluster.
func (o *Operation) DeleteClusterResourceFromSeed(ctx context.Context) error {
	if err := o.InitializeSeedClients(); err != nil {
		o.Logger.Errorf("Could not initialize a new Kubernetes client for the seed cluster: %s", err.Error())
		return err
	}

	return client.IgnoreNotFound(o.K8sSeedClient.Client().Delete(ctx, &extensionsv1alpha1.Cluster{ObjectMeta: metav1.ObjectMeta{Name: o.Shoot.SeedNamespace}}))
}

// SwitchBackupEntryToTargetSeed changes the BackupEntry in the Garden cluster to the Target Seed and removes it from the Source Seed
func (o *Operation) SwitchBackupEntryToTargetSeed(ctx context.Context) error {
	var (
		name              = common.GenerateBackupEntryName(o.Shoot.Info.Status.TechnicalID, o.Shoot.Info.Status.UID)
		gardenBackupEntry = &gardencorev1beta1.BackupEntry{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: o.Shoot.Info.Namespace,
			},
		}
	)

	return kutil.TryUpdate(ctx, retry.DefaultBackoff, o.K8sGardenClient.DirectClient(), gardenBackupEntry, func() error {
		gardenBackupEntry.Spec.SeedName = o.Shoot.Info.Spec.SeedName
		return nil
	})
}

// ComputeGrafanaHosts computes the host for both grafanas.
func (o *Operation) ComputeGrafanaHosts() []string {
	return []string{
		o.ComputeGrafanaOperatorsHostDeprecated(),
		o.ComputeGrafanaUsersHostDeprecated(),
		o.ComputeGrafanaOperatorsHost(),
		o.ComputeGrafanaUsersHost(),
	}
}

// ComputePrometheusHosts computes the hosts for prometheus.
func (o *Operation) ComputePrometheusHosts() []string {
	return []string{
		o.ComputePrometheusHostDeprecated(),
		o.ComputePrometheusHost(),
	}
}

// ComputeAlertManagerHosts computes the host for alert manager.
func (o *Operation) ComputeAlertManagerHosts() []string {
	return []string{
		o.ComputeAlertManagerHostDeprecated(),
		o.ComputeAlertManagerHost(),
	}
}

// ComputeGrafanaOperatorsHostDeprecated computes the host for users Grafana.
// TODO: timuthy - remove in the future. Old Grafana host is retained for migration reasons.
func (o *Operation) ComputeGrafanaOperatorsHostDeprecated() string {
	return o.ComputeIngressHostDeprecated(common.GrafanaOperatorsPrefix)
}

// ComputeGrafanaUsersHostDeprecated computes the host for operators Grafana.
// TODO: timuthy - remove in the future. Old Grafana host is retained for migration reasons.
func (o *Operation) ComputeGrafanaUsersHostDeprecated() string {
	return o.ComputeIngressHostDeprecated(common.GrafanaUsersPrefix)
}

// ComputeGrafanaOperatorsHost computes the host for users Grafana.
func (o *Operation) ComputeGrafanaOperatorsHost() string {
	return o.ComputeIngressHost(common.GrafanaOperatorsPrefix)
}

// ComputeGrafanaUsersHost computes the host for operators Grafana.
func (o *Operation) ComputeGrafanaUsersHost() string {
	return o.ComputeIngressHost(common.GrafanaUsersPrefix)
}

// ComputeAlertManagerHostDeprecated computes the host for alert manager.
// TODO: timuthy - remove in the future. Old AlertManager host is retained for migration reasons.
func (o *Operation) ComputeAlertManagerHostDeprecated() string {
	return o.ComputeIngressHostDeprecated(common.AlertManagerPrefix)
}

// ComputeAlertManagerHost computes the host for alert manager.
func (o *Operation) ComputeAlertManagerHost() string {
	return o.ComputeIngressHost(common.AlertManagerPrefix)
}

// ComputePrometheusHostDeprecated computes the host for prometheus.
// TODO: timuthy - remove in the future. Old Prometheus host is retained for migration reasons.
func (o *Operation) ComputePrometheusHostDeprecated() string {
	return o.ComputeIngressHostDeprecated(common.PrometheusPrefix)
}

// ComputePrometheusHost computes the host for prometheus.
func (o *Operation) ComputePrometheusHost() string {
	return o.ComputeIngressHost(common.PrometheusPrefix)
}

// ComputeIngressHostDeprecated computes the host for a given prefix.
// TODO: timuthy - remove in the future. Only retained for migration reasons.
func (o *Operation) ComputeIngressHostDeprecated(prefix string) string {
	return o.Seed.GetIngressFQDNDeprecated(prefix, o.Shoot.Info.Name, o.Garden.Project.Name)
}

// ComputeIngressHost computes the host for a given prefix.
func (o *Operation) ComputeIngressHost(prefix string) string {
	shortID := strings.Replace(o.Shoot.Info.Status.TechnicalID, shoot.TechnicalIDPrefix, "", 1)
	return fmt.Sprintf("%s-%s.%s", prefix, shortID, o.Seed.Info.Spec.DNS.IngressDomain)
}
