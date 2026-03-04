package client

import (
	"context"

	configCR "github.com/sdcio/config-server/pkg/generated/clientset/versioned"
	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
	"google.golang.org/protobuf/encoding/protojson"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type ConfigClient struct {
	c *configCR.Clientset
}

func NewConfigClient(restConfig *rest.Config) (*ConfigClient, error) {
	clientset, err := configCR.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	return &ConfigClient{
		c: clientset,
	}, nil
}

func (c *ConfigClient) GetBlameTree(ctx context.Context, namespace string, device string) (*sdcpb.BlameTreeElement, error) {
	resp, err := c.c.ConfigV1alpha1().ConfigBlames(namespace).Get(ctx, device, v1.GetOptions{})
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

func (c *ConfigClient) GetTargetNames(ctx context.Context, namespace string) ([]string, error) {
	resp, err := c.c.InvV1alpha1().Targets(namespace).List(ctx, v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	result := make([]string, 0, len(resp.Items))

	for _, i := range resp.Items {
		result = append(result, i.Name)
	}

	return result, nil
}
