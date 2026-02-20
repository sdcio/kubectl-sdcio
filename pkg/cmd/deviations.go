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
	"unicode"
	"unicode/utf8"

	fuzzyfinder "github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"

	"github.com/sdcio/kubectl-sdcio/pkg/client"
	"github.com/sdcio/kubectl-sdcio/pkg/types"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

func reasonInitial(reason string) string {
	if reason == "" {
		return "[?]"
	}
	r, _ := utf8.DecodeRuneInString(reason)
	if r == utf8.RuneError {
		return "[?]"
	}
	return fmt.Sprintf("[%s]", string(unicode.ToUpper(r)))
}

func alignLabel(label string, width int) string {
	if width <= len(label) {
		return label
	}
	return label + strings.Repeat(" ", width-len(label))
}

type DeviationOptions struct {
	namespace string
	deviation string
	preview   bool
	MyOptions
}

// NewDeviationOptions provides an instance of DeviationOptions with default values
func NewDeviationOptions(streams genericiooptions.IOStreams) *DeviationOptions {
	return &DeviationOptions{
		MyOptions: MyOptions{
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
	if o.deviation == "" {
		return fmt.Errorf("deviation not set")
	}
	if o.namespace == "" {
		return fmt.Errorf("namespace not set")
	}
	return nil
}

func addPreviewOpt(opts []fuzzyfinder.Option, deviations []types.Deviation) []fuzzyfinder.Option {
	opts = append(opts, fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
		if i == -1 {
			return ""
		}
		labels := []string{"Path:", "Actual:", "Desired:", "Reason:"}
		maxLabel := 0
		for _, label := range labels {
			if len(label) > maxLabel {
				maxLabel = len(label)
			}
		}
		preview := fmt.Sprintf(
			"%s %s\n%s %s\n%s %s\n%s %s\n",
			alignLabel("Path:", maxLabel), deviations[i].Path(),
			alignLabel("Actual:", maxLabel), deviations[i].ActualValue(),
			alignLabel("Desired:", maxLabel), deviations[i].DesiredValue(),
			alignLabel("Reason:", maxLabel), deviations[i].Reason(),
		)
		return preview
	}))
	return opts
}

func (o *DeviationOptions) Run(_ *cobra.Command) error {
	ctx := context.Background()
	cl, err := client.NewConfigClient(o.restConfig)
	if err != nil {
		return err
	}

	dev, err := cl.GetDeviations(ctx, o.namespace, o.deviation)
	if err != nil {
		return err
	}

	if dev.Length() == 0 {
		fmt.Fprintln(o.IOStreams.Out, "No deviations found")
		return nil
	}

	opts := []fuzzyfinder.Option{
		fuzzyfinder.WithHeader(fmt.Sprintf("Namespace: %s, Deviation: %s [%s]", dev.Namespace(), dev.Name(), dev.Type())),
	}

	// add preview as an option if the flag is set
	if o.preview {
		opts = addPreviewOpt(opts, dev.Deviations())
	}

	deviations := dev.Deviations()
	// Use fuzzy finder with multi-select to choose deviations to display
	idxs, err := fuzzyfinder.FindMulti(
		deviations,
		func(i int) string {
			return fmt.Sprintf("%s %s", reasonInitial(deviations[i].Reason()), deviations[i].Path())
		},
		opts...,
	)
	if err != nil {
		return err
	}

	// Display selected deviations
	for _, idx := range idxs {
		fmt.Fprintln(o.IOStreams.Out, deviations[idx].Path())
	}

	return nil
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

	cmd.Flags().StringVar(&o.deviation, "deviation", "", "deviation to get the blame config for")
	cmd.Flags().BoolVar(&o.preview, "preview", false, "show preview of deviations")
	err := cmd.MarkFlagRequired("deviation")
	if err != nil {
		return nil, err
	}

	if err := cmd.RegisterFlagCompletionFunc("deviation", deviationCompletionFunc(o)); err != nil {
		return nil, err
	}
	o.configFlags.AddFlags(cmd.Flags())

	return cmd, nil
}
