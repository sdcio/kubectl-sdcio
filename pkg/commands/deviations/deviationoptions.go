package deviations

type DeviationOptions struct {
	target    string
	namespace string
	// deviationName name which equals the config name
	deviationName string
	// preview show the preview pane
	preview bool
	// revert the selected entries
	revert bool
}

type DeviationOptionSetter func(d *DeviationOptions)

func NewDeviationOptions(namespace string, opts ...DeviationOptionSetter) *DeviationOptions {
	do := &DeviationOptions{
		namespace: namespace,
	}

	// apply options
	for _, o := range opts {
		o(do)
	}

	return do
}

// Getters
func (d *DeviationOptions) Target() string {
	return d.target
}

func (d *DeviationOptions) DeviationName() string {
	return d.deviationName
}

func (d *DeviationOptions) Preview() bool {
	return d.preview
}

func (d *DeviationOptions) Revert() bool {
	return d.revert
}

// Option setters
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

func WithTarget(target string) DeviationOptionSetter {
	return func(d *DeviationOptions) {
		d.target = target
	}
}

func WithDeviationName(name string) DeviationOptionSetter {
	return func(d *DeviationOptions) {
		d.deviationName = name
	}
}
