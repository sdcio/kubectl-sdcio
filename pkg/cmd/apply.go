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
	"github.com/sdcio/kubectl-sdc/pkg/commands/apply"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

type ApplyOptions struct {
	files []string
	GenericOptions
}

// NewApplyOptions provides an instance of ApplyOptions with default values
func NewApplyOptions(streams genericiooptions.IOStreams) *ApplyOptions {
	return &ApplyOptions{
		GenericOptions: GenericOptions{
			configFlags: genericclioptions.NewConfigFlags(true),
			IOStreams:   streams,
		},
	}
}

func (o *ApplyOptions) Complete(_ *cobra.Command, args []string) error {
	var err error
	clientConfig := o.configFlags.ToRawKubeConfigLoader()

	o.restConfig, err = o.configFlags.ToRESTConfig()
	if err != nil {
		return err
	}

	o.namespace, _, err = clientConfig.Namespace()
	if err != nil {
		return err
	}

	// Remaining positional args are treated as file paths (alongside --filename).
	o.files = append(o.files, args...)
	return nil
}

func (o *ApplyOptions) Validate() error {
	if len(o.files) == 0 {
		return fmt.Errorf("must provide at least one filename (use -f or pass paths as arguments)")
	}
	return nil
}

func (o *ApplyOptions) Run(_ *cobra.Command) error {
	ctx := context.Background()

	cl, err := client.NewConfigClient(o.restConfig)
	if err != nil {
		return err
	}

	return apply.Apply(ctx, cl, o.namespace, o.files, o.Out)
}

// NewCmdApply provides a cobra command wrapping ApplyOptions
func NewCmdApply(streams genericiooptions.IOStreams) (*cobra.Command, error) {
	o := NewApplyOptions(streams)

	cmd := &cobra.Command{
		Use:          "apply",
		Short:        "Apply a resource from a file or stdin",
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(c, args); err != nil {
				return err
			}
			if err := o.Validate(); err != nil {
				return err
			}
			return o.Run(c)
		},
	}

	cmd.Flags().StringSliceVarP(&o.files, "filename", "f", nil, "filename, directory, or URL to files to apply")
	o.configFlags.AddFlags(cmd.Flags())

	return cmd, nil
}
