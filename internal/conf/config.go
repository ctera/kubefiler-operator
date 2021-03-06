package conf

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// OperatorConfig is a type holding general configuration values.
// Most of the operator code that needs to reference configuration
// should do so via this type.
type OperatorConfig struct {
	// GatewayContainerImage can be used to select alternate container sources.
	GatewayContainerImage string `mapstructure:"gateway-container-image"`
	// GatewayContainerName can be used to set the name of the primary container,
	// the one running gateway, in the pod.
	GatewayContainerName string `mapstructure:"gateway-container-name"`
	// GatewayStorageMountPath is where the storage volume should be mounted to
	GatewayStorageMountPath string `mapstructure:"gateway-storage-path"`
	// GatewayOpenAPIContainerImage can be used to select alternate container sources.
	GatewayOpenAPIContainerImage string `mapstructure:"gateway-openapi-container-image"`
	// GatewayContainerName can be used to set the name of the primary container,
	// the one running gateway, in the pod.
	GatewayOpenAPIContainerName string `mapstructure:"gateway-openapi-container-name"`
	// WorkingNamespace defines the namespace for the operator's internal resources
	WorkingNamespace string `mapstructure:"working-namespace"`
	// LockerConfigMapName defines the name of the ConfigMap used as the backend for the distributed locking
	LockerConfigMapName string `mapstructure:"locker-configmap-name"`
}

// Validate the OperatorConfig returning an error if the config is not
// directly usable by the operator. This may occur if certain required
// values are unset or invalid.
func (*OperatorConfig) Validate() error {
	// Ensure that WorkingNamespace is set. We don't default it to anything.
	// It must be passed in, typically by the operator's own pod spec.
	// if oc.WorkingNamespace == "" {
	// 	return fmt.Errorf("WorkingNamespace value [%s] invalid", oc.WorkingNamespace)
	// }
	return nil
}

// Source is how external configuration sources populate the operator config.
type Source struct {
	v    *viper.Viper
	fset *pflag.FlagSet
}

// NewSource creates a new Source based on default configuration values.
func NewSource() *Source {
	v := viper.New()
	v.SetDefault("gateway-container-image", "192.168.9.174:5000/kubefiler-filer:last_build")
	v.SetDefault("gateway-container-name", "kubefiler-filer")
	v.SetDefault("gateway-storage-path", "/var/vol/2")
	v.SetDefault("gateway-openapi-container-image", "192.168.9.174:5000/gateway-openapi:last_build")
	v.SetDefault("gateway-openapi-container-name", "gateway-openapi")
	v.SetDefault("working-namespace", "kubefiler-operator-system")
	v.SetDefault("locker-configmap-name", "kubefiler-operator-locker")
	return &Source{v: v}
}

// Flags returns a pflag FlagSet populated with flags based on the default
// configuration. If used, flags allow changing configuration values on
// the CLI.
// Once parsed these flags act as a configuration source.
func (s *Source) Flags() *pflag.FlagSet {
	if s.fset != nil {
		return s.fset
	}
	s.fset = pflag.NewFlagSet("conf", pflag.ExitOnError)
	for _, k := range s.v.AllKeys() {
		s.fset.String(k, "",
			fmt.Sprintf("Specify the %q configuration parameter", k))
	}
	return s.fset
}

// Read a new OperatorConfig from all available sources.
func (s *Source) Read() (*OperatorConfig, error) {
	v := s.v

	// we look in /etc/kube-filer-operator and the working dir for
	// yaml/toml/etc config files (none are required)
	v.AddConfigPath("/etc/kube-filer-operator")
	v.AddConfigPath(".")
	v.SetConfigName("kube-filer-operator")
	err := v.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	// we automatically pull from the environment
	v.SetEnvPrefix("KUBE_FILER_OP")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()

	// use cli flags if available
	if s.fset != nil {
		v.BindPFlags(s.fset)
	}

	// we isolate config handling to this package. thus we marshal
	// our config to the public OperatorConfig type and return that.
	c := &OperatorConfig{}
	if err := v.Unmarshal(c); err != nil {
		return nil, err
	}
	return c, nil
}
