package client

import (
	"context"

	"github.com/sdcio/config-server/apis/config/v1alpha1"
	configCR "github.com/sdcio/config-server/pkg/generated/clientset/versioned"
	"github.com/sdcio/kubectl-sdcio/pkg/types"
	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
	"google.golang.org/protobuf/encoding/protojson"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/metadata"
	"k8s.io/client-go/rest"
)

type ConfigClient struct {
	c        *configCR.Clientset
	mdClient metadata.Interface
}

func NewConfigClient(restConfig *rest.Config) (*ConfigClient, error) {
	clientset, err := configCR.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	mdclient, err := metadata.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	return &ConfigClient{
		c:        clientset,
		mdClient: mdclient,
	}, nil
}

func (c *ConfigClient) GetDeviations(ctx context.Context, namespace string, deviationName string) (*types.Deviations, error) {
	resp, err := c.c.ConfigV1alpha1().Deviations(namespace).Get(ctx, deviationName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return ConvertDeviations(resp)
}

func (c *ConfigClient) ListDeviations(ctx context.Context, namespace string, labels map[string]string) ([]*types.Deviations, error) {
	labelselector := metav1.FormatLabelSelector(&metav1.LabelSelector{MatchLabels: labels})
	resp, err := c.c.ConfigV1alpha1().Deviations(namespace).List(ctx, metav1.ListOptions{LabelSelector: labelselector})
	if err != nil {
		return nil, err
	}

	var deviations []*types.Deviations
	for _, item := range resp.Items {
		dev, err := ConvertDeviations(&item)
		if err != nil {
			return nil, err
		}
		deviations = append(deviations, dev)
	}

	return deviations, nil
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

// ClearTargetDeviations clears the specified deviation paths on a target
func (c *ConfigClient) ClearTargetDeviations(ctx context.Context, namespace, targetName, configName string, paths []string) error {
	if len(paths) == 0 {
		return nil // Nothing to clear
	}

	// Get the REST client from the clientset
	restClient := c.c.ConfigV1alpha1().RESTClient()

	// Build the request body using the typed struct
	// The Config field expects a slice of TargetClearDeviationConfig
	// We create a single config entry with the target name and paths
	clearRequest := &v1alpha1.TargetClearDeviation{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.TargetClearDeviationKind,
			APIVersion: v1alpha1.SchemeGroupVersion.Identifier(),
		},
		Spec: &v1alpha1.TargetClearDeviationSpec{
			Config: []v1alpha1.TargetClearDeviationConfig{
				{
					Name:  configName,
					Paths: paths,
				},
			},
		},
	}
	clearRequest.SetName(targetName)
	clearRequest.SetNamespace(namespace)

	// Make the POST request to the cleardeviation subresource
	result := restClient.
		Post().
		Namespace(namespace).
		Resource("targets").
		Name(targetName).
		SubResource(clearRequest.SubResourceName()).
		Body(clearRequest).
		Do(ctx)

	return result.Error()
}
