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
)

const kubeFilerFinalizer = "kubefiler-operator.ctera.com/kubeFilerFinalizer"

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

// Update should be called when a KubeFiler resource changes.
func (m *KubeFilerManager) Update(ctx context.Context, instance *kubefilerv1alpha1.KubeFiler) Result {
	m.logger.Info("Updating state for KubeFiler",
		"name", instance.Name,
		"UID", instance.UID)

	kubeFilerPortal, err := getKubeFilerPortal(ctx, m.client, instance.GetNamespace(), instance.Spec.Portal)
	if err != nil {
		return Result{err: err}
	}

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

	serviceAccountName, result := getOrCreateFilerRole(ctx, m.client, instance, kubeFilerPortal)
	if result != Done {
		return result
	}

	deployment, created, err := getOrCreateGatewayDeployment(ctx, m.client, m.cfg, instance, gatewaySecret, kubeFilerPortal, serviceAccountName)
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

	_, created, err = getOrCreateGatewayService(ctx, m.client, instance)
	if err != nil {
		return Result{err: err}
	} else if created {
		m.logger.Info("Created service")
		return Requeue
	}

	m.logger.Info("Done updating KubeFiler resources")
	return Done
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
