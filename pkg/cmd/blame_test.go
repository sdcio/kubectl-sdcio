package cmd

import (
	"testing"

	"github.com/sdcio/kubectl-sdc/pkg/commands/blame"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

func TestBlameOptionsValidate(t *testing.T) {
	tests := []struct {
		name      string
		target    string
		namespace string
		wantErr   string
	}{
		{
			name:      "requires target",
			namespace: "default",
			wantErr:   "target not set",
		},
		{
			name:    "requires namespace",
			target:  "target-1",
			wantErr: "namespace not set",
		},
		{
			name:      "accepts target and namespace",
			target:    "target-1",
			namespace: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &BlameOptions{
				target: tt.target,
				GenericOptions: GenericOptions{
					namespace: tt.namespace,
				},
			}

			err := o.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() unexpected error: %v", err)
				}
				return
			}

			if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("Validate() error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestBlameFormat_DefaultAndCompletion(t *testing.T) {
	cmd, err := NewCmdBlame(genericiooptions.NewTestIOStreamsDiscard())
	if err != nil {
		t.Fatalf("NewCmdBlame() unexpected error: %v", err)
	}

	flag := cmd.Flags().Lookup("format")
	if flag == nil {
		t.Fatal("format flag not registered")
	}
	if got := flag.DefValue; got != string(blame.BlameFormatTree) {
		t.Fatalf("default format = %q, want %q", got, blame.BlameFormatTree)
	}

	comps, directive := formatCompletionFunc()(&cobra.Command{}, nil, "")
	if directive != cobra.ShellCompDirectiveDefault {
		t.Fatalf("completion directive = %v, want %v", directive, cobra.ShellCompDirectiveDefault)
	}
	if len(comps) != len(blame.BlameFormatOptions) {
		t.Fatalf("completion count = %d, want %d", len(comps), len(blame.BlameFormatOptions))
	}
}

func TestNewCmdBlameRequiresTarget(t *testing.T) {
	cmd, err := NewCmdBlame(genericiooptions.NewTestIOStreamsDiscard())
	if err != nil {
		t.Fatalf("NewCmdBlame() unexpected error: %v", err)
	}

	if err := cmd.ParseFlags(nil); err != nil {
		t.Fatalf("ParseFlags() unexpected error: %v", err)
	}

	err = cmd.ValidateRequiredFlags()
	if err == nil || err.Error() != `required flag(s) "target" not set` {
		t.Fatalf("ValidateRequiredFlags() error = %v, want %q", err, `required flag(s) "target" not set`)
	}
}
