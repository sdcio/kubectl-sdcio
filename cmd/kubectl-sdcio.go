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

package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	sdcioCmd "github.com/sdcio/kubectl-sdcio/pkg/cmd"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

func main() {
	flags := pflag.NewFlagSet("", pflag.ExitOnError)
	pflag.CommandLine = flags

	root := &cobra.Command{
		Use: "sdcio",
		Annotations: map[string]string{
			cobra.CommandDisplayNameAnnotation: "kubectl sdcio",
		},
	}

	blameCmd, err := sdcioCmd.NewCmdBlame(genericiooptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
	if err != nil {
		panic(err)
	}
	root.AddCommand(blameCmd)
	root.AddCommand(completionCmd)
	root.Version = "v0.0.0"
	root.CompletionOptions.DisableDefaultCmd = false

	cobra.EnableCommandSorting = false

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}

}

var err error
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script",
	Long: `To load completions:

Bash:

$ source <(kubectl sdcio completion bash)

Zsh:

$ source <(kubectl sdcio completion zsh)

Fish:

$ kubectl sdcio completion fish | source
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},

	Args: cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			err = cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			err = cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			err = cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			err = cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
		return err
	},
}
