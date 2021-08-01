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

func TestGetPvc(t *testing.T) {
	name := "testName"
	namespace := "testNamespace"

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	mockedClient := &kubefilertest.MockedClient{
		ReturnObject: pvc,
	}

	ctx := context.Background()
	mockedClient.On("Get", ctx, types.NamespacedName{Namespace: namespace, Name: name}, &corev1.PersistentVolumeClaim{}).Return(nil)

	ret, err := getPvc(ctx, mockedClient, namespace, name)

	mockedClient.AssertExpectations(t)
	assert.Equal(t, pvc, ret, "Returned object not as expected")
	assert.Nil(t, err)
}

func TestGetOrCreateGatewayPvc(t *testing.T) {
	instanceNamespace := "instanceNamespace"
	instanceName := "instanceName"
	serviceName := instanceName + "-kubefiler-pvc"

	pvc := &corev1.PersistentVolumeClaim{
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
					Spec: &corev1.PersistentVolumeClaimSpec{},
				},
			},
		},
	}

	testCases := []struct {
		testName           string
		clientGetReturn    error
		clientGetPvc       *corev1.PersistentVolumeClaim
		clientCreateReturn error
		apiReturn          error
		created            bool
	}{
		{
			"Pvc already exists",
			nil,
			pvc,
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
			"Failed to create the pvc",
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
			ReturnObject: tc.clientGetPvc,
		}
		mockedClient.On("Get", ctx, types.NamespacedName{Namespace: instanceNamespace, Name: serviceName}, &corev1.PersistentVolumeClaim{}).Return(tc.clientGetReturn)
		if errors.IsNotFound(tc.clientGetReturn) {
			mockedClient.On("Create", ctx, mock.AnythingOfType("*v1.PersistentVolumeClaim")).Return(tc.clientCreateReturn)
		}
		_, created, err := getOrCreateGatewayPvc(ctx, mockedClient, instance)

		mockedClient.AssertExpectations(t)
		assert.Equal(t, tc.created, created)
		assert.Equal(t, tc.apiReturn, err)
	}

}

func TestKubeFilerNeedsPvc(t *testing.T) {
	testCases := []struct {
		testName       string
		storageSpec    *kubefilerv1alpha1.KubeFilerStorageSpec
		expectedReturn bool
	}{
		{
			"No PVC",
			&kubefilerv1alpha1.KubeFilerStorageSpec{},
			false,
		},
		{
			"With PVC, Without Spec",
			&kubefilerv1alpha1.KubeFilerStorageSpec{
				Pvc: &kubefilerv1alpha1.KubeFilerPvcSpec{},
			},
			false,
		},
		{
			"With PVC and Spec",
			&kubefilerv1alpha1.KubeFilerStorageSpec{
				Pvc: &kubefilerv1alpha1.KubeFilerPvcSpec{
					Spec: &corev1.PersistentVolumeClaimSpec{},
				},
			},
			true,
		},
	}

	for _, tc := range testCases {
		t.Logf("Running test case: %s", tc.testName)

		instance := &kubefilerv1alpha1.KubeFiler{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "instanceName",
				Namespace: "instanceNamespace",
			},
			Spec: kubefilerv1alpha1.KubeFilerSpec{
				Storage: *tc.storageSpec,
			},
		}

		assert.Equal(t, tc.expectedReturn, kubeFilerNeedsPvc(instance))
	}
}

func TestGetGatewayPvcName(t *testing.T) {
	instanceName := "instanceName"
	pvcName := "testPvcName"

	testCases := []struct {
		testName       string
		storageSpec    *kubefilerv1alpha1.KubeFilerStorageSpec
		expectedReturn string
	}{
		{
			"With Name",
			&kubefilerv1alpha1.KubeFilerStorageSpec{
				Pvc: &kubefilerv1alpha1.KubeFilerPvcSpec{
					Name: pvcName,
				},
			},
			pvcName,
		},
		{
			"Without Name",
			&kubefilerv1alpha1.KubeFilerStorageSpec{
				Pvc: &kubefilerv1alpha1.KubeFilerPvcSpec{},
			},
			instanceName + "-kubefiler-pvc",
		},
	}

	for _, tc := range testCases {
		t.Logf("Running test case: %s", tc.testName)

		instance := &kubefilerv1alpha1.KubeFiler{
			ObjectMeta: metav1.ObjectMeta{
				Name:      instanceName,
				Namespace: "instanceNamespace",
			},
			Spec: kubefilerv1alpha1.KubeFilerSpec{
				Storage: *tc.storageSpec,
			},
		}

		assert.Equal(t, tc.expectedReturn, getGatewayPvcName(instance))
	}
}
