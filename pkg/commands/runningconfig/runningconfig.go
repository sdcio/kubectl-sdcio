package runningconfig

import (
	"context"
	"fmt"
	"strings"

	"github.com/sdcio/kubectl-sdc/pkg/client"
	corev1 "k8s.io/api/core/v1"
)

const defaultDataServicePort = 56000

// DataClient defines the subset of the data client used by runningconfig.
type DataClient interface {
	Connect(ctx context.Context) error
	GetIntent(ctx context.Context, format client.Format, datastoreName, intentName string) (client.Intent, error)
	Close() error
}

// ValidFormats lists all supported output formats.
var ValidFormats = []client.Format{
	client.FormatJSON,
	client.FormatJSONIETF,
	client.FormatXML,
	client.FormatXPath,
	client.FormatYAML,
}

// FormatListString returns a comma-separated string of valid formats.
func FormatListString() string {
	return strings.Join(ValidFormatStrings(), ", ")
}

// ValidFormatStrings returns the list of valid format strings.
func ValidFormatStrings() []string {
	formatted := make([]string, len(ValidFormats))
	for i, f := range ValidFormats {
		formatted[i] = string(f)
	}
	return formatted
}

// ParseFormat converts a format string to the internal format enum.
func ParseFormat(formatStr string) (client.Format, error) {
	switch client.Format(strings.ToLower(formatStr)) {
	case client.FormatJSON:
		return client.FormatJSON, nil
	case client.FormatJSONIETF:
		return client.FormatJSONIETF, nil
	case client.FormatXML:
		return client.FormatXML, nil
	case client.FormatXPath:
		return client.FormatXPath, nil
	case client.FormatYAML:
		return client.FormatYAML, nil
	default:
		return "", fmt.Errorf("invalid format %q, must be one of: %s", formatStr, FormatListString())
	}
}

// ResolveDataServicePort extracts the data-service port from the service, falling back to the default port.
func ResolveDataServicePort(svc *corev1.Service) (int, error) {
	if len(svc.Spec.Ports) == 0 {
		return 0, fmt.Errorf("data-server service has no ports")
	}

	for _, port := range svc.Spec.Ports {
		if port.Name == "data-service" {
			return int(port.TargetPort.IntVal), nil
		}
	}

	return defaultDataServicePort, nil
}

// Run connects to the data server and fetches the running configuration for the target.
func Run(ctx context.Context, dataClient DataClient, namespace, target string, format client.Format) (string, error) {
	if err := dataClient.Connect(ctx); err != nil {
		return "", fmt.Errorf("failed to connect to data-server: %w", err)
	}

	datastoreName := fmt.Sprintf("%s.%s", namespace, target)
	configOutput, err := dataClient.GetIntent(ctx, format, datastoreName, "running")
	if err != nil {
		return "", err
	}

	return configOutput.String(), nil
}
