package client

import (
	"fmt"

	"github.com/sdcio/config-server/apis/config/v1alpha1"
	"github.com/sdcio/kubectl-sdcio/pkg/types"
	"github.com/sdcio/sdc-protos/sdcpb"
	"google.golang.org/protobuf/encoding/prototext"
)

func ConvertDeviations(d *v1alpha1.Deviation) (*types.Deviations, error) {
	if d.Spec.DeviationType == nil {
		return nil, fmt.Errorf("deviation type field is nil for deviation %s", d.GetName())
	}
	dt, err := ConvertDeviationType(*d.Spec.DeviationType)
	if err != nil {
		return nil, err
	}

	// extract target name via label
	target, ok := d.Labels["config.sdcio.dev/targetName"]
	if !ok {
		return nil, fmt.Errorf("deviation %s is missing the target label 'config.sdcio.dev/targetName'", d.Name)
	}

	result := types.NewDeviations(target, d.GetName(), dt, len(d.Spec.Deviations)).SetNamespace(d.GetNamespace())

	for _, dev := range d.Spec.Deviations {
		dev, err := ConvertDeviation(&dev)
		if err != nil {
			return nil, err
		}
		result.AddDeviation(dev)
	}
	return result, nil
}

func ConvertDeviation(d *v1alpha1.ConfigDeviation) (*types.Deviation, error) {
	var err error
	desired := ""
	current := ""

	if d.DesiredValue != nil {
		desired, err = convertTypedValueTextToString(*d.DesiredValue)
		if err != nil {
			return nil, err
		}
	}
	if d.ActualValue != nil {
		current, err = convertTypedValueTextToString(*d.ActualValue)
		if err != nil {
			return nil, err
		}
	}
	return types.NewDeviation(d.Path, desired, current, d.Reason), nil
}

func ConvertDeviationType(dt v1alpha1.DeviationType) (types.DeviationType, error) {
	switch dt {
	case v1alpha1.DeviationType_TARGET:
		return types.DeviationTypeTarget, nil
	case v1alpha1.DeviationType_CONFIG:
		return types.DeviationTypeConfig, nil
	default:
		return types.DeviationTypeUnknown, fmt.Errorf("Unknown deviation type %v", dt)
	}
}

func convertTypedValueTextToString(tvText string) (string, error) {
	tv := &sdcpb.TypedValue{}
	err := prototext.Unmarshal([]byte(tvText), tv)
	if err != nil {
		return "", err
	}
	return tv.ToString(), nil
}
