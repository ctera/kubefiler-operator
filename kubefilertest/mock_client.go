package kubefilertest

import (
	"context"

	kubefilerv1alpha1 "github.com/ctera/kubefiler-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/jinzhu/copier"

	"github.com/stretchr/testify/mock"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(kubefilerv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

type MockedClient struct {
	mock.Mock
	client.Client

	ReturnObject client.Object
}

func (m *MockedClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object) error {
	args := m.Called(ctx, key, obj)
	copier.Copy(m.ReturnObject, obj)
	return args.Error(0)
}

func (m *MockedClient) Scheme() *runtime.Scheme {
	return scheme
}

func (m *MockedClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	args := m.Called(ctx, obj)
	return args.Error(0)
}
