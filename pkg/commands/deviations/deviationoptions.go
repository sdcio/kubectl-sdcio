package deviations

type DeviationOptions struct {
	target    string
	namespace string
	// deviationName name which equals the config name
	deviationName string
	interactive   bool
	// preview show the preview pane
	preview bool
	// revert the selected entries
	revert                     bool
	initialQuery               string
	selectPathPrefix           []string
	filterPath                 []string
	autoAcceptSelectPathPrefix bool
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

func (d *DeviationOptions) Interactive() bool {
	return d.interactive
}

func (d *DeviationOptions) Preview() bool {
	return d.preview
}

func (d *DeviationOptions) Revert() bool {
	return d.revert
}

func (d *DeviationOptions) SelectPathPrefix() []string {
	return d.selectPathPrefix
}

func (d *DeviationOptions) FilterPath() []string {
	return d.filterPath
}

func (d *DeviationOptions) InitialQuery() string {
	return d.initialQuery
}

func (d *DeviationOptions) AutoAcceptSelectPathPrefix() bool {
	return d.autoAcceptSelectPathPrefix
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

func WithInteractive(interactive bool) DeviationOptionSetter {
	return func(d *DeviationOptions) {
		d.interactive = interactive
	}
}

func WithInitialQuery(query string) DeviationOptionSetter {
	return func(d *DeviationOptions) {
		d.initialQuery = query
	}
}

func WithSelectPathPrefix(prefixes []string) DeviationOptionSetter {
	return func(d *DeviationOptions) {
		d.selectPathPrefix = prefixes
	}
}

func WithFilterPath(prefixes []string) DeviationOptionSetter {
	return func(d *DeviationOptions) {
		d.filterPath = prefixes
	}
}

func WithAutoAcceptSelectPathPrefix(autoAccept bool) DeviationOptionSetter {
	return func(d *DeviationOptions) {
		d.autoAcceptSelectPathPrefix = autoAccept
	}
}
