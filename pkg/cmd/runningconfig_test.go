package cmd

import (
	"testing"

	"github.com/sdcio/kubectl-sdc/pkg/client"
	"github.com/sdcio/kubectl-sdc/pkg/commands/runningconfig"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

func TestRunningConfigOptionsValidate(t *testing.T) {
	tests := []struct {
		name      string
		target    string
		formatStr string
		namespace string
		want      client.Format
		wantErr   string
	}{
		{name: "requires target", namespace: "default", wantErr: "target not set"},
		{name: "requires namespace", target: "srl1", wantErr: "namespace not set"},
		{name: "invalid format", target: "srl1", namespace: "default", formatStr: "bogus", wantErr: `invalid format "bogus", must be one of: json, json-ietf, xml, xpath, yaml`},
		{name: "valid format", target: "srl1", namespace: "default", formatStr: "yaml", want: client.FormatYAML},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &RunningConfigOptions{
				target:    tt.target,
				formatStr: tt.formatStr,
				GenericOptions: GenericOptions{
					namespace: tt.namespace,
				},
			}

			err := o.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() unexpected error: %v", err)
				}
				if o.format != tt.want {
					t.Fatalf("format = %q, want %q", o.format, tt.want)
				}
				return
			}

			if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("Validate() error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestNewCmdRunningConfig_DefaultFormat(t *testing.T) {
	cmd, err := NewCmdRunningConfig(genericiooptions.NewTestIOStreamsDiscard())
	if err != nil {
		t.Fatalf("NewCmdRunningConfig() unexpected error: %v", err)
	}

	flag := cmd.Flags().Lookup("format")
	if flag == nil {
		t.Fatal("format flag not registered")
	}
	if got := flag.DefValue; got != string(client.FormatXPath) {
		t.Fatalf("default format = %q, want %q", got, client.FormatXPath)
	}
}

func TestRunningConfigFormatStrings(t *testing.T) {
	formats := runningconfig.ValidFormatStrings()
	if len(formats) != len(runningconfig.ValidFormats) {
		t.Fatalf("format count = %d, want %d", len(formats), len(runningconfig.ValidFormats))
	}
}

func TestNewCmdRunningConfigRequiresTarget(t *testing.T) {
	cmd, err := NewCmdRunningConfig(genericiooptions.NewTestIOStreamsDiscard())
	if err != nil {
		t.Fatalf("NewCmdRunningConfig() unexpected error: %v", err)
	}

	if err := cmd.ParseFlags(nil); err != nil {
		t.Fatalf("ParseFlags() unexpected error: %v", err)
	}

	err = cmd.ValidateRequiredFlags()
	if err == nil || err.Error() != `required flag(s) "target" not set` {
		t.Fatalf("ValidateRequiredFlags() error = %v, want %q", err, `required flag(s) "target" not set`)
	}
}
