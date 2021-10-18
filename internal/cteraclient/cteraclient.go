/*
Copyright 2021, CTERA Networks.

Portions Copyright 2019 The Kubernetes Authors.

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

package cteraclient

import (
	"context"
	"fmt"

	ctera "github.com/ctera/ctera-gateway-openapi-go-client"
)

// Error error type for Ctera Client errors
type Error struct {
	error string
}

// Error returns non-empty string if there was an error.
func (e Error) Error() string {
	return e.error
}

// CteraClient wrapper on top the actual ctera-openapi
type CteraClient struct {
	client *ctera.APIClient
	ctx    *context.Context
	logger Logger
}

// GetAuthenticatedCteraClient Create a CteraClient object and login to the filer with the provided credentials
func GetAuthenticatedCteraClient(ctx context.Context, logger Logger, filerAddress, username, password string) (*CteraClient, error) {
	client, err := NewCteraClient(logger, filerAddress)
	if err != nil {
		return nil, err
	}

	err = client.Authenticate(ctx, username, password)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// NewCteraClient Get an unauthenticated CteraClient object
func NewCteraClient(logger Logger, filerAddress string) (*CteraClient, error) {
	configuration := ctera.NewConfiguration()
	configuration.Host = fmt.Sprintf("%s:9090", filerAddress)
	configuration.Servers = ctera.ServerConfigurations{
		{
			URL:         fmt.Sprintf("http://%s:9090/v1.0", filerAddress),
			Description: "Main address",
		},
	}

	return &CteraClient{
		client: ctera.NewAPIClient(configuration),
		ctx:    nil,
		logger: logger,
	}, nil
}

// Authenticate login to the filer with the provided credentials
func (c *CteraClient) Authenticate(ctx context.Context, username, password string) error {
	if c.ctx != nil {
		return Error{
			error: "Already authenticated",
		}
	}

	credentials := ctera.NewCredentials(username, password)
	jwt, _, err := c.client.LoginApi.LoginPost(ctx).Credentials(*credentials).Execute()
	if err != nil {
		c.logger.Error(err, "Failed to login")
		return err
	}

	authenticatedCtx := context.WithValue(ctx, ctera.ContextAccessToken, jwt)
	c.ctx = &authenticatedCtx

	return nil
}

func (c *CteraClient) InitializeFiler(ctx context.Context, username, password string) (bool, error) {
	if c.ctx != nil {
		return false, Error{error: "Already authenticated"}
	}

	initialized, _, err := c.client.InitializationApi.CteraGatewayOpenapiApiInitializedIsInitialized(ctx).Execute()
	if err != nil {
		c.logger.Error(err, "Failed to check if the Filer is initialized")
		return false, err
	}

	if initialized {
		return false, nil
	}

	user := ctera.NewUser(username)
	user.Password = &password
	_, err = c.client.UsersApi.CteraGatewayOpenapiApiUsersAddFirstUser(ctx).User(*user).Execute()
	if err != nil {
		c.logger.Error(err, "Failed to initialize the Filer")
		return false, err
	}

	return true, nil
}

func (c *CteraClient) SetLinuxAvoidUsingFanotify(newValue bool) (bool, error) {
	value, _, err := c.client.SyncApi.CteraGatewayOpenapiApiSyncGetLinuxAvoidUsingFanotify(*c.ctx).Execute()
	if err != nil {
		c.logger.Error(err, "Failed to check value")
		return false, err
	}

	if value == newValue {
		return false, nil
	}

	if newValue {
		_, err = c.client.SyncApi.CteraGatewayOpenapiApiSyncSetLinuxAvoidUsingFanotify(*c.ctx).Execute()
	} else {
		_, err = c.client.SyncApi.CteraGatewayOpenapiApiSyncUnsetLinuxAvoidUsingFanotify(*c.ctx).Execute()
	}
	if err != nil {
		c.logger.Error(err, "Failed to set LinuxAvoidUsingFanotify to: ", newValue)
		return false, err
	}

	return true, nil
}

func (c *CteraClient) ConnectToPortal(portalAddress, portalUser, portalPassword string, trust bool) (bool, error) {
	connectionStatus, _, err := c.client.ServicesApi.CteraGatewayOpenapiApiServicesGetStatus(*c.ctx).Execute()
	if err != nil {
		c.logger.Error(err, "Failed to check is already connected to Portal")
		return false, err
	} else if connectionStatus.Connected {
		if *connectionStatus.ServerAddress == portalAddress {
			return false, nil
		}
		err = Error{error: "Already connected"}
		c.logger.Error(err, "Filer is connected to: ", portalAddress)
		return false, err
	}

	portalConnectionDetails := ctera.NewPortalConnectionDetails(portalAddress, portalUser)
	portalConnectionDetails.Password = &portalPassword
	portalConnectionDetails.Trust = &trust

	_, err = c.client.ServicesApi.CteraGatewayOpenapiApiServicesConnect(*c.ctx).PortalConnectionDetails(*portalConnectionDetails).Execute()
	if err != nil {
		c.logger.Error(err, "Failed to connect to Portal")
		return false, err
	}

	return true, nil
}

func (c *CteraClient) EnableCache() (bool, error) {
	cacheIsEnabled, _, err := c.client.CacheApi.CteraGatewayOpenapiApiCacheIsEnabled(*c.ctx).Execute()
	if err != nil {
		c.logger.Error(err, "Failed to check is cache is enabled")
		return false, err
	} else if cacheIsEnabled {
		return false, nil
	}

	_, err = c.client.CacheApi.CteraGatewayOpenapiApiCacheEnable(*c.ctx).Execute()
	if err != nil {
		c.logger.Error(err, "Failed to enable cache")
		return false, err
	}

	return true, nil
}

func (c *CteraClient) UnsuspendSync() (bool, error) {
	syncIsEnabled, _, err := c.client.SyncApi.CteraGatewayOpenapiApiSyncIsEnabled(*c.ctx).Execute()
	if err != nil {
		c.logger.Error(err, "Failed to check is sync is enabled")
		return false, err
	} else if syncIsEnabled {
		return false, nil
	}

	_, err = c.client.SyncApi.CteraGatewayOpenapiApiSyncUnsuspend(*c.ctx).Execute()
	if err != nil {
		c.logger.Error(err, "Failed to unsuspend sync")
		return false, err
	}

	return true, nil
}

func (c *CteraClient) RefreshFolders() error {
	_, err := c.client.SyncApi.CteraGatewayOpenapiApiSyncRefresh(*c.ctx).Execute()
	if err != nil {
		c.logger.Error(err, "Failed to refresh folders")
		return err
	}

	return nil
}

// GetShareSafe returns the share by name. If the share does not exist returns nil without an error
func (c *CteraClient) GetShareSafe(name string) (*ctera.Share, error) {
	share, _, err := c.client.SharesApi.SharesNameGet(*c.ctx, name).Execute()
	if err != nil {
		if c.getStatusFromError(err) != 404 {
			return nil, err
		}
		return nil, nil
	}
	return &share, nil
}

// CreateShare creates a share with the provided name and path and exports it to NFS
func (c *CteraClient) CreateShare(name, path string, uuid *string, trustedNfsClients []ctera.NFSv3AccessControlEntry) (*ctera.Share, error) {
	share := ctera.NewShare(name)
	share.Directory = &path
	exportToNfs := true
	share.ExportToNfs = &exportToNfs
	if trustedNfsClients != nil {
		share.TrustedNfsClients = &trustedNfsClients
	}
	if uuid != nil {
		share.Uuid = uuid
	}
	// TODO Do we need to override any default parameters

	_, err := c.client.SharesApi.SharesPost(*c.ctx).Share(*share).Execute()
	if err != nil {
		return nil, err
	}

	return c.GetShareSafe(name)
}

// DeleteShareSafe deletes the share. Does not return an error if the share does not exists
func (c *CteraClient) DeleteShareSafe(name string) error {
	_, err := c.client.SharesApi.SharesNameDelete(*c.ctx, name).Execute()
	if err != nil {
		if c.getStatusFromError(err) != 404 {
			return err
		}
	}
	return nil
}

// AddTrustedNfsClient adds a trusted NFS client definition to the share
func (c *CteraClient) AddTrustedNfsClient(shareName, address, netmask string, perm ctera.FileAccessMode) error {
	trustedNfsClients := []ctera.NFSv3AccessControlEntry{*ctera.NewNFSv3AccessControlEntry(address, netmask, perm)}
	_, err := c.client.SharesApi.CteraGatewayOpenapiApiSharesAddTrustedNfsClients(*c.ctx, shareName).NFSv3AccessControlEntry(trustedNfsClients).Execute()
	return err
}

// RemoveTrustedNfsClient removes the trusted NFS client definition from the share
func (c *CteraClient) RemoveTrustedNfsClient(shareName, address, netmask string) error {
	_, err := c.client.SharesApi.CteraGatewayOpenapiApiSharesRemoveTrustedNfsClient(*c.ctx, shareName).Address(address).Netmask(netmask).Execute()
	return err
}

func (*CteraClient) getStatusFromError(err error) int32 {
	genericOpenAPIError, ok := err.(ctera.GenericOpenAPIError)
	if !ok {
		return -1
	}

	errorMessage, ok := genericOpenAPIError.Model().(ctera.ErrorMessage)
	if !ok {
		return -1
	}

	return errorMessage.GetStatus()
}
