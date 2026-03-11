package client

import (
	"context"

	"github.com/sdcio/config-server/apis/config/v1alpha1"
	configCR "github.com/sdcio/config-server/pkg/generated/clientset/versioned"
	"github.com/sdcio/kubectl-sdc/pkg/types"
	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
	"google.golang.org/protobuf/encoding/protojson"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/metadata"
	"k8s.io/client-go/rest"
)

const (
	// TargetLabel is the label used to identify targets in Kubernetes
	TargetLabel = "config.sdcio.dev/targetName"
)

type ConfigClient struct {
	// c is the clientset for interacting with the config server CRDs
	c *configCR.Clientset
	// mdClient is used for fetching metadata like names of resources without fetching the entire object
	mdClient metadata.Interface
}

func NewConfigClient(restConfig *rest.Config) (*ConfigClient, error) {
	// Create the clientset for the config server CRDs
	clientset, err := configCR.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	// Create a dynamic client for metadata operations
	mdclient, err := metadata.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	return &ConfigClient{
		c:        clientset,
		mdClient: mdclient,
	}, nil
}

// GetDeviationByName retrieves a specific deviation by name and converts it to the internal type
func (c *ConfigClient) GetDeviationByName(ctx context.Context, namespace string, deviationName string) (*types.IntentDeviations, error) {
	resp, err := c.c.ConfigV1alpha1().Deviations(namespace).Get(ctx, deviationName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return ConvertDeviationIntent(resp)
}

// GetDeviationsByTarget retrieves all deviations for a given target and converts them to the internal type
func (c *ConfigClient) GetDeviationsByTarget(ctx context.Context, namespace string, targetName string) (types.Deviations, error) {
	labelselector := metav1.FormatLabelSelector(&metav1.LabelSelector{MatchLabels: map[string]string{TargetLabel: targetName}})

	resp, err := c.c.ConfigV1alpha1().Deviations(namespace).List(ctx, metav1.ListOptions{LabelSelector: labelselector})
	if err != nil {
		return nil, err
	}

	return ConvertDeviations(resp)
}

func (c *ConfigClient) ListDeviationNames(ctx context.Context, namespace string, labels map[string]string) ([]string, error) {
	// Define the GVR for your CRD
	gvr := schema.GroupVersionResource{
		Group:    "config.sdcio.dev", // Replace with your actual group
		Version:  "v1alpha1",
		Resource: "deviations",
	}

	listOptions := metav1.ListOptions{}
	if len(labels) > 0 {
		listOptions.LabelSelector = metav1.FormatLabelSelector(&metav1.LabelSelector{MatchLabels: labels})
	}

	// This call sends the "Accept: application/json;as=PartialObjectMetadataList" header automatically
	list, err := c.mdClient.Resource(gvr).Namespace(namespace).List(ctx, listOptions)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(list.Items))
	for _, item := range list.Items {
		names = append(names, item.Name)
	}

	return names, nil
}

func (c *ConfigClient) GetBlameTree(ctx context.Context, namespace string, device string) (*sdcpb.BlameTreeElement, error) {
	resp, err := c.c.ConfigV1alpha1().ConfigBlames(namespace).Get(ctx, device, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	bte := &sdcpb.BlameTreeElement{}
	err = protojson.Unmarshal([]byte(resp.Status.Value.Raw), bte)
	if err != nil {
		return nil, err
	}
	return bte, nil
}

func (c *ConfigClient) ListTargetNames(ctx context.Context, namespace string) ([]string, error) {
	gvr := schema.GroupVersionResource{
		Group:    "config.sdcio.dev",
		Version:  "v1alpha1",
		Resource: "targets",
	}

	resp, err := c.mdClient.Resource(gvr).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	result := make([]string, 0, len(resp.Items))

	for _, i := range resp.Items {
		result = append(result, i.Name)
	}

	return result, nil
}

// ListRunningConfigNames lists all running config names in a namespace
func (c *ConfigClient) ListRunningConfigNames(ctx context.Context, namespace string) ([]string, error) {
	gvr := schema.GroupVersionResource{
		Group:    "config.sdcio.dev",
		Version:  "v1alpha1",
		Resource: "runningconfigs",
	}

	resp, err := c.mdClient.Resource(gvr).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	result := make([]string, 0, len(resp.Items))

	for _, i := range resp.Items {
		result = append(result, i.Name)
	}

	return result, nil
}

// NewTargetClearDeviation builds a TargetClearDeviation request body from
// internal deviation state, ready to be posted to the cleardeviation subresource.
func NewTargetClearDeviation(namespace, targetName string, devs types.Deviations) *v1alpha1.TargetClearDeviation {
	content := make([]v1alpha1.TargetClearDeviationConfig, 0, len(devs))
	for _, d := range devs {
		content = append(content, v1alpha1.TargetClearDeviationConfig{
			Name:  d.Name(),
			Paths: d.DeviationPaths(),
		})
	}

	r := &v1alpha1.TargetClearDeviation{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.TargetClearDeviationKind,
			APIVersion: v1alpha1.SchemeGroupVersion.Identifier(),
		},
		Spec: &v1alpha1.TargetClearDeviationSpec{
			Config: content,
		},
	}
	r.SetName(targetName)
	r.SetNamespace(namespace)
	return r
}

// ClearTargetDeviations posts a pre-built TargetClearDeviation to the
// cleardeviation subresource. Use NewTargetClearDeviation (convert.go) to
// construct the resource from a types.Deviations value.
func (c *ConfigClient) ClearTargetDeviations(ctx context.Context, resource *v1alpha1.TargetClearDeviation) error {
	restClient := c.c.ConfigV1alpha1().RESTClient()

	result := restClient.
		Post().
		Namespace(resource.Namespace).
		Resource("targets").
		Name(resource.Name).
		SubResource(resource.SubResourceName()).
		Body(resource).
		Do(ctx)

	return result.Error()
}
