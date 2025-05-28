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

	"github.com/sdcio/kubectl-sdcio/pkg/client"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

type BlameOptions struct {
	namespace string
	target    string
	MyOptions
}

// NewBlameOptions provides an instance of NamespaceOptions with default values
func NewBlameOptions(streams genericiooptions.IOStreams) *BlameOptions {
	return &BlameOptions{
		MyOptions: MyOptions{
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

func (o *BlameOptions) Run(cmd *cobra.Command) error {
	ctx := context.Background()
	cl, err := client.NewConfigClient(o.restConfig)
	if err != nil {
		return err
	}

	bt, err := cl.GetBlameTree(ctx, o.namespace, o.target)
	if err != nil {
		return err
	}

	fmt.Println(bt.ToString())
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
	cmd.MarkFlagRequired("target")

	if err := cmd.RegisterFlagCompletionFunc("target", targetCompletionFunc(o)); err != nil {
		return nil, err
	}
	o.configFlags.AddFlags(cmd.Flags())

	return cmd, nil
}
