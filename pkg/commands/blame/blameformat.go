package blame

import (
	"fmt"
	"strings"
)

type BlameFormat string

const (
	BlameFormatTree  BlameFormat = "tree"
	BlameFormatXPath BlameFormat = "xpath"
)

var BlameFormatOptions = []BlameFormat{BlameFormatTree, BlameFormatXPath}

// FormatOptionsString returns a formatted string of all available format options
func FormatOptionsString() string {
	opts := make([]string, len(BlameFormatOptions))
	for i, opt := range BlameFormatOptions {
		opts[i] = string(opt)
	}
	return strings.Join(opts, " or ")
}

// ParseFormat parses a string into a BlameFormat
func ParseFormat(s string) (BlameFormat, error) {
	switch strings.ToLower(s) {
	case "tree":
		return BlameFormatTree, nil
	case "xpath":
		return BlameFormatXPath, nil
	default:
		return "", fmt.Errorf("invalid format: %q (must be one of %v)", s, BlameFormatOptions)
	}
}
