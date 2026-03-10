/*
Copyright 2018 The Kubernetes Authors.

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

package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/sdcio/kubectl-sdc/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/client-go/kubernetes"
)

// ValidFormats lists all supported output formats
var ValidFormats = []client.Format{
	client.FormatJSON,
	client.FormatJSONIETF,
	client.FormatXML,
	client.FormatXPath,
	client.FormatYAML,
}

// formatListString returns a comma-separated string of valid formats
func formatListString() string {
	return strings.Join(validFormatStrings(), ", ")
}

// validFormatStrings returns the list of valid format strings
func validFormatStrings() []string {
	formatted := make([]string, len(ValidFormats))
	for i, f := range ValidFormats {
		formatted[i] = string(f)
	}
	return formatted
}

// ParseFormat converts a format string to internal Format enum
func ParseFormat(formatStr string) (client.Format, error) {
	switch client.Format(strings.ToLower(formatStr)) {
	case client.FormatJSON:
		return client.FormatJSON, nil
	case client.FormatJSONIETF:
		return client.FormatJSONIETF, nil
	case client.FormatXML:
		return client.FormatXML, nil
	case client.FormatXPath:
		return client.FormatXPath, nil
	case client.FormatYAML:
		return client.FormatYAML, nil
	default:
		return "", fmt.Errorf("invalid format %q, must be one of: %s", formatStr, formatListString())
	}
}

type RunningConfigOptions struct {
	target    string
	formatStr string
	format    client.Format
	GenericOptions
}

// NewRunningConfigOptions provides an instance of RunningConfigOptions with default values
func NewRunningConfigOptions(streams genericiooptions.IOStreams) *RunningConfigOptions {
	return &RunningConfigOptions{
		GenericOptions: GenericOptions{
			configFlags: genericclioptions.NewConfigFlags(true),
			IOStreams:   streams,
		},
	}
}

func (o *RunningConfigOptions) Complete(_ *cobra.Command, _ []string) error {
	var err error
	clientConfig := o.configFlags.ToRawKubeConfigLoader()

	o.restConfig, err = o.configFlags.ToRESTConfig()
	if err != nil {
		return err
	}

	// retrieve the actual namespace from clientConfig
	o.namespace, _, err = clientConfig.Namespace()
	if err != nil {
		return err
	}

	return nil
}

// Validate validates the options
func (o *RunningConfigOptions) Validate() error {
	if o.target == "" {
		return fmt.Errorf("target not set")
	}
	if o.namespace == "" {
		return fmt.Errorf("namespace not set")
	}
	// Parse format string
	format, err := ParseFormat(o.formatStr)
	if err != nil {
		return err
	}
	o.format = format
	return nil
}

func (o *RunningConfigOptions) Run(_ *cobra.Command) error {
	ctx := context.Background()

	// Get the data-server service port from Kubernetes
	clientset, err := kubernetes.NewForConfig(o.restConfig)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}

	svc, err := clientset.CoreV1().Services("sdc-system").Get(ctx, "data-server", metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get data-server service: %w", err)
	}

	if len(svc.Spec.Ports) == 0 {
		return fmt.Errorf("data-server service has no ports")
	}

	// Find the port named "data-service"
	var port int
	for _, p := range svc.Spec.Ports {
		if p.Name == "data-service" {
			port = int(p.TargetPort.IntVal)
			break
		}
	}

	if port == 0 {
		// Default to common port if not found
		port = 56000
	}

	// Create data client to fetch running config from data-server
	dataClient, err := client.NewDataClient(o.restConfig, "sdc-system", "data-server", port)
	if err != nil {
		return fmt.Errorf("failed to create data client: %w", err)
	}
	defer func() {
		if err := dataClient.Close(); err != nil {
			_, _ = fmt.Fprintf(o.ErrOut, "warning: failed to close data client: %v\n", err)
		}
	}()

	// Connect to data-server (establishes port-forward)
	if err := dataClient.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to data-server: %w", err)
	}

	// Construct datastore name as namespace.target
	datastoreName := fmt.Sprintf("%s.%s", o.namespace, o.target)

	// Fetch configuration in the requested format
	configOutput, err := dataClient.GetIntent(ctx, o.format, datastoreName, "running")
	if err != nil {
		return err
	}

	// Display the formatted output
	fmt.Println(configOutput.String())
	return nil
}

// NewCmdRunningConfig provides a cobra command wrapping RunningConfigOptions
func NewCmdRunningConfig(streams genericiooptions.IOStreams) (*cobra.Command, error) {

	o := NewRunningConfigOptions(streams)
	cmd := &cobra.Command{
		Use:          "runningconfig",
		Short:        "Get running configuration for a target",
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {

			if err := o.Complete(c, args); err != nil {
				return err
			}
			if err := o.Validate(); err != nil {
				return err
			}
			if err := o.Run(c); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&o.target, "target", "", "target to get the running config for")
	err := cmd.MarkFlagRequired("target")
	if err != nil {
		return nil, err
	}

	// Build format help text dynamically
	formatHelp := fmt.Sprintf("output format (%s)", formatListString())
	cmd.Flags().StringVar(&o.formatStr, "format", "xpath", formatHelp)

	// Format flag completion
	if err := cmd.RegisterFlagCompletionFunc("format", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return validFormatStrings(), cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		return nil, err
	}

	if err := cmd.RegisterFlagCompletionFunc("target", targetCompletionFunc(o)); err != nil {
		return nil, err
	}
	o.configFlags.AddFlags(cmd.Flags())

	return cmd, nil
}
