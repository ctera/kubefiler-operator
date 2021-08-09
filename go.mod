module github.com/ctera/kubefiler-operator

go 1.16

require (
	github.com/ctera/kubefiler-operator/pkg/ctera-openapi v1.0.0
	github.com/go-logr/logr v0.3.0
	github.com/jinzhu/copier v0.3.2
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	github.com/sethvargo/go-password v0.2.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.6.1
	google.golang.org/grpc v1.31.0
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	sigs.k8s.io/controller-runtime v0.8.3
)

replace github.com/ctera/kubefiler-operator/pkg/ctera-openapi v1.0.0 => ./pkg/ctera-openapi
