// Package client provides a gRPC data service client with Kubernetes port-forwarding support.
//
// The DataClient establishes a connection to the data-server service running in a
// Kubernetes cluster by:
//   1. Looking up pods matching the service selector
//   2. Establishing a port-forward tunnel to a ready pod
//   3. Connecting via gRPC over the local port-forward
//
// Usage example:
//
//	client, err := NewDataClient(restConfig, "sdc-system", "data-server", 56000)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer client.Close()
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	if err := client.Connect(ctx); err != nil {
//		log.Fatal(err)
//	}
//
//	// Fetch configuration in XPath/proto format
//	protoOutput, err := client.GetIntent(ctx, client.FormatXPath, "sdc.device1", "running")
//	if err != nil {
//		log.Fatal(err)
//	}
//	intent := protoOutput.GetProto()
//
//	// Or fetch configuration in JSON format
//	jsonOutput, err := client.GetIntent(ctx, client.FormatJSON, "sdc.device1", "running")
//	if err != nil {
//		log.Fatal(err)
//	}
//	jsonData := jsonOutput.GetBlob()
//	// Check what data type is available
//	if output.GetProto() != nil {
//		// Work with structured proto data
//	}
//	if output.GetBlob() != nil {
//		// Work with formatted blob data
//	}
//
// The unified Intent interface provides:
//   - String(): Formatted output for display
//   - GetBlob(): Raw blob data (JSON, XML, YAML) or nil for proto format
//   - GetProto(): Structured proto message data (XPath format) or nil for blob formats
//   - GetType(): Indicates the format of this output
//
// Error handling:
//
// The client provides custom error types for better context:
//   - ConnectionError: for port-forward or gRPC connection failures
//   - DataFetchError: for data retrieval failures
//

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/beevik/etree"
	"github.com/fatih/color"
	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

// DataClient is a gRPC client for the data-server service with automatic port-forwarding.
// It manages the lifecycle of port-forwarding tunnels and gRPC connections to the data server.
type DataClient struct {
	restConfig *rest.Config
	clientset  *kubernetes.Clientset
	namespace  string
	service    string
	port       int

	// Active port-forward management
	pf        *portforward.PortForwarder
	stopChan  chan struct{}
	localPort int
	conn      *grpc.ClientConn
}

// NewDataClient creates a new data service client that will connect via port-forward
func NewDataClient(restConfig *rest.Config, namespace, service string, port int) (*DataClient, error) {
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}

	return &DataClient{
		restConfig: restConfig,
		clientset:  clientset,
		namespace:  namespace,
		service:    service,
		port:       port,
	}, nil
}

// Connect establishes the port-forward and gRPC connection
func (d *DataClient) Connect(ctx context.Context) error {
	// Find a free local port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("failed to find free port: %w", err)
	}
	d.localPort = listener.Addr().(*net.TCPAddr).Port
	if err := listener.Close(); err != nil {
		return fmt.Errorf("failed to close listener: %w", err)
	}

	// Set up port-forward (this blocks until ready or fails)
	if err := d.setupPortForward(ctx); err != nil {
		return fmt.Errorf("failed to setup port-forward: %w", err)
	}

	conn, err := grpc.NewClient(fmt.Sprintf("127.0.0.1:%d", d.localPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		if closeErr := d.Close(); closeErr != nil {
			return fmt.Errorf("failed to connect to data service: %w (cleanup failed: %v)", err, closeErr)
		}
		return fmt.Errorf("failed to connect to data service: %w", err)
	}

	d.conn = conn
	return nil
}

// findDataPod finds a ready pod for the data service using the service's selector
func (d *DataClient) findDataPod(ctx context.Context) (*corev1.Pod, error) {
	// Get the service to find its selector
	svc, err := d.clientset.CoreV1().Services(d.namespace).Get(ctx, d.service, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get service %s: %w", d.service, err)
	}

	if len(svc.Spec.Selector) == 0 {
		return nil, fmt.Errorf("service %s has no selector", d.service)
	}

	// List pods matching the service selector
	labelSelector := metav1.FormatLabelSelector(&metav1.LabelSelector{
		MatchLabels: svc.Spec.Selector,
	})

	pods, err := d.clientset.CoreV1().Pods(d.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	if len(pods.Items) == 0 {
		return nil, fmt.Errorf("no pods found for service %s with selector %s", d.service, labelSelector)
	}

	// Find a ready pod
	for i := range pods.Items {
		pod := &pods.Items[i]
		if pod.Status.Phase == corev1.PodRunning {
			for _, cond := range pod.Status.Conditions {
				if cond.Type == corev1.PodReady && cond.Status == corev1.ConditionTrue {
					return pod, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("no ready pods found for service %s", d.service)
}

// setupPortForward creates a programmatic port-forward to the data service pod
func (d *DataClient) setupPortForward(ctx context.Context) error {
	// Find a ready pod
	pod, err := d.findDataPod(ctx)
	if err != nil {
		return err
	}

	// Build the URL for port-forward API
	hostIP := d.restConfig.Host
	parsedURL, err := url.Parse(hostIP)
	if err != nil {
		return fmt.Errorf("failed to parse host: %w", err)
	}

	// Construct the port-forward request URL to the specific pod
	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", d.namespace, pod.Name)
	pfURL := &url.URL{
		Scheme: parsedURL.Scheme,
		Host:   parsedURL.Host,
		Path:   path,
	}

	// Create SPDY round tripper
	transport, upgrader, err := spdy.RoundTripperFor(d.restConfig)
	if err != nil {
		return fmt.Errorf("failed to create round tripper: %w", err)
	}

	// Create dialer
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, pfURL)

	d.stopChan = make(chan struct{}, 1)
	readyChan := make(chan struct{})

	// Create port-forward
	ports := []string{fmt.Sprintf("%d:%d", d.localPort, d.port)}

	// Use io.Discard for quiet operation
	out, errOut := io.Discard, io.Discard

	pf, err := portforward.New(dialer, ports, d.stopChan, readyChan, out, errOut)
	if err != nil {
		return fmt.Errorf("failed to create port forwarder: %w", err)
	}

	d.pf = pf

	// Start port-forward in background
	errChan := make(chan error, 1)
	go func() {
		if err := pf.ForwardPorts(); err != nil {
			errChan <- err
		}
	}()

	// Wait for ready or error
	select {
	case err := <-errChan:
		return fmt.Errorf("port-forward failed: %w", err)
	case <-readyChan:
		return nil
	}
}

// GetConnection returns the gRPC connection for use with data service clients
func (d *DataClient) GetConnection() *grpc.ClientConn {
	return d.conn
}

// getIntentResponse is a helper that fetches an intent response from the data server
func (d *DataClient) getIntentResponse(ctx context.Context, datastoreName, intentName string, format sdcpb.Format) (*sdcpb.GetIntentResponse, error) {
	if d.conn == nil {
		return nil, fmt.Errorf("not connected to data server")
	}

	client := sdcpb.NewDataServerClient(d.conn)

	req := &sdcpb.GetIntentRequest{
		DatastoreName: datastoreName,
		Intent:        intentName,
		Format:        format,
	}

	return client.GetIntent(ctx, req)
}

// GetIntent fetches configuration data in the specified format.
// Accepts internal Format enum and returns a BlobOutput interface that provides
// access to the data in the requested format. For proto/xpath format, call GetProto().
// For blob formats (JSON, XML, YAML), call GetBlob(). Use GetType() to determine
// the underlying format type.
func (d *DataClient) GetIntent(ctx context.Context, format Format, datastoreName, intentName string) (Intent, error) {
	// Convert internal Format to sdcpb.Format for fetching
	var sdcpbFormat sdcpb.Format

	switch format {
	case FormatJSON:
		sdcpbFormat = sdcpb.Format_Intent_Format_JSON
	case FormatJSONIETF:
		sdcpbFormat = sdcpb.Format_Intent_Format_JSON_IETF
	case FormatXML:
		sdcpbFormat = sdcpb.Format_Intent_Format_XML
	case FormatXPath:
		sdcpbFormat = sdcpb.Format_Intent_Format_PROTO
	case FormatYAML:
		// YAML is not a native backend format; fetch as JSON
		sdcpbFormat = sdcpb.Format_Intent_Format_JSON
	default:
		return nil, fmt.Errorf("invalid format %q", format)
	}

	resp, err := d.getIntentResponse(ctx, datastoreName, intentName, sdcpbFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to get intent: %w", err)
	}

	// Create and return the appropriate output type based on format
	return newIntentOutput(format, resp)
}

// Close terminates the gRPC connection and port-forward
func (d *DataClient) Close() error {
	var errs []error

	if d.conn != nil {
		if err := d.conn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close gRPC connection: %w", err))
		}
		d.conn = nil
	}

	if d.stopChan != nil {
		close(d.stopChan)
		d.stopChan = nil
	}

	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

// ConnectionError represents a failure in establishing or maintaining a connection to the data server.
// It includes information about which component failed and the underlying error.
type ConnectionError struct {
	Component string // The component that failed (e.g., "port-forward", "grpc")
	Reason    string // Human-readable reason for the failure
	Err       error  // The underlying error
}

func (e *ConnectionError) Error() string {
	return fmt.Sprintf("connection error (%s): %s: %v", e.Component, e.Reason, e.Err)
}

func (e *ConnectionError) Unwrap() error {
	return e.Err
}

// DataFetchError represents a failure when fetching configuration data from the server.
// It includes the datastore name, intent name, and specific reason for the failure.
type DataFetchError struct {
	DatastoreName string // The datastore being accessed (e.g., "sdc.device1")
	IntentName    string // The intent being fetched (e.g., "running", "config")
	Reason        string // Human-readable reason for the failure
	Err           error  // The underlying error
}

func (e *DataFetchError) Error() string {
	return fmt.Sprintf("failed to fetch %s/%s: %s: %v", e.DatastoreName, e.IntentName, e.Reason, e.Err)
}

func (e *DataFetchError) Unwrap() error {
	return e.Err
}

// newIntentOutput is a factory function that creates the appropriate Intent implementation
// based on the requested format and the response data.
func newIntentOutput(format Format, resp *sdcpb.GetIntentResponse) (Intent, error) {
	switch format {
	case FormatXPath:
		proto := resp.GetProto()
		if proto == nil {
			return nil, fmt.Errorf("no proto data in response")
		}
		return &ProtoConfigOutput{intent: proto}, nil
	case FormatJSON, FormatJSONIETF:
		blob := resp.GetBlob()
		if blob == nil {
			return nil, fmt.Errorf("no blob data in response")
		}
		return &JSONBlobConfigOutput{blob: blob}, nil
	case FormatXML:
		blob := resp.GetBlob()
		if blob == nil {
			return nil, fmt.Errorf("no blob data in response")
		}
		return &XMLBlobConfigOutput{blob: blob}, nil
	case FormatYAML:
		blob := resp.GetBlob()
		if blob == nil {
			return nil, fmt.Errorf("no blob data in response")
		}
		return &YAMLBlobConfigOutput{blob: blob}, nil
	default:
		blob := resp.GetBlob()
		if blob == nil {
			return nil, fmt.Errorf("no blob data in response")
		}
		return &RawBlobConfigOutput{blob: blob}, nil
	}
}

// ProtoConfigOutput holds configuration data in PROTO/XPath format.
// Implements BlobOutput interface with proto data access.
type ProtoConfigOutput struct {
	intent *sdcpb.Intent
}

// String returns a string representation of the proto output with formatted XPath entries.
// Displays the intent data as sorted XPath path: value pairs with colored values.
func (f *ProtoConfigOutput) String() string {
	if f.intent == nil {
		return ""
	}

	lines := make([]string, 0, len(f.intent.Update))
	valueReplacer := strings.NewReplacer(
		"\n", "\\n",
		"\r", "\\r",
		"\t", "\\t",
	)
	for _, update := range f.intent.Update {
		path := update.Path.ToXPath(false)
		value := update.GetValue().ToString()
		// Escape special characters to show them literally (\n, \t, etc)
		escapedValue := valueReplacer.Replace(value)
		// Color the value in cyan for better distinction
		coloredValue := color.CyanString(escapedValue)
		lines = append(lines, fmt.Sprintf("%s: %s", path, coloredValue))
	}

	slices.Sort(lines)
	var result strings.Builder
	for i, line := range lines {
		if i > 0 {
			result.WriteString("\n")
		}
		result.WriteString(line)
	}
	return result.String()
}

// GetBlob returns nil for proto output (proto data is accessed via GetProto).
func (f *ProtoConfigOutput) GetBlob() []byte {
	return nil
}

// GetProto returns the underlying protobuf Intent message containing the structured configuration data.
func (f *ProtoConfigOutput) GetProto() *sdcpb.Intent {
	return f.intent
}

// GetType returns the format type for this output.
func (f *ProtoConfigOutput) GetType() Format {
	return FormatXPath
}

// Format represents the output format for configuration data
type Format string

const (
	FormatJSON     Format = "json"
	FormatJSONIETF Format = "json-ietf"
	FormatXML      Format = "xml"
	FormatXPath    Format = "xpath"
	FormatYAML     Format = "yaml"
)

// Intent is the interface for configuration output (both blob and proto formats).
// All implementations provide String(), GetBlob(), GetProto(), and GetType() methods.
// For proto format, GetProto() returns the intent and GetBlob() returns nil.
// For blob formats (JSON, XML, YAML), GetBlob() returns the data and GetProto() returns nil.
type Intent interface {
	String() string
	GetBlob() []byte
	GetProto() *sdcpb.Intent
	GetType() Format
}

// JSONBlobConfigOutput holds configuration data in JSON format.
// String() returns automatically pretty-printed JSON output.
type JSONBlobConfigOutput struct {
	blob []byte
}

// String returns the blob data as a formatted JSON string.
// Automatically pretty-prints for readability.
// Returns raw data if pretty-printing fails.
func (f *JSONBlobConfigOutput) String() string {
	var obj interface{}
	if err := json.Unmarshal(f.blob, &obj); err != nil {
		// If parsing fails, return raw data
		return string(f.blob)
	}
	pretty, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		// If re-marshaling fails, return raw data
		return string(f.blob)
	}
	return string(pretty)
}

// GetBlob returns the raw configuration data as bytes.
func (f *JSONBlobConfigOutput) GetBlob() []byte {
	return f.blob
}

// GetProto returns nil for blob output (blob data is accessed via GetBlob).
func (f *JSONBlobConfigOutput) GetProto() *sdcpb.Intent {
	return nil
}

// GetType returns the format type for this output.
func (f *JSONBlobConfigOutput) GetType() Format {
	return FormatJSON
}

// XMLBlobConfigOutput holds configuration data in XML format.
// String() returns automatically pretty-printed XML output.
type XMLBlobConfigOutput struct {
	blob []byte
}

// String returns the blob data as a formatted XML string.
// Automatically pretty-prints with indentation using etree.
// Returns raw data if parsing fails.
func (f *XMLBlobConfigOutput) String() string {
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(f.blob); err != nil {
		// If parsing fails, return raw data
		return string(f.blob)
	}
	// Format with indentation
	doc.Indent(2)
	output, err := doc.WriteToString()
	if err != nil {
		// If formatting fails, return raw data
		return string(f.blob)
	}
	return output
}

// GetBlob returns the raw configuration data as bytes.
func (f *XMLBlobConfigOutput) GetBlob() []byte {
	return f.blob
}

// GetProto returns nil for blob output (blob data is accessed via GetBlob).
func (f *XMLBlobConfigOutput) GetProto() *sdcpb.Intent {
	return nil
}

// GetType returns the format type for this output.
func (f *XMLBlobConfigOutput) GetType() Format {
	return FormatXML
}

// YAMLBlobConfigOutput holds configuration data in YAML format.
// String() returns automatically formatted YAML output (converted from JSON).
type YAMLBlobConfigOutput struct {
	blob []byte
}

// String returns the blob data as a formatted YAML string.
// Converts from JSON to YAML format.
// Returns raw data if conversion fails.
func (f *YAMLBlobConfigOutput) String() string {
	var obj interface{}
	// First unmarshal from JSON
	if err := json.Unmarshal(f.blob, &obj); err != nil {
		// If parsing fails, return raw data
		return string(f.blob)
	}
	// Re-marshal as YAML
	yamlData, err := yaml.Marshal(obj)
	if err != nil {
		// If re-marshaling fails, return raw data
		return string(f.blob)
	}
	return string(yamlData)
}

// GetBlob returns the raw configuration data as bytes.
func (f *YAMLBlobConfigOutput) GetBlob() []byte {
	return f.blob
}

// GetProto returns nil for blob output (blob data is accessed via GetBlob).
func (f *YAMLBlobConfigOutput) GetProto() *sdcpb.Intent {
	return nil
}

// GetType returns the format type for this output.
func (f *YAMLBlobConfigOutput) GetType() Format {
	return FormatYAML
}

// RawBlobConfigOutput holds configuration data in raw/unspecified format.
// String() returns the raw data as-is without any formatting or pretty-printing.
type RawBlobConfigOutput struct {
	blob []byte
}

// String returns the raw blob data as a string without any formatting.
func (f *RawBlobConfigOutput) String() string {
	return string(f.blob)
}

// GetBlob returns the raw configuration data as bytes.
func (f *RawBlobConfigOutput) GetBlob() []byte {
	return f.blob
}

// GetProto returns nil for blob output (blob data is accessed via GetBlob).
func (f *RawBlobConfigOutput) GetProto() *sdcpb.Intent {
	return nil
}

// GetType returns the format type for this output.
func (f *RawBlobConfigOutput) GetType() Format {
	return ""
}
