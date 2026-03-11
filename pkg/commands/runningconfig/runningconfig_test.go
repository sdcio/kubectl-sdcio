package runningconfig

import (
	"context"
	"errors"
	"testing"

	"github.com/sdcio/kubectl-sdc/pkg/client"
	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type stubDataClient struct {
	connectErr    error
	getIntentErr  error
	output        client.Intent
	connected     bool
	format        client.Format
	datastoreName string
	intentName    string
}

func (s *stubDataClient) Connect(context.Context) error {
	s.connected = true
	return s.connectErr
}

func (s *stubDataClient) GetIntent(_ context.Context, format client.Format, datastoreName, intentName string) (client.Intent, error) {
	s.format = format
	s.datastoreName = datastoreName
	s.intentName = intentName
	if s.getIntentErr != nil {
		return nil, s.getIntentErr
	}
	return s.output, nil
}

func (s *stubDataClient) Close() error {
	return nil
}

type stubIntent struct {
	value string
}

func (s stubIntent) String() string          { return s.value }
func (s stubIntent) GetBlob() []byte         { return nil }
func (s stubIntent) GetProto() *sdcpb.Intent { return nil }
func (s stubIntent) GetType() client.Format  { return client.FormatXPath }

func TestParseFormat(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    client.Format
		wantErr string
	}{
		{name: "json", input: "json", want: client.FormatJSON},
		{name: "uppercase", input: "YAML", want: client.FormatYAML},
		{name: "invalid", input: "bogus", wantErr: `invalid format "bogus", must be one of: json, json-ietf, xml, xpath, yaml`},
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

func TestResolveDataServicePort(t *testing.T) {
	t.Run("named port", func(t *testing.T) {
		svc := &corev1.Service{Spec: corev1.ServiceSpec{Ports: []corev1.ServicePort{{Name: "data-service", TargetPort: intstr.FromInt(12345)}}}}
		port, err := ResolveDataServicePort(svc)
		if err != nil {
			t.Fatalf("ResolveDataServicePort() unexpected error: %v", err)
		}
		if port != 12345 {
			t.Fatalf("port = %d, want 12345", port)
		}
	})

	t.Run("fallback port", func(t *testing.T) {
		svc := &corev1.Service{Spec: corev1.ServiceSpec{Ports: []corev1.ServicePort{{Name: "other", TargetPort: intstr.FromInt(1)}}}}
		port, err := ResolveDataServicePort(svc)
		if err != nil {
			t.Fatalf("ResolveDataServicePort() unexpected error: %v", err)
		}
		if port != defaultDataServicePort {
			t.Fatalf("port = %d, want %d", port, defaultDataServicePort)
		}
	})

	t.Run("no ports", func(t *testing.T) {
		svc := &corev1.Service{}
		_, err := ResolveDataServicePort(svc)
		if err == nil || err.Error() != "data-server service has no ports" {
			t.Fatalf("ResolveDataServicePort() error = %v, want %q", err, "data-server service has no ports")
		}
	})
}

func TestRun(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dataClient := &stubDataClient{output: stubIntent{value: "/system/name: srl1"}}
		output, err := Run(context.Background(), dataClient, "default", "srl1", client.FormatXPath)
		if err != nil {
			t.Fatalf("Run() unexpected error: %v", err)
		}
		if !dataClient.connected {
			t.Fatal("expected data client to connect")
		}
		if dataClient.datastoreName != "default.srl1" {
			t.Fatalf("datastore name = %q, want %q", dataClient.datastoreName, "default.srl1")
		}
		if dataClient.intentName != "running" {
			t.Fatalf("intent name = %q, want %q", dataClient.intentName, "running")
		}
		if output != "/system/name: srl1" {
			t.Fatalf("output = %q, want %q", output, "/system/name: srl1")
		}
	})

	t.Run("connect error", func(t *testing.T) {
		dataClient := &stubDataClient{connectErr: errors.New("boom")}
		_, err := Run(context.Background(), dataClient, "default", "srl1", client.FormatXPath)
		if err == nil || err.Error() != "failed to connect to data-server: boom" {
			t.Fatalf("Run() error = %v, want %q", err, "failed to connect to data-server: boom")
		}
	})

	t.Run("get intent error", func(t *testing.T) {
		dataClient := &stubDataClient{getIntentErr: errors.New("fetch failed")}
		_, err := Run(context.Background(), dataClient, "default", "srl1", client.FormatXPath)
		if err == nil || err.Error() != "fetch failed" {
			t.Fatalf("Run() error = %v, want %q", err, "fetch failed")
		}
	})
}
