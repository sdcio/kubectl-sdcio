package cmd

import (
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/client-go/rest"
)

// GenericOptions holds common options for all commands and is embedded in specific command options structs
type GenericOptions struct {
	restConfig  *rest.Config
	configFlags *genericclioptions.ConfigFlags
	namespace   string
	genericiooptions.IOStreams
}

func (o *GenericOptions) RESTConfig() *rest.Config {
	return o.restConfig
}

func (o *GenericOptions) GetNamespace() string {
	return o.namespace
}
