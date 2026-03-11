package types

import "strings"

type Deviations map[string]*IntentDeviations

func (d Deviations) AddDeviation(deviation *IntentDeviations) {
	d[deviation.Name()] = deviation
}

func (d Deviations) AddDeviationItem(deviation *Deviation) {
	// find the IntentDeviations for the deviation's parent target, or create a new one if it doesn't exist
	intentDev, ok := d[deviation.parent.Name()]
	if !ok {
		intentDev = &IntentDeviations{
			name:       deviation.parent.Name(),
			target:     deviation.parent.Target(),
			namespace:  deviation.parent.Namespace(),
			typ:        deviation.parent.Type(),
			deviations: []*Deviation{},
		}
		d.AddDeviation(intentDev)
	}
	intentDev.deviations = append(intentDev.deviations, deviation)
}

func (d Deviations) HasDeviations() bool {
	for _, dev := range d {
		if dev.Length() > 0 {
			return true
		}
	}
	return false
}

func (d Deviations) MultipleIntents() bool {
	return len(d) > 1
}

func (d Deviations) First() *IntentDeviations {
	for _, dev := range d {
		return dev
	}
	return nil
}

// Items return the deviations of all the IntentDeviations as a single slice
func (d Deviations) Items() DeviationSlice {
	length := 0
	for _, dev := range d {
		length += dev.Length()
	}
	allDevs := make(DeviationSlice, 0, length)
	for _, dev := range d {
		allDevs = append(allDevs, dev.Deviations()...)
	}
	return allDevs
}

func (d Deviations) MaxDeviationNameLength() int {
	maxLength := 0
	for name := range d {
		if len(name) > maxLength {
			maxLength = len(name)
		}
	}
	return maxLength
}

func (d Deviations) String() string {
	var b strings.Builder
	for _, dev := range d {
		b.WriteString(dev.String())
		b.WriteString("\n")
	}
	return b.String()
}

type IntentDeviations struct {
	target     string
	name       string
	namespace  string
	typ        DeviationType
	deviations []*Deviation
}

func NewDeviations(target string, name string, deviationType DeviationType, length int) *IntentDeviations {
	return &IntentDeviations{
		target:     target,
		name:       name,
		typ:        deviationType,
		deviations: make([]*Deviation, 0, length),
	}
}

func (d *IntentDeviations) Name() string {
	return d.name
}

func (d *IntentDeviations) Length() int {
	return len(d.deviations)
}

func (d *IntentDeviations) Target() string {
	return d.target
}

func (d *IntentDeviations) Namespace() string {
	return d.namespace
}

func (d *IntentDeviations) Type() DeviationType {
	return d.typ
}

func (d *IntentDeviations) SetNamespace(namespace string) *IntentDeviations {
	d.namespace = namespace
	return d
}

func (d *IntentDeviations) Deviations() []*Deviation {
	return d.deviations
}

func (d *IntentDeviations) DeviationPaths() []string {
	paths := make([]string, 0, len(d.deviations))
	for _, dev := range d.deviations {
		paths = append(paths, dev.Path)
	}
	return paths
}

func (d *IntentDeviations) String() string {
	var b strings.Builder
	b.WriteString("Name: ")
	b.WriteString(d.Name())
	b.WriteString("\nNamespace: ")
	b.WriteString(d.Namespace())
	b.WriteString("\nType: ")
	b.WriteString(string(d.Type()))
	b.WriteString("\nDeviations:\n")
	for _, dev := range d.deviations {
		b.WriteString(dev.StringIndent("  "))
		b.WriteString("\n")
	}
	return b.String()
}

type Deviation struct {
	parent       *IntentDeviations
	ActualValue  string `json:"actualValue" yaml:"actualValue"`
	DesiredValue string `json:"desiredValue" yaml:"desiredValue"`
	Path         string `json:"path" yaml:"path"`
	Reason       string `json:"reason" yaml:"reason"`
}

func NewDeviation(path string, desiredValue string, actualValue string, reason string) *Deviation {
	return &Deviation{ActualValue: actualValue, DesiredValue: desiredValue, Path: path, Reason: reason}
}

func (d *Deviation) SetParent(deviations *IntentDeviations) {
	d.parent = deviations
}

func (d *Deviation) Name() string {
	return d.parent.Name()
}

func (d *Deviation) StringIndent(indent string) string {
	return indent + strings.ReplaceAll(d.String(), "\n", "\n"+indent)
}

func (d *Deviation) String() string {
	var b strings.Builder
	b.WriteString("Path: ")
	b.WriteString(d.Path)
	b.WriteString("\nActual Value: ")
	b.WriteString(d.ActualValue)
	b.WriteString("\nReason: ")
	b.WriteString(d.Reason)
	b.WriteString("\nDesired Value: ")
	b.WriteString(d.DesiredValue)
	return b.String()
}

type DeviationType string

const (
	DeviationTypeUnknown DeviationType = "unknown"
	DeviationTypeTarget  DeviationType = "target"
	DeviationTypeConfig  DeviationType = "config"
)

func (d *IntentDeviations) AddDeviation(dev *Deviation) {
	dev.SetParent(d)
	d.deviations = append(d.deviations, dev)
}

type DeviationSlice []*Deviation

func (d DeviationSlice) Len() int {
	return len(d)
}

func (d DeviationSlice) FilterByIndexes(indexes []int) Deviations {
	filtered := Deviations{}
	for _, idx := range indexes {
		if idx >= 0 && idx < len(d) {
			filtered.AddDeviationItem(d[idx])
		}
	}
	return filtered
}
