package types

import "strings"

type Deviations struct {
	name       string
	namespace  string
	typ        DeviationType
	deviations []Deviation
}

func NewDeviations(name string, deviationType DeviationType, length int) *Deviations {
	return &Deviations{
		name:       name,
		typ:        deviationType,
		deviations: make([]Deviation, 0, length),
	}
}

func (d *Deviations) Name() string {
	return d.name
}

func (d *Deviations) Length() int {
	return len(d.deviations)
}

func (d *Deviations) Namespace() string {
	return d.namespace
}

func (d *Deviations) Type() DeviationType {
	return d.typ
}

func (d *Deviations) SetNamespace(namespace string) *Deviations {
	d.namespace = namespace
	return d
}

func (d *Deviations) Deviations() []Deviation {
	return d.deviations
}

func (d *Deviations) String() string {
	var b strings.Builder
	b.WriteString("Name: ")
	b.WriteString(d.Name())
	b.WriteString("\nNamespace: ")
	b.WriteString(d.Namespace())
	b.WriteString("\nType: ")
	b.WriteString(string(d.Type()))
	b.WriteString("\nDeviations:\n")
	for _, dev := range d.deviations {
		b.WriteString(dev.String())
		b.WriteString("\n\n")
	}
	return b.String()
}

type Deviation struct {
	parent       *Deviations
	actualValue  string
	desiredValue string
	path         string
	reason       string
}

func NewDeviation(path string, desiredValue string, actualValue string, reason string) *Deviation {
	return &Deviation{actualValue: actualValue, desiredValue: desiredValue, path: path, reason: reason}
}

func (d *Deviation) SetParent(deviations *Deviations) {
	d.parent = deviations
}

func (d *Deviation) Path() string {
	return d.path
}

func (d *Deviation) DesiredValue() string {
	return d.desiredValue
}

func (d *Deviation) ActualValue() string {
	return d.actualValue
}

func (d *Deviation) Reason() string {
	return d.reason
}

func (d *Deviation) String() string {
	var b strings.Builder
	b.WriteString("Path: ")
	b.WriteString(d.path)
	b.WriteString("\nActual Value: ")
	b.WriteString(d.actualValue)
	b.WriteString("\nReason: ")
	b.WriteString(d.reason)
	b.WriteString("\nDesired Value: ")
	b.WriteString(d.desiredValue)
	return b.String()
}

type DeviationType string

const (
	DeviationTypeUnknown DeviationType = "unknown"
	DeviationTypeTarget  DeviationType = "target"
	DeviationTypeConfig  DeviationType = "config"
)

func (d *Deviations) AddDeviation(dev *Deviation) {
	dev.SetParent(d)
	d.deviations = append(d.deviations, *dev)
}
