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

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kubefilerv1alpha1 "github.com/ctera/kubefiler-operator/api/v1alpha1"
	"github.com/ctera/kubefiler-operator/internal/resources"
)

// KubeFilerReconciler reconciles a KubeFiler object
type KubeFilerReconciler struct {
	client.Client
	Log      logr.Logger
	recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=kubefiler.ctera.com,resources=kubefilers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubefiler.ctera.com,resources=kubefilers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubefiler.ctera.com,resources=kubefilers/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles;rolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list
// +kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=events,verbs=create

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the KubeFiler object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *KubeFilerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := r.Log.WithValues("kubefiler", req.NamespacedName)
	reqLogger.Info("Reconciling KubeFiler")

	kubeFilerManager := resources.NewKubeFilerManager(r, r.recorder, reqLogger)
	res := kubeFilerManager.Process(ctx, req.NamespacedName)
	err := res.Err()
	if res.Requeue() {
		return ctrl.Result{Requeue: true}, err
	}
	return ctrl.Result{}, err
}

func (r *KubeFilerReconciler) setRecorder(mgr ctrl.Manager) {
	if r.recorder == nil {
		r.recorder = mgr.GetEventRecorderFor("kubefiler-controller")
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *KubeFilerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.setRecorder(mgr)
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubefilerv1alpha1.KubeFiler{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
