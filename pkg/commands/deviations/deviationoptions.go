package deviations

type DeviationOptions struct {
	namespace string
	// deviationName name which equals the config name
	deviationName string
	// preview show the preview pane
	preview bool
	// revert the selected entries
	revert bool
}

type DeviationOptionSetter func(d *DeviationOptions)

func NewDeviationOptions(deviationName, namespace string, opts ...DeviationOptionSetter) *DeviationOptions {
	do := &DeviationOptions{
		deviationName: deviationName,
		namespace:     namespace,
	}

	// apply options
	for _, o := range opts {
		o(do)
	}

	return do
}

func WithPreview(b bool) DeviationOptionSetter {
	return func(d *DeviationOptions) {
		d.preview = b
	}
}

func WithRevert(b bool) DeviationOptionSetter {
	return func(d *DeviationOptions) {
		d.revert = b
	}
}
