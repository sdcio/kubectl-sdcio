package apply

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sdcio/config-server/apis/config/v1alpha1"
	mockapply "github.com/sdcio/kubectl-sdc/mocks/apply"
	"go.uber.org/mock/gomock"
)

func writeManifest(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	return path
}

func TestApply_SingleResourceWithNamespaceFromManifest(t *testing.T) {
	t.Parallel()

	manifest := `apiVersion: config.sdcio.dev/v1alpha1
kind: TargetClearDeviation
metadata:
  name: srl1
  namespace: from-manifest
spec:
  config:
    - name: intent-a
      paths:
        - /system/name
`

	ctrl := gomock.NewController(t)
	cl := mockapply.NewMockApplyClient(ctrl)

	var got *v1alpha1.TargetClearDeviation
	cl.EXPECT().
		ClearTargetDeviations(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, resource *v1alpha1.TargetClearDeviation) error {
			got = resource
			return nil
		})

	out := &bytes.Buffer{}
	err := Apply(context.Background(), cl, "from-cli", []string{writeManifest(t, manifest)}, out)
	if err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}

	if got == nil {
		t.Fatal("expected resource to be passed to client")
	}
	if got.Namespace != "from-manifest" {
		t.Fatalf("expected namespace from-manifest, got %q", got.Namespace)
	}
	if s := strings.TrimSpace(out.String()); s != "targetcleardeviation/srl1 applied" {
		t.Fatalf("unexpected output: %q", s)
	}
}

func TestApply_FallsBackToCLINamespace(t *testing.T) {
	t.Parallel()

	manifest := `apiVersion: config.sdcio.dev/v1alpha1
kind: TargetClearDeviation
metadata:
  name: srl1
spec:
  config:
    - name: intent-a
      paths:
        - /system/name
`

	ctrl := gomock.NewController(t)
	cl := mockapply.NewMockApplyClient(ctrl)

	var got *v1alpha1.TargetClearDeviation
	cl.EXPECT().
		ClearTargetDeviations(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, resource *v1alpha1.TargetClearDeviation) error {
			got = resource
			return nil
		})

	out := &bytes.Buffer{}
	err := Apply(context.Background(), cl, "from-cli", []string{writeManifest(t, manifest)}, out)
	if err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}

	if got == nil {
		t.Fatal("expected resource to be passed to client")
	}
	if got.Namespace != "from-cli" {
		t.Fatalf("expected namespace from-cli, got %q", got.Namespace)
	}
}

func TestApply_MultiDocumentYAML(t *testing.T) {
	t.Parallel()

	manifest := `apiVersion: config.sdcio.dev/v1alpha1
kind: TargetClearDeviation
metadata:
  name: srl1
spec:
  config:
    - name: intent-a
      paths:
        - /system/name
---
apiVersion: config.sdcio.dev/v1alpha1
kind: TargetClearDeviation
metadata:
  name: srl2
spec:
  config:
    - name: intent-b
      paths:
        - /system/clock/timezone
`

	ctrl := gomock.NewController(t)
	cl := mockapply.NewMockApplyClient(ctrl)
	cl.EXPECT().ClearTargetDeviations(gomock.Any(), gomock.Any()).Return(nil).Times(2)

	out := &bytes.Buffer{}
	err := Apply(context.Background(), cl, "default", []string{writeManifest(t, manifest)}, out)
	if err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}
}

func TestApply_UnsupportedKind(t *testing.T) {
	t.Parallel()

	manifest := `apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
`

	ctrl := gomock.NewController(t)
	cl := mockapply.NewMockApplyClient(ctrl)

	out := &bytes.Buffer{}
	err := Apply(context.Background(), cl, "default", []string{writeManifest(t, manifest)}, out)
	if err == nil {
		t.Fatal("expected error for unsupported kind, got nil")
	}
	if !strings.Contains(err.Error(), `unsupported kind "ConfigMap"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApply_ClientErrorPropagates(t *testing.T) {
	t.Parallel()

	manifest := `apiVersion: config.sdcio.dev/v1alpha1
kind: TargetClearDeviation
metadata:
  name: srl1
spec:
  config:
    - name: intent-a
      paths:
        - /system/name
`

	ctrl := gomock.NewController(t)
	cl := mockapply.NewMockApplyClient(ctrl)
	wantErr := errors.New("backend failed")
	cl.EXPECT().ClearTargetDeviations(gomock.Any(), gomock.Any()).Return(wantErr)

	out := &bytes.Buffer{}
	err := Apply(context.Background(), cl, "default", []string{writeManifest(t, manifest)}, out)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), wantErr.Error()) {
		t.Fatalf("expected propagated client error, got %v", err)
	}
}
