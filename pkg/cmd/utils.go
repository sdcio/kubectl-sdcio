package cmd

import (
	"context"

	"github.com/sdcio/kubectl-sdc/pkg/client"
	"github.com/sdcio/kubectl-sdc/pkg/commands/blame"
	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
)

func compError(err error) ([]string, cobra.ShellCompDirective) {
	cobra.CompError(err.Error())
	return nil, cobra.ShellCompDirectiveError
}

// targetCompletionFunc is a completion function that completes target
// that match the toComplete prefix.
func targetCompletionFunc(o k8sCompletion) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if err := o.Complete(nil, nil); err != nil {
			return compError(err)
		}

		cl, err := client.NewConfigClient(o.RESTConfig())
		if err != nil {
			return compError(err)
		}

		comps, err := cl.ListTargetNames(context.Background(), o.GetNamespace())
		if err != nil {
			return compError(err)
		}

		return comps, cobra.ShellCompDirectiveNoFileComp
	}
}

// deviationCompletionFunc is a completion function that completes deviations
// that match the toComplete prefix.
func deviationCompletionFunc(o *DeviationOptions) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if err := o.Complete(nil, nil); err != nil {
			return compError(err)
		}

		cl, err := client.NewConfigClient(o.RESTConfig())
		if err != nil {
			return compError(err)
		}

		var selectLabels map[string]string
		if o.target != "" {
			selectLabels = map[string]string{"config.sdcio.dev/targetName": o.target}
		}

		comps, err := cl.ListDeviationNames(context.Background(), o.GetNamespace(), selectLabels)
		if err != nil {
			return compError(err)
		}

		return comps, cobra.ShellCompDirectiveNoFileComp
	}
}

// formatCompletionFunc is a completion function that completes format options
func formatCompletionFunc() func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var formats []string
		for _, f := range blame.BlameFormatOptions {
			formats = append(formats, string(f))
		}
		return formats, cobra.ShellCompDirectiveDefault
	}
}

type k8sCompletion interface {
	RESTConfig() *rest.Config
	Complete(*cobra.Command, []string) error
	GetNamespace() string
}
