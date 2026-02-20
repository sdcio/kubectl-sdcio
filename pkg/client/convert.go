package client

import (
	"strings"

	"github.com/sdcio/config-server/apis/config/v1alpha1"
	"github.com/sdcio/kubectl-sdcio/pkg/types"
)

func ConvertDeviations(d *v1alpha1.Deviation) *types.Deviations {
	dt := ConvertDeviationType(*d.Spec.DeviationType)
	result := types.NewDeviations(d.GetName(), dt, len(d.Spec.Deviations)).SetNamespace(d.GetNamespace())

	for _, dev := range d.Spec.Deviations {
		result.AddDeviation(ConvertDeviation(&dev))
	}
	return result
}

func stripValuePrefix(value string) string {
	// Remove prefixes like "json_val:", "string_val:", "uint_val:", etc.
	if idx := strings.Index(value, "_val:"); idx != -1 {
		return value[idx+5:]
	}
	return value
}

func ConvertDeviation(d *v1alpha1.ConfigDeviation) *types.Deviation {
	desired := ""
	current := ""
	if d.DesiredValue != nil {
		desired = stripValuePrefix(*d.DesiredValue)
	}
	if d.CurrentValue != nil {
		current = stripValuePrefix(*d.CurrentValue)
	}
	return types.NewDeviation(d.Path, desired, current, d.Reason)
}

func ConvertDeviationType(dt v1alpha1.DeviationType) types.DeviationType {
	switch dt {
	case v1alpha1.DeviationType_TARGET:
		return types.DeviationTypeTarget
	case v1alpha1.DeviationType_CONFIG:
		return types.DeviationTypeConfig
	default:
		return types.DeviationTypeUnknown
	}
}
