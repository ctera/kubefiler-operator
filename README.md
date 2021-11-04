# Kubefiler Operator
Kubernetes Operator for KubeFiler

## Operator SDK
The kubefiler-operator is based on the Operator SDK [Operator SDK](https://sdk.operatorframework.io/).
See the [Golang](https://sdk.operatorframework.io/docs/building-operators/golang/) documentation

## Build and Deply

### Makefile
See all make options by running `make help`

### Building the Executable
If you wish to only build the executable run `make` or `make build`

### Building the Image
The image build process uses [multi-stage](https://docs.docker.com/develop/develop-images/multistage-build/#use-multi-stage-builds) build. Run `make docker-build`

:information_source: The image name is defines by the Make variable IMG (see Makefile)

### Pushing the Image
Run `make docker-push` to push the image to the remote container registry

:warning: - Update image name to actual repository as it is currently pointing to a private temporary repository

### Deploy and Undeploy the Operator
Once you have your Kubernetes Configuration file in place (~/.kube/config), run `make deploy` or `make undeploy` to deploy or undeploy the operator respectfully