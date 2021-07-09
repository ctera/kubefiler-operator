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

	kubefilerv1alpha1 "github.com/ctera/ctera-gateway-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/sethvargo/go-password/password"
)

const (
	// UsernameKey is the name of the key to read the username from when reading the secret
	GatewayUsernameKey = "username"
	// GatewayDefaultUsername is the username for the secret
	GatewayDefaultUsername = "admin"
	// GatewayPasswordKey is the name of the key to read the password from when reading the secret
	GatewayPasswordKey = "password"

	// GatewayPasswordLength is the length of the generated password
	GatewayPasswordLength = 16
	// GatewayPasswordNumDigits is the amount of digits in the generated password
	GatewayPasswordNumDigits = 2
	// GatewayPasswordNumSymbols is the amount of symbols in the generated password
	GatewayPasswordNumSymbols = 2
	// GatewayPasswordNoUpper sets whether upper case letters may be used
	GatewayPasswordNoUpper = false
	// GatewayPasswordAllowRepeat sets whether the same charcter may repeat
	GatewayPasswordAllowRepeat = false
)

func getSecret(ctx context.Context, client client.Client, ns, name string) (*corev1.Secret, error) {
	secret := &corev1.Secret{}
	err := client.Get(
		ctx,
		types.NamespacedName{
			Namespace: ns,
			Name:      name,
		},
		secret,
	)

	return secret, err
}

func getOrCreateGatewaySecret(ctx context.Context, client client.Client, instance *kubefilerv1alpha1.KubeFiler, ns string) (*corev1.Secret, bool, error) {
	secretName := getGatewaySecretName(instance)

	// fetch the existing secret, if available
	secret, err := getSecret(ctx, client, ns, secretName)
	if err == nil {
		return secret, false, nil
	}

	if errors.IsNotFound(err) {
		secret, err = generateGatewaySecret(client, instance, ns, secretName)
		if err != nil {
			return secret, false, err
		}

		err = client.Create(ctx, secret)
		if err != nil {
			return secret, false, err
		}
		// Deployment created successfully
		return secret, true, nil
	}

	return nil, false, err
}

func getGatewaySecretName(instance *kubefilerv1alpha1.KubeFiler) string {
	return strings.Join(
		[]string{
			instance.GetNamespace(),
			instance.GetName(),
		},
		"-",
	)
}

func generateGatewaySecret(client client.Client, instance *kubefilerv1alpha1.KubeFiler, ns, name string) (*corev1.Secret, error) {
	pass, err := password.Generate(
		GatewayPasswordLength,
		GatewayPasswordNumDigits,
		GatewayPasswordNumSymbols,
		GatewayPasswordNoUpper,
		GatewayPasswordAllowRepeat,
	)
	if err != nil {
		return nil, err
	}

	immutable := true
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Immutable: &immutable,
		StringData: map[string]string{
			GatewayUsernameKey: GatewayDefaultUsername,
			GatewayPasswordKey: pass,
		},
	}

	// Set the instance as the owner of the secret
	err = controllerutil.SetControllerReference(instance, secret, client.Scheme())

	return secret, err
}
