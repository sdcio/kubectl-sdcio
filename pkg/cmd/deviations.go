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

	"github.com/spf13/cobra"

	"github.com/sdcio/kubectl-sdc/pkg/client"
	"github.com/sdcio/kubectl-sdc/pkg/commands/deviations"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

// DeviationOptions defines raw options for the deviation command as provided by the user via cobra flags
type DeviationOptions struct {
	target       string
	deviation    string
	format       string
	preview      bool
	revert       bool
	initialQuery string
	preselect    string
	GenericOptions
}

// NewDeviationOptions provides an instance of DeviationOptions with default values
func NewDeviationOptions(streams genericiooptions.IOStreams) *DeviationOptions {
	return &DeviationOptions{
		GenericOptions: GenericOptions{
			configFlags: genericclioptions.NewConfigFlags(true),
			IOStreams:   streams,
		},
	}
}

func (o *DeviationOptions) Complete(_ *cobra.Command, _ []string) error {
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
func (o *DeviationOptions) Validate() error {
	if o.deviation == "" && o.target == "" {
		return fmt.Errorf("deviation or target not set")
	}
	if o.namespace == "" {
		return fmt.Errorf("namespace not set")
	}
	if _, err := parseDeviationOutputFormat(o.format); err != nil {
		return err
	}
	return nil
}

func (o *DeviationOptions) Run(_ *cobra.Command) error {
	ctx := context.Background()
	cl, err := client.NewConfigClient(o.restConfig)
	if err != nil {
		return err
	}

	opts := []deviations.DeviationOptionSetter{
		deviations.WithPreview(o.preview),
		deviations.WithRevert(o.revert),
		deviations.WithDeviationName(o.deviation),
		deviations.WithTarget(o.target),
		deviations.WithInitialQuery(o.initialQuery),
		deviations.WithPreSelect(o.preselect),
	}

	// Run the deviation selection
	selectedDeviations, err := deviations.Run(ctx, cl, deviations.NewDeviationOptions(o.namespace, opts...))
	if err != nil {
		return err
	}

	if selectedDeviations == nil || !selectedDeviations.HasDeviations() {
		return nil
	}

	format, err := parseDeviationOutputFormat(o.format)
	if err != nil {
		return err
	}

	output, err := formatSelectedDeviations(selectedDeviations, format)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(o.Out, output)

	return err
}

// NewCmdDeviation provides a cobra command wrapping DeviationOptions
func NewCmdDeviation(streams genericiooptions.IOStreams) (*cobra.Command, error) {

	o := NewDeviationOptions(streams)

	cmd := &cobra.Command{
		Use:          "deviation",
		Short:        "deviations",
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

	cmd.Flags().StringVar(&o.target, "target", "", "target to get the deviations for. This is simply to assist auto-completion of the deviation name")
	cmd.Flags().StringVar(&o.deviation, "deviation", "", "deviation resource name to query")
	cmd.Flags().StringVar(&o.preselect, "preselect", "", "preselect a deviation path for fuzzy finder")
	cmd.Flags().StringVar(&o.format, "format", string(deviationOutputFormatText), fmt.Sprintf("output format (%s)", deviationOutputFormatListString()))
	cmd.Flags().BoolVar(&o.preview, "preview", false, "show preview of deviations")
	cmd.Flags().BoolVar(&o.revert, "revert", false, "revert deviations")
	cmd.Flags().StringVar(&o.initialQuery, "query", "", "initial query for fuzzy finder")
	cmd.MarkFlagsOneRequired("deviation", "target")

	if err := cmd.RegisterFlagCompletionFunc("target", targetCompletionFunc(o)); err != nil {
		return nil, err
	}
	if err := cmd.RegisterFlagCompletionFunc("deviation", deviationCompletionFunc(o)); err != nil {
		return nil, err
	}
	if err := cmd.RegisterFlagCompletionFunc("format", deviationFormatCompletionFunc()); err != nil {
		return nil, err
	}
	o.configFlags.AddFlags(cmd.Flags())

	return cmd, nil
}
