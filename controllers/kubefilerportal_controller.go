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

package controllers

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kubefilerv1alpha1 "github.com/ctera/ctera-gateway-operator/api/v1alpha1"
)

// KubeFilerPortalReconciler reconciles a KubeFilerPortal object
type KubeFilerPortalReconciler struct {
	client.Client
}

//+kubebuilder:rbac:groups=kubefiler.ctera.com,resources=kubefilerportals,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubefiler.ctera.com,resources=kubefilerportals/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubefiler.ctera.com,resources=kubefilerportals/finalizers,verbs=update

// Reconcile The KubeFilerPortal resource is only configuration. Currently, updating it does not have any effect
func (*KubeFilerPortalReconciler) Reconcile(ctx context.Context, _ ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KubeFilerPortalReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubefilerv1alpha1.KubeFilerPortal{}).
		Complete(r)
}
