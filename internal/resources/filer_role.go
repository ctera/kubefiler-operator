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
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getOrCreateFilerRole(ctx context.Context, client client.Client, instance *kubefilerv1alpha1.KubeFiler, portal *kubefilerv1alpha1.KubeFilerPortal) (string, Result) {
	serviceAccount, created, err := getOrCreateServiceAccount(ctx, client, instance)
	if err != nil {
		return "", Result{err: err}
	} else if created {
		return "", Requeue
	}

	role, created, err := getOrCreateRole(ctx, client, instance, portal)
	if err != nil {
		return "", Result{err: err}
	} else if created {
		return "", Requeue
	}

	_, created, err = getOrCreateRoleBinding(ctx, client, instance, serviceAccount.GetName(), role.GetName())
	if err != nil {
		return "", Result{err: err}
	} else if created {
		return "", Requeue
	}

	return serviceAccount.GetName(), Done
}
