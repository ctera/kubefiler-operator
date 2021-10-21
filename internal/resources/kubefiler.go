/*
Copyright 2021, CTERA Networks

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package resources

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	ctera "github.com/ctera/ctera-gateway-openapi-go-client"
	kubefilerv1alpha1 "github.com/ctera/kubefiler-operator/api/v1alpha1"
	"github.com/ctera/kubefiler-operator/internal/conf"
	"github.com/ctera/kubefiler-operator/internal/cteraclient"
)

const (
	kubeFilerFinalizer = "kubefiler-operator.ctera.com/kubeFilerFinalizer"

	kubeFilerNfsMountdPort = 40892
	kubeFilerNfsStatdPort  = 40662
	kubeFilerNfsNfsdHost   = "127.0.0.1"
)

// KubeFilerManager is used to manage KubeFiler resources.
type KubeFilerManager struct {
	client   client.Client
	recorder record.EventRecorder
	logger   Logger
	cfg      *conf.OperatorConfig
}

// NewKubeFilerManager creates a KubeFilerManager.
func NewKubeFilerManager(client client.Client, recorder record.EventRecorder, logger Logger) *KubeFilerManager {
	return &KubeFilerManager{
		client:   client,
		recorder: recorder,
		logger:   logger,
		cfg:      conf.Get(),
	}
}

// Process is called by the controller on any type of reconciliation.
func (m *KubeFilerManager) Process(ctx context.Context, nsname types.NamespacedName) Result {
	// fetch our resource to determine what to do next
	instance := &kubefilerv1alpha1.KubeFiler{}
	err := m.client.Get(ctx, nsname, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found. Not a fatal error.
			return Done
		}
		m.logger.Error(err, "get failed for KubeFiler",
			"ns", nsname.Namespace,
			"name", nsname.Name)
		return Result{err: err}
	}

	// now that we have the resource. determine if its alive or pending deletion
	if instance.GetDeletionTimestamp() != nil {
		// its being deleted
		if controllerutil.ContainsFinalizer(instance, kubeFilerFinalizer) {
			// and our finalizer is present
			return m.Finalize(ctx, instance)
		}
		return Done
	}
	// resource is alive
	return m.Update(ctx, instance)
}

func (m *KubeFilerManager) Update(ctx context.Context, instance *kubefilerv1alpha1.KubeFiler) Result {
	var result Result

	m.logger.Info("Updating state for KubeFiler",
		"name", instance.Name,
		"UID", instance.UID,
		"Phase", instance.Status.Phase,
	)

	switch instance.Status.Phase {
	case "":
		result = m.StartDeployment(ctx, instance)
	case kubefilerv1alpha1.KubeFilerDeploying:
		result = m.Deploy(ctx, instance)
	case kubefilerv1alpha1.KubeFilerDeployed:
		result = m.CanConfigure(ctx, instance)
	case kubefilerv1alpha1.KubeFilerConfiguring:
		result = m.Configure(ctx, instance)
	case kubefilerv1alpha1.KubeFilerRunning:
		result = m.IsRunning(ctx, instance)
	case kubefilerv1alpha1.KubeFilerError:
		result = m.CanConfigure(ctx, instance)
	}

	return result
}

func (m *KubeFilerManager) StartDeployment(ctx context.Context, instance *kubefilerv1alpha1.KubeFiler) Result {
	instance.Status.Phase = kubefilerv1alpha1.KubeFilerDeploying
	m.client.Status().Update(ctx, instance)
	return Done
}

// Update should be called when a KubeFiler resource changes.
func (m *KubeFilerManager) Deploy(ctx context.Context, instance *kubefilerv1alpha1.KubeFiler) Result {
	m.logger.Info("Deploying KubeFiler",
		"name", instance.Name,
		"UID", instance.UID)

	changed, err := m.addFinalizer(ctx, instance)
	if err != nil {
		return Result{err: err}
	} else if changed {
		m.logger.Info("Added finalizer")
		return Requeue
	}

	gatewaySecret, created, err := getOrCreateGatewaySecret(ctx, m.client, instance)
	if err != nil {
		return Result{err: err}
	} else if created {
		m.logger.Info("Created Gateway secret")
		return Requeue
	}

	if kubeFilerNeedsPvc(instance) {
		pvc, created, err := getOrCreateGatewayPvc(ctx, m.client, instance)
		if err != nil {
			return Result{err: err}
		} else if created {
			m.logger.Info("Created PVC")
			m.recorder.Eventf(instance,
				EventNormal,
				ReasonCreatedPersistentVolumeClaim,
				"Created PVC %s for KubeFiler", pvc.Name)
			return Requeue
		}
		// if name is unset in the YAML, set it here
		instance.Spec.Storage.Pvc.Name = pvc.Name
	}

	deployment, created, err := getOrCreateGatewayDeployment(ctx, m.client, m.cfg, instance, gatewaySecret)
	if err != nil {
		return Result{err: err}
	} else if created {
		// Deployment created successfully - return and requeue
		m.logger.Info("Created deployment")
		m.recorder.Eventf(instance,
			EventNormal,
			ReasonCreatedDeployment,
			"Created deployment %s for KubeFiler", deployment.Name)
		return Requeue
	}

	service, created, err := getOrCreateGatewayService(ctx, m.client, instance)
	if err != nil {
		return Result{err: err}
	} else if created {
		m.logger.Info("Created service")
		m.recorder.Eventf(instance,
			EventNormal,
			ReasonCreatedService,
			"Created service %s for KubeFiler", service.Name)
		return Requeue
	}

	instance.Status.Phase = kubefilerv1alpha1.KubeFilerDeployed
	m.client.Status().Update(ctx, instance)
	m.logger.Info("Done updating KubeFiler resources")
	return Done
}

func (m *KubeFilerManager) CanConfigure(ctx context.Context, instance *kubefilerv1alpha1.KubeFiler) Result {
	result, ready := m.isFilerContainerReady(ctx, instance)
	if result != Done {
		return result
	}

	if ready {
		instance.Status.Phase = kubefilerv1alpha1.KubeFilerConfiguring
		m.client.Status().Update(ctx, instance)
		m.recorder.Eventf(instance,
			EventNormal,
			ReasonReadyForConfiguration,
			"The KubeFiler is ready to be configured")
	}

	return Done
}

func (m *KubeFilerManager) Configure(ctx context.Context, instance *kubefilerv1alpha1.KubeFiler) Result {
	cteraClient, result := m.initializeFiler(ctx, instance)
	if result != Done {
		return result
	}

	result = m.setLinuxAvoidUsingFanotify(cteraClient)
	if result != Done {
		return result
	}

	result = m.connectToPortal(ctx, instance, cteraClient)
	if result != Done {
		return result
	}

	result = m.startCachingGateway(instance, cteraClient)
	if result != Done {
		return result
	}

	result = m.configureNfsServer(cteraClient)
	if result != Done {
		return result
	}

	result = m.configureShares(ctx, instance, cteraClient)
	if result != Done {
		return result
	}

	instance.Status.Phase = kubefilerv1alpha1.KubeFilerRunning
	m.client.Status().Update(ctx, instance)
	m.logger.Info("Done configuring the KubeFiler")
	m.recorder.Eventf(instance,
		EventNormal,
		ReasonConfiguredSuccessfully,
		"The Filer was configured successfully - moving to Running")

	return Done
}

func (m *KubeFilerManager) initializeFiler(ctx context.Context, instance *kubefilerv1alpha1.KubeFiler) (*cteraclient.CteraClient, Result) {
	kubeFilerSecret, err := getSecret(ctx, m.client, instance.GetNamespace(), getGatewaySecretName(instance))
	if err != nil {
		m.logger.Error(err, "Failed to get the KubeFiler secret")
		return nil, Result{err: err}
	}

	kubeFilerService, err := getService(ctx, m.client, instance.GetNamespace(), getGatewayServiceName(instance))
	if err != nil {
		m.logger.Error(err, "Failed to get the KubeFiler service")
		return nil, Result{err: err}
	}

	cteraClient, err := cteraclient.NewCteraClient(m.logger, kubeFilerService.Spec.ClusterIP)
	if err != nil {
		return nil, Result{err: err}
	}

	initialized, err := cteraClient.InitializeFiler(
		ctx,
		string(kubeFilerSecret.Data[GatewayUsernameKey]),
		string(kubeFilerSecret.Data[GatewayPasswordKey]),
	)
	if err != nil {
		m.logger.Error(err, "Failed to initialize the KubeFiler")
		return nil, Result{err: err}
	} else if initialized {
		m.logger.Info("First user was created")
		m.recorder.Eventf(instance,
			EventNormal,
			ReasonFirstUserCreated,
			"The Filer's first user was created")
		return nil, Requeue
	}

	err = cteraClient.Authenticate(ctx,
		string(kubeFilerSecret.Data[GatewayUsernameKey]),
		string(kubeFilerSecret.Data[GatewayPasswordKey]),
	)
	if err != nil {
		m.logger.Error(err, "Failed to login to the KubeFiler")
		return nil, Result{err: err}
	}

	return cteraClient, Done
}

func (m *KubeFilerManager) setLinuxAvoidUsingFanotify(cteraClient *cteraclient.CteraClient) Result {
	changed, err := cteraClient.SetLinuxAvoidUsingFanotify(true)
	if err != nil {
		m.logger.Error(err, "Failed to set LinuxAvoidUsingFanotify")
		return Result{err: err}
	} else if changed {
		m.logger.Info("LinuxAvoidUsingFanotify was set")
		return Requeue
	}

	return Done
}

func (m *KubeFilerManager) connectToPortal(ctx context.Context, instance *kubefilerv1alpha1.KubeFiler, cteraClient *cteraclient.CteraClient) Result {
	kubeFilerPortal, err := getKubeFilerPortal(ctx, m.client, instance.GetNamespace(), instance.Spec.Portal)
	if err != nil {
		return Result{err: err}
	}

	kubeFilerPortalSecret, err := getSecret(ctx, m.client, instance.GetNamespace(), kubeFilerPortal.Spec.Credentials.Secret)
	if err != nil {
		return Result{err: err}
	}

	changed, err := cteraClient.ConnectToPortal(
		kubeFilerPortal.Spec.Address,
		string(kubeFilerPortalSecret.Data[kubeFilerPortal.Spec.Credentials.UsernameKey]),
		string(kubeFilerPortalSecret.Data[kubeFilerPortal.Spec.Credentials.PasswordKey]),
		kubeFilerPortal.Spec.Trust,
	)
	if err != nil {
		return Result{err: err}
	} else if changed {
		m.logger.Info("Connected to Portal successfully")
		m.recorder.Eventf(instance,
			EventNormal,
			ReasonConnectedToPortal,
			"The Filer was connected to the portal @%s", kubeFilerPortal.Spec.Address)
		return Requeue
	}

	return Done
}

func (m *KubeFilerManager) startCachingGateway(instance *kubefilerv1alpha1.KubeFiler, cteraClient *cteraclient.CteraClient) Result {
	changed, err := cteraClient.EnableCache()
	if err != nil {
		return Result{err: err}
	} else if changed {
		m.logger.Info("Cache was enabled successfully")
		return Requeue
	}

	changed, err = cteraClient.UnsuspendSync()
	if err != nil {
		return Result{err: err}
	} else if changed {
		m.logger.Info("Sync was unsuspended successfully")
		return Requeue
	}

	m.recorder.Eventf(instance,
		EventNormal,
		ReasonCacheGatewayStarted,
		"Filer was configured as Cache Gateway and Sync was unsuspended successfully",
	)

	return Done
}

func (m *KubeFilerManager) configureNfsServer(cteraClient *cteraclient.CteraClient) Result {
	conf, err := cteraClient.GetNfsConfiguration()
	if err != nil {
		return Result{err: err}
	}

	changed := m.alignNfsConfiguration(conf)
	if !changed {
		return Done
	}

	err = cteraClient.SetNfsConfiguration(conf)
	if err != nil {
		return Result{err: err}
	}

	return Requeue
}

func (m *KubeFilerManager) alignNfsConfiguration(conf *ctera.NfsConfiguration) bool {
	changed := false

	var mountdPort int32 = kubeFilerNfsMountdPort
	if !conf.MountdPort.IsSet() || conf.MountdPort.Get() == nil || *conf.MountdPort.Get() != mountdPort {
		conf.MountdPort.Set(&mountdPort)
		changed = true
	}

	var statdPort int32 = kubeFilerNfsStatdPort
	if !conf.StatdPort.IsSet() || conf.StatdPort.Get() == nil || *conf.StatdPort.Get() != statdPort {
		conf.StatdPort.Set(&statdPort)
		changed = true
	}

	nfsdHost := kubeFilerNfsNfsdHost
	if !conf.NfsdHost.IsSet() || conf.NfsdHost.Get() == nil || *conf.NfsdHost.Get() != nfsdHost {
		conf.NfsdHost.Set(&nfsdHost)
		changed = true
	}

	return changed
}

func (m *KubeFilerManager) configureShares(ctx context.Context, instance *kubefilerv1alpha1.KubeFiler, cteraClient *cteraclient.CteraClient) Result {
	kubeFilerExports, err := m.getKubeFilerExportsForKubeFiler(ctx, instance)
	if err != nil {
		return Result{err: err}
	} else if len(kubeFilerExports) == 0 {
		m.recorder.Eventf(instance, EventNormal, ReasonExportsConfigured, "No exports are configured for the KubeFiler")
		return Done
	}

	kubeFilerExportsPVCs, err := m.getPVCsForKubeFilerExports(ctx, instance.GetNamespace(), kubeFilerExports)
	if err != nil {
		return Result{err: err}
	}

	kubeFilerVolumeAttachments, err := m.getKubeFilerVolumeAttachments(ctx)
	if err != nil {
		return Result{err: err}
	}

	for exportName, export := range kubeFilerExports {
		_, created, err := getOrCreateShare(cteraClient, &export, m.getTrustedNfsClientsForExport(kubeFilerExportsPVCs[exportName], kubeFilerVolumeAttachments))
		if err != nil {
			return Result{err: err}
		} else if created {
			return Requeue
		}
	}

	m.recorder.Eventf(instance, EventNormal, ReasonExportsConfigured, "All exports for the KubeFiler are configured")

	return Done
}

func (m *KubeFilerManager) getKubeFilerExportsForKubeFiler(ctx context.Context, instance *kubefilerv1alpha1.KubeFiler) (map[string]kubefilerv1alpha1.KubeFilerExport, error) {
	kubeFilerExportList := kubefilerv1alpha1.KubeFilerExportList{}
	err := m.client.List(
		ctx,
		&kubeFilerExportList,
		client.InNamespace(instance.GetNamespace()),
	)
	if err != nil {
		return nil, err
	}

	kubeFilerExportsMap := make(map[string]kubefilerv1alpha1.KubeFilerExport)
	for _, kubeFilerExport := range kubeFilerExportList.Items {
		if kubeFilerExport.Spec.KubeFiler == instance.GetName() {
			kubeFilerExportsMap[kubeFilerExport.GetName()] = kubeFilerExport
		}
	}

	return kubeFilerExportsMap, nil
}

func (m *KubeFilerManager) getPVCsForKubeFilerExports(ctx context.Context, namespace string, kubeFilerExports map[string]kubefilerv1alpha1.KubeFilerExport) (map[string][]corev1.PersistentVolumeClaim, error) {
	persistentVolumeClaimList := corev1.PersistentVolumeClaimList{}
	err := m.client.List(
		ctx,
		&persistentVolumeClaimList,
		client.InNamespace(namespace),
	)
	if err != nil {
		return nil, err
	}

	persistentVolumeClaimMap := make(map[string][]corev1.PersistentVolumeClaim)
	for _, persistentVolumeClaim := range persistentVolumeClaimList.Items {
		if kubeFilerExportName, isKubeFilerPvc := persistentVolumeClaim.ObjectMeta.Annotations["kubefiler.ctera.com/kubefilerexport"]; isKubeFilerPvc {
			if _, inExportsMap := kubeFilerExports[kubeFilerExportName]; inExportsMap {
				persistentVolumeClaimMap[kubeFilerExportName] = append(persistentVolumeClaimMap[kubeFilerExportName], persistentVolumeClaim)
			}
		}
	}

	return persistentVolumeClaimMap, nil
}

func (m *KubeFilerManager) getKubeFilerVolumeAttachments(ctx context.Context) (map[string]storagev1.VolumeAttachment, error) {
	volumeAttachmentList := storagev1.VolumeAttachmentList{}
	err := m.client.List(
		ctx,
		&volumeAttachmentList,
	)
	if err != nil {
		return nil, err
	}

	volumeAttachmentMap := make(map[string]storagev1.VolumeAttachment)
	for _, volumeAttachment := range volumeAttachmentList.Items {
		if volumeAttachment.Spec.Attacher == "csi.kubefiler.ctera.com" {
			volumeAttachmentMap[*volumeAttachment.Spec.Source.PersistentVolumeName] = volumeAttachment
		}
	}

	return volumeAttachmentMap, nil
}

func (m *KubeFilerManager) getTrustedNfsClientsForExport(pvcs []corev1.PersistentVolumeClaim, volumeAttachments map[string]storagev1.VolumeAttachment) []ctera.NFSv3AccessControlEntry {
	node_ip_list := make(map[string]bool)
	for _, pvc := range pvcs {
		volumeName := pvc.Spec.VolumeName
		if len(volumeName) != 0 {
			if volumeAttachment, found := volumeAttachments[volumeName]; found {
				node_ip_list[volumeAttachment.ObjectMeta.Annotations["csi.alpha.kubernetes.io/node-id"]] = true
			}
		}
	}

	trustedNfsClients := make([]ctera.NFSv3AccessControlEntry, 0)
	for node_ip := range node_ip_list {
		trustedNfsClients = append(trustedNfsClients, *ctera.NewNFSv3AccessControlEntry(node_ip, "255.255.255.255", ctera.RW))
	}

	return trustedNfsClients
}

func (m *KubeFilerManager) IsRunning(ctx context.Context, instance *kubefilerv1alpha1.KubeFiler) Result {
	result, ready := m.isFilerContainerReady(ctx, instance)
	if result != Done {
		return result
	}

	if !ready {
		instance.Status.Phase = kubefilerv1alpha1.KubeFilerDeployed
		m.client.Status().Update(ctx, instance)
		m.recorder.Eventf(instance,
			EventWarning,
			ReasonNotRunning,
			"The Filer is not running")
	}

	return Done
}

func (m *KubeFilerManager) isFilerContainerReady(ctx context.Context, instance *kubefilerv1alpha1.KubeFiler) (Result, bool) {
	pods, err := getPodsForInstance(ctx, m.client, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			m.logger.Info("Pods not found - probably quick deletion")
			return Done, false
		}
		m.logger.Error(err, "Failed to get deployment")
		return Result{err: err}, false
	}

	if len(pods.Items) != 1 {
		m.logger.Error(nil, "Unexpected number of pods: ", len(pods.Items))
		return Result{err: errors.NewInternalError(nil)}, false
	}

	var ready bool
	if len(pods.Items[0].Status.ContainerStatuses) == 0 {
		ready = false
	} else {
		ready = true
		for _, containerStatus := range pods.Items[0].Status.ContainerStatuses {
			if !containerStatus.Ready {
				ready = false
				break
			}
		}
	}

	return Done, ready
}

// Finalize should be called when there's a finalizer on the resource
// and we need to do some cleanup.
func (m *KubeFilerManager) Finalize(ctx context.Context, instance *kubefilerv1alpha1.KubeFiler) Result {
	err := m.removeFinalizer(ctx, instance)
	if err != nil {
		return Result{err: err}
	}
	return Done
}

func (m *KubeFilerManager) addFinalizer(ctx context.Context, instance *kubefilerv1alpha1.KubeFiler) (bool, error) {
	if controllerutil.ContainsFinalizer(instance, kubeFilerFinalizer) {
		return false, nil
	}
	controllerutil.AddFinalizer(instance, kubeFilerFinalizer)
	return true, m.client.Update(ctx, instance)
}

func (m *KubeFilerManager) removeFinalizer(ctx context.Context, instance *kubefilerv1alpha1.KubeFiler) error {
	m.logger.Info("Removing finalizer")

	controllerutil.RemoveFinalizer(instance, kubeFilerFinalizer)
	return m.client.Update(ctx, instance)
}
