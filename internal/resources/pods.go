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
	kubefilerv1alpha1 "github.com/ctera/kubefiler-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/ctera/kubefiler-operator/internal/conf"
)

const (
	// RunDirMountPath is the path for the "run" directory inside the container
	RunDirMountPath = "/run"
	// TmpDirMountPath is the path for the "tmp" directory inside the container
	TmpDirMountPath = "/tmp"
	// CgroupDirLocalPath is the path for the "cgroup" directory on the Host machine
	CgroupDirLocalPath = "/sys/fs/cgroup"
	// CgroupDirMountPath is the path for the "cgroup" directory inside the container
	CgroupDirMountPath = "/sys/fs/cgroup"
	// kubeFilerInitCommand is the path for the Gateway initialization script
	kubeFilerInitCommand = "/kubefiler_init.py"
)

func buildGatewayPodSpec(cfg *conf.OperatorConfig, instance *kubefilerv1alpha1.KubeFiler, gatewaySecretName string, serviceAccountName string) corev1.PodSpec {
	volumes, mounts := getPodVolumesAndMounts(cfg, instance)
	filerPodEnv := getFilerPodEnv(instance)
	openAPIPodEnv := getOpenAPIPodEnv(gatewaySecretName)
	privileged := true

	podSpec := corev1.PodSpec{
		ServiceAccountName: serviceAccountName,
		Volumes:            volumes,
		Containers: []corev1.Container{
			{
				Name:            cfg.GatewayContainerName,
				Image:           cfg.GatewayContainerImage,
				ImagePullPolicy: corev1.PullAlways,
				Ports: []corev1.ContainerPort{{
					ContainerPort: 443,
					Name:          "mgmt",
				}},
				SecurityContext: &corev1.SecurityContext{
					Privileged: &privileged,
				},
				Env:          filerPodEnv,
				VolumeMounts: mounts,
				LivenessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						TCPSocket: &corev1.TCPSocketAction{
							Port: intstr.FromInt(443),
						},
					},
				},
				Lifecycle: &corev1.Lifecycle{
					PostStart: &corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: []string{kubeFilerInitCommand},
						},
					},
				},
			},
			{
				Name:            cfg.GatewayOpenAPIContainerName,
				Image:           cfg.GatewayOpenAPIContainerImage,
				ImagePullPolicy: corev1.PullAlways,
				Ports: []corev1.ContainerPort{{
					ContainerPort: 9090,
					Name:          "api",
				}},
				Env: openAPIPodEnv,
			},
		},
	}
	return podSpec
}

func getPodVolumesAndMounts(cfg *conf.OperatorConfig, instance *kubefilerv1alpha1.KubeFiler) ([]corev1.Volume, []corev1.VolumeMount) {
	volumes := []corev1.Volume{}
	mounts := []corev1.VolumeMount{}

	shareVol, shareMount := storageVolumeAndMount(cfg, instance)
	volumes = append(volumes, shareVol)
	mounts = append(mounts, shareMount)

	osRunVol, osRunMount := osRunVolumeAndMount(instance.GetName())
	volumes = append(volumes, osRunVol)
	mounts = append(mounts, osRunMount)

	osTmpVol, osTmpMount := osTmpVolumeAndMount(instance.GetName())
	volumes = append(volumes, osTmpVol)
	mounts = append(mounts, osTmpMount)

	cgroupVol, cgroupMount := cgroupVolumeAndMount(instance.GetName())
	volumes = append(volumes, cgroupVol)
	mounts = append(mounts, cgroupMount)

	return volumes, mounts
}

func storageVolumeAndMount(cfg *conf.OperatorConfig, instance *kubefilerv1alpha1.KubeFiler) (corev1.Volume, corev1.VolumeMount) {
	// volume
	pvcName := instance.Spec.Storage.Pvc.Name
	pvcVolName := instance.GetName() + "-storage"

	volume := corev1.Volume{
		Name: pvcVolName,
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: pvcName,
			},
		},
	}
	// mount
	mount := corev1.VolumeMount{
		MountPath: cfg.GatewayStorageMountPath,
		Name:      pvcVolName,
	}
	return volume, mount
}

func osRunVolumeAndMount(instaceName string) (corev1.Volume, corev1.VolumeMount) {
	volumeName := instaceName + "-run"
	return memoryBackedVolumeAndMount(volumeName, RunDirMountPath)
}

func osTmpVolumeAndMount(instaceName string) (corev1.Volume, corev1.VolumeMount) {
	volumeName := instaceName + "-tmp"
	return memoryBackedVolumeAndMount(volumeName, TmpDirMountPath)
}

func memoryBackedVolumeAndMount(volumeName, mountPath string) (corev1.Volume, corev1.VolumeMount) {
	// volume
	volume := corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{
				Medium: corev1.StorageMediumMemory,
			},
		},
	}
	// mount
	mount := corev1.VolumeMount{
		MountPath: mountPath,
		Name:      volumeName,
	}
	return volume, mount
}

func cgroupVolumeAndMount(instaceName string) (corev1.Volume, corev1.VolumeMount) {
	volumeName := instaceName + "-cgroup"
	return localPathVolumeAndMount(volumeName, CgroupDirLocalPath, CgroupDirMountPath, true, corev1.HostPathDirectory)
}

func localPathVolumeAndMount(volumeName, localPath, mountPath string, readOnly bool, volumeType corev1.HostPathType) (corev1.Volume, corev1.VolumeMount) {
	// volume
	volume := corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: localPath,
				Type: &volumeType,
			},
		},
	}
	// mount
	mount := corev1.VolumeMount{
		MountPath: mountPath,
		Name:      volumeName,
		ReadOnly:  readOnly,
	}
	return volume, mount
}

func getFilerPodEnv(instance *kubefilerv1alpha1.KubeFiler) []corev1.EnvVar {
	env := []corev1.EnvVar{
		{
			Name:  "FILER_KUBEFILER_NAME",
			Value: instance.GetName(),
		},
	}
	return env
}

func getOpenAPIPodEnv(secretName string) []corev1.EnvVar {
	env := []corev1.EnvVar{
		{
			Name: "CTERA_FILER_JWT_SECRET",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: secretName,
					},
					Key: GatewayJwtSecretKey,
				},
			},
		},
		{
			Name: "CTERA_FILER_TRUST_SSL",
			// TODO Get from KubeFiler
			Value: "Trust",
		},
	}
	return env
}
