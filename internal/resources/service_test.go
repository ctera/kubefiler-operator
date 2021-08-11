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

func TestGetService(t *testing.T) {
	name := "testName"
	namespace := "testNamespace"

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	mockedClient := &kubefilertest.MockedClient{
		ReturnObject: service,
	}

	ctx := context.Background()
	mockedClient.On("Get", ctx, types.NamespacedName{Namespace: namespace, Name: name}, &corev1.Service{}).Return(nil)

	ret, err := getService(ctx, mockedClient, namespace, name)

	mockedClient.AssertExpectations(t)
	assert.Equal(t, service, ret, "Returned object not as expected")
	assert.Nil(t, err)
}

func TestGetOrCreateGatewayService(t *testing.T) {
	instanceNamespace := "instanceNamespace"
	instanceName := "instanceName"
	serviceName := instanceName + "-kubefiler"

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
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
		clientGetService   *corev1.Service
		clientCreateReturn error
		apiReturn          error
		created            bool
	}{
		{
			"Service already exists",
			nil,
			service,
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
			errors.NewNotFound(schema.GroupResource{}, serviceName),
			nil,
			nil,
			nil,
			true,
		},
		{
			"Failed to create the service",
			errors.NewNotFound(schema.GroupResource{}, serviceName),
			nil,
			errors.NewBadRequest("Bad Request"),
			errors.NewBadRequest("Bad Request"),
			false,
		},
	}

	for _, tc := range testCases {
		t.Logf("Running test case: %s", tc.testName)

		mockedClient := &kubefilertest.MockedClient{
			ReturnObject: tc.clientGetService,
		}
		mockedClient.On("Get", ctx, types.NamespacedName{Namespace: instanceNamespace, Name: serviceName}, &corev1.Service{}).Return(tc.clientGetReturn)
		if errors.IsNotFound(tc.clientGetReturn) {
			mockedClient.On("Create", ctx, mock.AnythingOfType("*v1.Service")).Return(tc.clientCreateReturn)
		}
		_, created, err := getOrCreateGatewayService(ctx, mockedClient, instance)

		mockedClient.AssertExpectations(t)
		assert.Equal(t, tc.created, created)
		assert.Equal(t, tc.apiReturn, err)
	}

}
