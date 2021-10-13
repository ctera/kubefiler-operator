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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	kubefilerv1alpha1 "github.com/ctera/kubefiler-operator/api/v1alpha1"

	ctera "github.com/ctera/ctera-gateway-openapi-go-client"
	"github.com/ctera/kubefiler-operator/internal/cteraclient"
)

func getOrCreateShare(cteraClient *cteraclient.CteraClient, instance *kubefilerv1alpha1.KubeFilerExport, trustedNfsClients []ctera.NFSv3AccessControlEntry) (*ctera.Share, bool, error) {
	share, err := cteraClient.GetShareSafe(instance.GetName())
	if err != nil {
		return nil, false, err
	}

	if share != nil {
		if canReuseShare(share, instance.Spec.Path) {
			return share, false, nil
		}
		return nil, false, status.Errorf(codes.AlreadyExists, "Share already exists with different parameters. Share Path=%s, Requested Path=%s", share.GetDirectory(), instance.Spec.Path)
	}

	if trustedNfsClients == nil {
		trustedNfsClients = make([]ctera.NFSv3AccessControlEntry, 0)
	}
	trustedNfsClients = append(trustedNfsClients, *ctera.NewNFSv3AccessControlEntry("127.0.0.1", "255.255.255.255", ctera.RW))

	shareUuid := instance.Annotations[shareUuidAnnotation]
	share, err = cteraClient.CreateShare(instance.GetName(), instance.Spec.Path, &shareUuid, trustedNfsClients)
	if err != nil {
		return nil, false, status.Error(codes.Internal, err.Error())
	}

	return share, true, nil
}

func canReuseShare(share *ctera.Share, path string) bool {
	// The directory in the object is absolute while the path is relative
	return share.GetDirectory() == ("/" + path)
}

func removeShare(cteraClient *cteraclient.CteraClient, instance *kubefilerv1alpha1.KubeFilerExport) (bool, error) {
	share, err := cteraClient.GetShareSafe(instance.GetName())
	if err != nil {
		return false, err
	} else if share == nil {
		return false, nil
	}

	err = cteraClient.DeleteShareSafe(share.Name)
	if err != nil {
		return false, err
	}

	return true, nil
}
