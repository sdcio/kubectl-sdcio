package apply

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/sdcio/config-server/apis/config/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
)

// ApplyClient defines the interface for apply operations.
type ApplyClient interface {
	ClearTargetDeviations(ctx context.Context, resource *v1alpha1.TargetClearDeviation) error
}

// Apply reads each file (or stdin when path is "-"), decodes all YAML/JSON
// documents inside, and applies each one via the appropriate client method.
func Apply(ctx context.Context, cl ApplyClient, namespace string, filePaths []string, out io.Writer) error {
	for _, filePath := range filePaths {
		var data []byte
		var err error
		if filePath == "-" {
			data, err = io.ReadAll(os.Stdin)
		} else {
			data, err = os.ReadFile(filePath) // #nosec G304 – user-supplied path is intentional
		}
		if err != nil {
			return fmt.Errorf("reading %s: %w", filePath, err)
		}

		if err := applyDocuments(ctx, cl, namespace, data, out); err != nil {
			return fmt.Errorf("%s: %w", filePath, err)
		}
	}
	return nil
}

// applyDocuments handles multi-document YAML/JSON files.
func applyDocuments(ctx context.Context, cl ApplyClient, namespace string, data []byte, out io.Writer) error {
	decoder := k8syaml.NewYAMLOrJSONDecoder(bytes.NewReader(data), 4096)
	for {
		var raw json.RawMessage
		if err := decoder.Decode(&raw); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("decoding document: %w", err)
		}

		var tm metav1.TypeMeta
		if err := json.Unmarshal(raw, &tm); err != nil {
			return fmt.Errorf("reading type metadata: %w", err)
		}

		if err := applyResource(ctx, cl, namespace, tm.Kind, raw, out); err != nil {
			return err
		}
	}
	return nil
}

func applyResource(ctx context.Context, cl ApplyClient, namespace string, kind string, raw json.RawMessage, out io.Writer) error {
	switch kind {
	case v1alpha1.TargetClearDeviationKind:
		return applyTargetClearDeviation(ctx, cl, namespace, raw, out)
	default:
		return fmt.Errorf("unsupported kind %q", kind)
	}
}

func applyTargetClearDeviation(ctx context.Context, cl ApplyClient, namespace string, raw json.RawMessage, out io.Writer) error {
	var resource v1alpha1.TargetClearDeviation
	if err := json.Unmarshal(raw, &resource); err != nil {
		return fmt.Errorf("decoding TargetClearDeviation: %w", err)
	}

	// Prefer namespace from the resource manifest; fall back to the CLI namespace.
	if resource.Namespace == "" {
		resource.Namespace = namespace
	}

	if err := cl.ClearTargetDeviations(ctx, &resource); err != nil {
		return err
	}

	fmt.Fprintf(out, "targetcleardeviation/%s applied\n", resource.Name)
	return nil
}
