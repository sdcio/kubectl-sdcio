package cmd

import (
	"encoding/json"
	"testing"

	"github.com/sdcio/config-server/apis/config/v1alpha1"
	"github.com/sdcio/kubectl-sdc/pkg/types"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"sigs.k8s.io/yaml"
)

func TestDeviationOptionsValidate(t *testing.T) {
	tests := []struct {
		name      string
		target    string
		deviation string
		format    string
		namespace string
		wantErr   string
	}{
		{
			name:      "requires target or deviation",
			namespace: "default",
			wantErr:   "deviation or target not set",
		},
		{
			name:      "accepts deviation only",
			deviation: "dev-1",
			format:    string(deviationOutputFormatText),
			namespace: "default",
		},
		{
			name:      "accepts target only",
			target:    "target-1",
			format:    string(deviationOutputFormatText),
			namespace: "default",
		},
		{
			name:      "accepts both",
			target:    "target-1",
			deviation: "dev-1",
			format:    string(deviationOutputFormatText),
			namespace: "default",
		},
		{
			name:      "requires namespace",
			deviation: "dev-1",
			format:    string(deviationOutputFormatText),
			wantErr:   "namespace not set",
		},
		{
			name:      "rejects invalid format",
			deviation: "dev-1",
			format:    "bogus",
			namespace: "default",
			wantErr:   "invalid format \"bogus\", must be one of: text, resource-yaml, resource-json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &DeviationOptions{
				target:    tt.target,
				deviation: tt.deviation,
				format:    tt.format,
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

func TestDeviationOutputFormat_DefaultAndCompletion(t *testing.T) {
	cmd, err := NewCmdDeviation(genericiooptions.NewTestIOStreamsDiscard())
	if err != nil {
		t.Fatalf("NewCmdDeviation() unexpected error: %v", err)
	}

	flag := cmd.Flags().Lookup("format")
	if flag == nil {
		t.Fatal("format flag not registered")
	}
	if got := flag.DefValue; got != string(deviationOutputFormatText) {
		t.Fatalf("default format = %q, want %q", got, deviationOutputFormatText)
	}

	comps, directive := deviationFormatCompletionFunc()(&cobra.Command{}, nil, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("completion directive = %v, want %v", directive, cobra.ShellCompDirectiveNoFileComp)
	}
	if len(comps) != len(deviationOutputFormats) {
		t.Fatalf("completion count = %d, want %d", len(comps), len(deviationOutputFormats))
	}
}

func TestFormatSelectedDeviations(t *testing.T) {
	devs := types.Deviations{}
	intent := types.NewDeviations("target-1", "dev-1", types.DeviationTypeConfig, 1).SetNamespace("default")
	intent.AddDeviation(types.NewDeviation("/system/name", "router-1", "router-2", "mismatch"))
	devs.AddDeviation(intent)

	t.Run("text", func(t *testing.T) {
		out, err := formatSelectedDeviations(devs, deviationOutputFormatText)
		if err != nil {
			t.Fatalf("formatSelectedDeviations() unexpected error: %v", err)
		}
		if out == "" {
			t.Fatal("text output is empty")
		}
	})

	t.Run("resource yaml", func(t *testing.T) {
		out, err := formatSelectedDeviations(devs, deviationOutputFormatResourceYAML)
		if err != nil {
			t.Fatalf("formatSelectedDeviations() unexpected error: %v", err)
		}

		var resource v1alpha1.TargetClearDeviation
		if err := yaml.Unmarshal([]byte(out), &resource); err != nil {
			t.Fatalf("yaml.Unmarshal() unexpected error: %v", err)
		}
		if resource.Name != "target-1" {
			t.Fatalf("resource name = %q, want %q", resource.Name, "target-1")
		}
		if resource.Namespace != "default" {
			t.Fatalf("resource namespace = %q, want %q", resource.Namespace, "default")
		}
		if len(resource.Spec.Config) != 1 {
			t.Fatalf("config length = %d, want 1", len(resource.Spec.Config))
		}
	})

	t.Run("resource json", func(t *testing.T) {
		out, err := formatSelectedDeviations(devs, deviationOutputFormatResourceJSON)
		if err != nil {
			t.Fatalf("formatSelectedDeviations() unexpected error: %v", err)
		}

		var resource v1alpha1.TargetClearDeviation
		if err := json.Unmarshal([]byte(out), &resource); err != nil {
			t.Fatalf("json.Unmarshal() unexpected error: %v", err)
		}
		if resource.Name != "target-1" {
			t.Fatalf("resource name = %q, want %q", resource.Name, "target-1")
		}
	})

	t.Run("nil selected deviations", func(t *testing.T) {
		out, err := formatSelectedDeviations(nil, deviationOutputFormatText)
		if err != nil {
			t.Fatalf("formatSelectedDeviations() unexpected error: %v", err)
		}
		if out != "" {
			t.Fatalf("output = %q, want empty string", out)
		}
	})
}

func TestNewCmdDeviationRequiresOneOfTargetOrDeviation(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "requires one flag",
			args:    nil,
			wantErr: "at least one of the flags in the group [deviation target] is required",
		},
		{
			name: "accepts target",
			args: []string{"--target=target-1"},
		},
		{
			name: "accepts deviation",
			args: []string{"--deviation=dev-1"},
		},
		{
			name: "accepts both",
			args: []string{"--target=target-1", "--deviation=dev-1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := NewCmdDeviation(genericiooptions.NewTestIOStreamsDiscard())
			if err != nil {
				t.Fatalf("NewCmdDeviation() unexpected error: %v", err)
			}

			if err := cmd.ParseFlags(tt.args); err != nil {
				t.Fatalf("ParseFlags() unexpected error: %v", err)
			}

			err = cmd.ValidateFlagGroups()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("ValidateFlagGroups() unexpected error: %v", err)
				}
				return
			}

			if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("ValidateFlagGroups() error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}
