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
	"time"

	"github.com/spf13/cobra"

	"github.com/sdcio/kubectl-sdcio/pkg/client"
	"github.com/sdcio/kubectl-sdcio/pkg/commands/blame"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

type BlameOptions struct {
	namespace       string
	target          string
	format          string
	filterLeaf      []string
	filterOwner     []string
	filterPath      []string
	filterDeviation bool
	GenericOptions
}

// NewBlameOptions provides an instance of NamespaceOptions with default values
func NewBlameOptions(streams genericiooptions.IOStreams) *BlameOptions {
	return &BlameOptions{
		GenericOptions: GenericOptions{
			configFlags: genericclioptions.NewConfigFlags(true),
			IOStreams:   streams,
		},
	}
}

func (o *BlameOptions) Complete(_ *cobra.Command, _ []string) error {
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
func (o *BlameOptions) Validate() error {
	if o.target == "" {
		return fmt.Errorf("target not set")
	}
	if o.namespace == "" {
		return fmt.Errorf("namespace not set")
	}
	return nil
}

func (o *BlameOptions) Run(_ *cobra.Command) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cl, err := client.NewConfigClient(o.restConfig)
	if err != nil {
		return fmt.Errorf("failed to create config client: %w", err)
	}

	// Parse the output format
	format, err := blame.ParseFormat(o.format)
	if err != nil {
		return fmt.Errorf("failed to parse format: %w", err)
	}

	// setup the content filter
	filter := blame.BuildFilters(o.filterLeaf, o.filterOwner, o.filterDeviation)

	// setup the path filter
	pathFilter, err := blame.BuildPathFilters(o.filterPath)
	if err != nil {
		return fmt.Errorf("failed to build path filters: %w", err)
	}

	// run the blame command with the filter
	out, err := blame.Run(ctx, cl, o.namespace, o.target, pathFilter, filter)
	if err != nil {
		return fmt.Errorf("failed to run blame: %w", err)
	}

	if out == nil {
		return fmt.Errorf("blame returned no output")
	}

	// generate the output based on the format
	var result string
	switch format {
	case blame.BlameFormatTree:
		result = out.ToString()
	case blame.BlameFormatXPath:
		result = out.StringXPath()
	default:
		return fmt.Errorf("unknown blame format: %v", format)
	}

	fmt.Fprintln(o.Out, result)
	return nil
}

// NewCmdBlame provides a cobra command wrapping BlameOptions
func NewCmdBlame(streams genericiooptions.IOStreams) (*cobra.Command, error) {

	o := NewBlameOptions(streams)

	cmd := &cobra.Command{
		Use:          "blame",
		Short:        "config blame",
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

	cmd.Flags().StringVar(&o.target, "target", "", "target to get the blame config for")
	err := cmd.MarkFlagRequired("target")

	// filter flags
	cmd.Flags().StringSliceVar(&o.filterLeaf, "filter-leaf", nil, "filter by leaf name (supports wildcards, can be specified multiple times)")
	cmd.Flags().StringSliceVar(&o.filterOwner, "filter-owner", nil, "filter by owner name (supports wildcards, can be specified multiple times)")
	cmd.Flags().StringSliceVar(&o.filterPath, "filter-path", nil, "filter by full path (supports wildcards, can be specified multiple times)")
	cmd.Flags().StringVar(&o.format, "format", "tree", fmt.Sprintf("output format (%s)", blame.FormatOptionsString()))
	cmd.Flags().BoolVar(&o.filterDeviation, "filter-deviation", false, "filter deviations only")

	if err != nil {
		return nil, err
	}

	if err := cmd.RegisterFlagCompletionFunc("target", targetCompletionFunc(o)); err != nil {
		return nil, err
	}
	if err := cmd.RegisterFlagCompletionFunc("format", formatCompletionFunc()); err != nil {
		return nil, err
	}

	o.configFlags.AddFlags(cmd.Flags())

	return cmd, nil
}
