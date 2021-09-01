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

func getRole(ctx context.Context, client client.Client, ns, name string) (*rbacv1.Role, error) {
	role := &rbacv1.Role{}
	err := client.Get(
		ctx,
		types.NamespacedName{
			Namespace: ns,
			Name:      name,
		},
		role,
	)

	return role, err
}

func getOrCreateRole(ctx context.Context, client client.Client, instance *kubefilerv1alpha1.KubeFiler, portal *kubefilerv1alpha1.KubeFilerPortal) (*rbacv1.Role, bool, error) {
	roleName := getRoleName(instance)
	role, err := getRole(ctx, client, instance.GetNamespace(), roleName)
	if err == nil {
		return role, false, nil
	}

	if errors.IsNotFound(err) {
		role, err = generateRole(client, instance, portal, roleName)
		if err != nil {
			return role, false, err
		}

		err = client.Create(ctx, role)
		if err != nil {
			return role, false, err
		}
		// RoleBinding created successfully
		return role, true, nil
	}
	return nil, false, err
}

func getRoleName(instance *kubefilerv1alpha1.KubeFiler) string {
	return instance.GetName() + "-filer-role"
}

func generateRole(client client.Client, instance *kubefilerv1alpha1.KubeFiler, portal *kubefilerv1alpha1.KubeFilerPortal, name string) (*rbacv1.Role, error) {
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: instance.GetNamespace(),
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups:     []string{"kubefiler.ctera.com"},
				Resources:     []string{"kubefilers"},
				Verbs:         []string{"get"},
				ResourceNames: []string{instance.GetName()},
			},
			{
				APIGroups:     []string{"kubefiler.ctera.com"},
				Resources:     []string{"kubefilerportals"},
				Verbs:         []string{"get"},
				ResourceNames: []string{instance.Spec.Portal},
			},
			{
				APIGroups: []string{"kubefiler.ctera.com"},
				Resources: []string{"kubefilerexports"},
				Verbs:     []string{"get", "list"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"get"},
				ResourceNames: []string{
					getGatewaySecretName(instance),
					portal.Spec.Credentials.Secret,
				},
			},
		},
	}

	// Set the instance as the owner of the secret
	err := controllerutil.SetControllerReference(instance, role, client.Scheme())

	return role, err
}
