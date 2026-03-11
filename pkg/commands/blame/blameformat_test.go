package blame

import "testing"

func TestParseFormat(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    BlameFormat
		wantErr string
	}{
		{
			name:  "tree",
			input: "tree",
			want:  BlameFormatTree,
		},
		{
			name:  "xpath uppercase",
			input: "XPATH",
			want:  BlameFormatXPath,
		},
		{
			name:    "invalid",
			input:   "bogus",
			wantErr: `invalid format: "bogus" (must be one of [tree xpath])`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFormat(tt.input)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("ParseFormat() unexpected error: %v", err)
				}
				if got != tt.want {
					t.Fatalf("ParseFormat() = %q, want %q", got, tt.want)
				}
				return
			}

			if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("ParseFormat() error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}
