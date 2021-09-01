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

	kubefilerv1alpha1 "github.com/ctera/kubefiler-operator/api/v1alpha1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func getRoleBinding(ctx context.Context, client client.Client, ns, name string) (*rbacv1.RoleBinding, error) {
	roleBinding := &rbacv1.RoleBinding{}
	err := client.Get(
		ctx,
		types.NamespacedName{
			Namespace: ns,
			Name:      name,
		},
		roleBinding,
	)

	return roleBinding, err
}

func getOrCreateRoleBinding(ctx context.Context, client client.Client, instance *kubefilerv1alpha1.KubeFiler, account, role string) (*rbacv1.RoleBinding, bool, error) {
	roleBindingName := getRoleBindingName(instance)
	roleBinding, err := getRoleBinding(ctx, client, instance.GetNamespace(), roleBindingName)
	if err == nil {
		return roleBinding, false, nil
	}

	if errors.IsNotFound(err) {
		roleBinding, err = generateRoleBinding(client, instance, roleBindingName, account, role)
		if err != nil {
			return roleBinding, false, err
		}

		err = client.Create(ctx, roleBinding)
		if err != nil {
			return roleBinding, false, err
		}
		// RoleBinding created successfully
		return roleBinding, true, nil
	}
	return nil, false, err
}

func getRoleBindingName(instance *kubefilerv1alpha1.KubeFiler) string {
	return instance.GetName() + "-filer-rolebinding"
}

func generateRoleBinding(client client.Client, instance *kubefilerv1alpha1.KubeFiler, name, account, role string) (*rbacv1.RoleBinding, error) {
	roleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: instance.GetNamespace(),
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     role,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      account,
				Namespace: instance.GetNamespace(),
			},
		},
	}

	// Set the instance as the owner of the secret
	err := controllerutil.SetControllerReference(instance, roleBinding, client.Scheme())

	return roleBinding, err
}
