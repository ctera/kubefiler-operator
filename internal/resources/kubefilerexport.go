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

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	kubefilerv1alpha1 "github.com/ctera/kubefiler-operator/api/v1alpha1"
	"github.com/ctera/kubefiler-operator/internal/conf"
	"github.com/ctera/kubefiler-operator/internal/cteraclient"
)

const kubeFilerExportFinalizer = "kubefiler-operator.ctera.com/kubeFilerExportFinalizer"

// KubeFilerExportManager is used to manage KubeFilerExport resources.
type KubeFilerExportManager struct {
	client   client.Client
	recorder record.EventRecorder
	logger   Logger
	cfg      *conf.OperatorConfig
}

// NewKubeFilerExportManager creates a KubeFilerExportManager.
func NewKubeFilerExportManager(client client.Client, recorder record.EventRecorder, logger Logger) *KubeFilerExportManager {
	return &KubeFilerExportManager{
		client:   client,
		recorder: recorder,
		logger:   logger,
		cfg:      conf.Get(),
	}
}

// Process is called by the controller on any type of reconciliation.
func (m *KubeFilerExportManager) Process(ctx context.Context, nsname types.NamespacedName) Result {
	// fetch our resource to determine what to do next
	instance := &kubefilerv1alpha1.KubeFilerExport{}
	err := m.client.Get(ctx, nsname, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found. Not a fatal error.
			return Done
		}
		m.logger.Error(err, "get failed for KubeFilerExport",
			"ns", nsname.Namespace,
			"name", nsname.Name)
		return Result{err: err}
	}

	// now that we have the resource. determine if its alive or pending deletion
	if instance.GetDeletionTimestamp() != nil {
		// its being deleted
		if controllerutil.ContainsFinalizer(instance, kubeFilerExportFinalizer) {
			// and our finalizer is present
			return m.Finalize(ctx, instance)
		}
		return Done
	}
	// resource is alive
	return m.Update(ctx, instance)
}

// Update should be called when a KubeFiler resource changes.
func (m *KubeFilerExportManager) Update(ctx context.Context, instance *kubefilerv1alpha1.KubeFilerExport) Result {
	m.logger.Info("Updating state for KubeFilerExport",
		"name", instance.Name,
		"UID", instance.UID)

	filerAddress, filerUsername, filerPassword, err := m.getFilerLoginDetails(ctx, instance)
	if err != nil {
		m.logger.Error(err, "Failed to get the Filer's login details")
		return Result{err: err}
	}

	changed, err := m.addFinalizer(ctx, instance)
	if err != nil {
		return Result{err: err}
	} else if changed {
		m.logger.Info("Added finalizer")
		return Requeue
	}

	cteraClient, err := cteraclient.GetAuthenticatedCteraClient(ctx, m.logger, filerAddress, filerUsername, filerPassword)
	if err != nil {
		return Result{err: err}
	}

	_, created, err := getOrCreateShare(cteraClient, instance)
	if err != nil {
		return Result{err: err}
	} else if created {
		m.logger.Info("Created share")
		return Requeue
	}

	m.logger.Info("Done updating KubeFilerExport resources")
	return Done
}

// Finalize should be called when there's a finalizer on the resource
// and we need to do some cleanup.
func (m *KubeFilerExportManager) Finalize(ctx context.Context, instance *kubefilerv1alpha1.KubeFilerExport) Result {
	filerAddress, filerUsername, filerPassword, err := m.getFilerLoginDetails(ctx, instance)
	if err != nil && !errors.IsNotFound(err) {
		m.logger.Error(err, "Failed to get the Filer's login details")
		return Result{err: err}
	}
	if filerAddress != "" {
		cteraClient, err := cteraclient.GetAuthenticatedCteraClient(ctx, m.logger, filerAddress, filerUsername, filerPassword)
		if err != nil {
			return Result{err: err}
		}

		deleted, err := removeShare(cteraClient, instance)
		if err != nil {
			return Result{err: err}
		} else if deleted {
			m.logger.Info("Share removed")
			return Requeue
		}
	}

	err = m.removeFinalizer(ctx, instance)
	if err != nil {
		return Result{err: err}
	}
	return Done
}

func (m *KubeFilerExportManager) getFilerLoginDetails(ctx context.Context, instance *kubefilerv1alpha1.KubeFilerExport) (string, string, string, error) {
	m.logger.Info("Getting the KubeFiler associated with the KubeFilerExport",
		"ns", instance.GetNamespace(),
		"name", instance.Spec.KubeFiler,
	)
	kubeFiler := &kubefilerv1alpha1.KubeFiler{}
	err := m.client.Get(
		ctx,
		types.NamespacedName{
			Namespace: instance.GetNamespace(),
			Name:      instance.Spec.KubeFiler,
		},
		kubeFiler,
	)
	if err != nil {
		m.logger.Error(err, "get failed for KubeFiler",
			"ns", instance.GetNamespace(),
			"name", instance.Spec.KubeFiler)
		return "", "", "", err
	}

	kubeFilerSecret, err := getSecret(ctx, m.client, instance.GetNamespace(), getGatewaySecretName(kubeFiler))
	if err != nil {
		m.logger.Error(err, "Failed to get Filer secret")
		return "", "", "", err
	}

	kubeFilerService, err := getService(ctx, m.client, instance.GetNamespace(), getGatewayServiceName(kubeFiler))
	if err != nil {
		m.logger.Error(err, "Failed to get Filer service")
		return "", "", "", err
	}

	return kubeFilerService.Spec.ClusterIP,
		string(kubeFilerSecret.Data[GatewayUsernameKey]),
		string(kubeFilerSecret.Data[GatewayPasswordKey]),
		nil
}

func (m *KubeFilerExportManager) addFinalizer(ctx context.Context, instance *kubefilerv1alpha1.KubeFilerExport) (bool, error) {
	if controllerutil.ContainsFinalizer(instance, kubeFilerExportFinalizer) {
		return false, nil
	}
	controllerutil.AddFinalizer(instance, kubeFilerExportFinalizer)
	return true, m.client.Update(ctx, instance)
}

func (m *KubeFilerExportManager) removeFinalizer(ctx context.Context, instance *kubefilerv1alpha1.KubeFilerExport) error {
	m.logger.Info("Removing finalizer")

	controllerutil.RemoveFinalizer(instance, kubeFilerExportFinalizer)
	return m.client.Update(ctx, instance)
}
