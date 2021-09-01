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
	"strings"

	kubefilerv1alpha1 "github.com/ctera/kubefiler-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	"github.com/ctera/kubefiler-operator/internal/conf"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ctera/kubefiler-operator/kubefilertest"
)

func TestGetDeployment(t *testing.T) {
	name := "testName"
	namespace := "testNamespace"

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	mockedClient := &kubefilertest.MockedClient{
		ReturnObject: deployment,
	}

	ctx := context.Background()
	obj := &appsv1.Deployment{}
	mockedClient.On("Get", ctx, types.NamespacedName{Namespace: namespace, Name: name}, obj).Return(nil)

	ret, err := getDeployment(ctx, mockedClient, namespace, name)

	mockedClient.AssertExpectations(t)
	assert.Equal(t, deployment, ret, "Returned object not as expected")
	assert.Nil(t, err)
}

func TestGetOrCreateGatewayDeployment(t *testing.T) {
	instanceNamespace := "instanceNamespace"
	instanceName := "instanceName"
	deploymentName := instanceName + "-kubefiler"

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: instanceNamespace,
		},
	}

	ctx := context.Background()
	obj := &appsv1.Deployment{}
	instance := &kubefilerv1alpha1.KubeFiler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instanceName,
			Namespace: instanceNamespace,
		},
		Spec: kubefilerv1alpha1.KubeFilerSpec{
			Storage: kubefilerv1alpha1.KubeFilerStorageSpec{
				Pvc: &kubefilerv1alpha1.KubeFilerPvcSpec{
					Name: "volume",
				},
			},
		},
	}

	cfg, _ := conf.NewSource().Read()
	gatewaySecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      strings.Join([]string{instanceNamespace, instanceName}, "-"),
			Namespace: instanceNamespace,
		},
	}

	kubeFilerPortal := &kubefilerv1alpha1.KubeFilerPortal{
		Spec: kubefilerv1alpha1.KubeFilerPortalSpec{
			Address: "192.168.1.1",
			Credentials: &kubefilerv1alpha1.KubeFilerPortalCredentialsSpec{
				Secret: "GatewaySecret",
			},
		},
	}

	serviceAccountName := instanceName + "-filer"

	testCases := []struct {
		testName            string
		clientGetReturn     error
		clientGetDeployment *appsv1.Deployment
		clientCreateReturn  error
		apiReturn           error
		created             bool
	}{
		{
			"Deployment already exists",
			nil,
			deployment,
			nil,
			nil,
			false,
		},
		{
			"Get returned an error",
			errors.NewBadRequest("Bad Request"),
			nil,
			nil,
			errors.NewBadRequest("Bad Request"),
			false,
		},
		{
			"Created successfully",
			errors.NewNotFound(schema.GroupResource{}, deploymentName),
			nil,
			nil,
			nil,
			true,
		},
		{
			"Failed to create the secret",
			errors.NewNotFound(schema.GroupResource{}, deploymentName),
			nil,
			errors.NewBadRequest("Bad Request"),
			errors.NewBadRequest("Bad Request"),
			false,
		},
	}

	for _, tc := range testCases {
		t.Logf("Running test case: %s", tc.testName)

		mockedClient := &kubefilertest.MockedClient{
			ReturnObject: tc.clientGetDeployment,
		}
		mockedClient.On("Get", ctx, types.NamespacedName{Namespace: instanceNamespace, Name: deploymentName}, obj).Return(tc.clientGetReturn)
		if errors.IsNotFound(tc.clientGetReturn) {
			mockedClient.On("Create", ctx, mock.AnythingOfType("*v1.Deployment")).Return(tc.clientCreateReturn)
		}

		_, created, err := getOrCreateGatewayDeployment(ctx, mockedClient, cfg, instance, gatewaySecret, kubeFilerPortal, serviceAccountName)

		mockedClient.AssertExpectations(t)
		assert.Equal(t, tc.created, created)
		assert.Equal(t, tc.apiReturn, err)
	}

}
