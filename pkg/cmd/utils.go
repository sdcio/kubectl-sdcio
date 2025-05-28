package cmd

import (
	"context"

	"github.com/sdcio/kubectl-sdcio/pkg/client"
	"github.com/spf13/cobra"
)

func compError(err error) ([]string, cobra.ShellCompDirective) {
	cobra.CompError(err.Error())
	return nil, cobra.ShellCompDirectiveError
}

// targetCompletionFunc is a completion function that completes target
// that match the toComplete prefix.
func targetCompletionFunc(o *BlameOptions) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if err := o.Complete(nil, nil); err != nil {
			return compError(err)
		}

		cl, err := client.NewConfigClient(o.restConfig)
		if err != nil {
			return compError(err)
		}

		comps, err := cl.GetTargetNames(context.Background(), o.namespace)
		if err != nil {
			return compError(err)
		}

		return comps, cobra.ShellCompDirectiveNoFileComp
	}
}
