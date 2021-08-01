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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ctera/kubefiler-operator/kubefilertest"
)

func TestGetSecret(t *testing.T) {
	name := "testName"
	namespace := "testNamespace"

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	mockedClient := &kubefilertest.MockedClient{
		ReturnObject: secret,
	}

	ctx := context.Background()
	mockedClient.On("Get", ctx, types.NamespacedName{Namespace: namespace, Name: name}, &corev1.Secret{}).Return(nil)

	ret, err := getSecret(ctx, mockedClient, namespace, name)

	mockedClient.AssertExpectations(t)
	assert.Equal(t, secret, ret, "Returned object not as expected")
	assert.Nil(t, err)
}

func TestGetOrCreateGatewaySecret(t *testing.T) {
	instanceNamespace := "instanceNamespace"
	instanceName := "instanceName"
	secretName := instanceName + "-kubefiler-credentials"

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: instanceNamespace,
		},
	}

	ctx := context.Background()
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

	testCases := []struct {
		testName           string
		clientGetReturn    error
		clientGetSecret    *corev1.Secret
		clientCreateReturn error
		apiReturn          error
		created            bool
	}{
		{
			"Secret already exists",
			nil,
			secret,
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
			errors.NewNotFound(schema.GroupResource{}, secretName),
			nil,
			nil,
			nil,
			true,
		},
		{
			"Failed to create the secret",
			errors.NewNotFound(schema.GroupResource{}, secretName),
			nil,
			errors.NewBadRequest("Bad Request"),
			errors.NewBadRequest("Bad Request"),
			false,
		},
	}

	for _, tc := range testCases {
		t.Logf("Running test case: %s", tc.testName)

		mockedClient := &kubefilertest.MockedClient{
			ReturnObject: tc.clientGetSecret,
		}
		mockedClient.On("Get", ctx, types.NamespacedName{Namespace: instanceNamespace, Name: secretName}, &corev1.Secret{}).Return(tc.clientGetReturn)
		if errors.IsNotFound(tc.clientGetReturn) {
			mockedClient.On("Create", ctx, mock.AnythingOfType("*v1.Secret")).Return(tc.clientCreateReturn)
		}
		_, created, err := getOrCreateGatewaySecret(ctx, mockedClient, instance)

		mockedClient.AssertExpectations(t)
		assert.Equal(t, tc.created, created)
		assert.Equal(t, tc.apiReturn, err)
	}

}
