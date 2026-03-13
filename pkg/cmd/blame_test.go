package cmd

import (
	"errors"
	"testing"

	fuzzyfinder "github.com/ktr0731/go-fuzzyfinder"
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

	interactiveFlag := cmd.Flags().Lookup("interactive")
	if interactiveFlag == nil {
		t.Fatal("interactive flag not registered")
	}
	if got := interactiveFlag.DefValue; got != "false" {
		t.Fatalf("default interactive = %q, want %q", got, "false")
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

func TestSelectXPathResult(t *testing.T) {
	tests := []struct {
		name        string
		lines       []string
		interactive bool
		selector    func(lines []string, opts ...fuzzyfinder.Option) ([]int, error)
		want        string
		wantErr     bool
	}{
		{
			name:        "non interactive returns full output",
			lines:       []string{"/a", "/b"},
			interactive: false,
			want:        "/a\n/b",
		},
		{
			name:        "interactive returns selected lines",
			lines:       []string{"/a", "/b", "/c"},
			interactive: true,
			selector: func(lines []string, opts ...fuzzyfinder.Option) ([]int, error) {
				if len(lines) != 3 {
					t.Fatalf("expected 3 lines, got %d", len(lines))
				}
				return []int{0, 2}, nil
			},
			want: "/a\n/c",
		},
		{
			name:        "interactive no selection returns empty",
			lines:       []string{"/a", "/b"},
			interactive: true,
			selector: func(_ []string, _ ...fuzzyfinder.Option) ([]int, error) {
				return []int{}, nil
			},
			want: "",
		},
		{
			name:        "interactive abort returns empty",
			lines:       []string{"/a", "/b"},
			interactive: true,
			selector: func(_ []string, _ ...fuzzyfinder.Option) ([]int, error) {
				return nil, fuzzyfinder.ErrAbort
			},
			want: "",
		},
		{
			name:        "interactive selector error bubbles up",
			lines:       []string{"/a", "/b"},
			interactive: true,
			selector: func(_ []string, _ ...fuzzyfinder.Option) ([]int, error) {
				return nil, errors.New("boom")
			},
			wantErr: true,
		},
		{
			name:        "interactive invalid selected index returns error",
			lines:       []string{"/a", "/b"},
			interactive: true,
			selector: func(_ []string, _ ...fuzzyfinder.Option) ([]int, error) {
				return []int{4}, nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orig := findBlameXPathIndexes
			defer func() { findBlameXPathIndexes = orig }()

			if tt.selector != nil {
				findBlameXPathIndexes = tt.selector
			}

			got, err := selectXPathResult(tt.lines, tt.interactive)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.want {
				t.Fatalf("result = %q, want %q", got, tt.want)
			}
		})
	}
}
