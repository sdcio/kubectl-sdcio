package cmd

import (
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/client-go/rest"
)

type GenericOptions struct {
	restConfig  *rest.Config
	configFlags *genericclioptions.ConfigFlags
	genericiooptions.IOStreams
}
