package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sdcio/kubectl-sdc/pkg/client"
	"github.com/sdcio/kubectl-sdc/pkg/types"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

type deviationOutputFormat string

const (
	deviationOutputFormatText         deviationOutputFormat = "text"
	deviationOutputFormatResourceYAML deviationOutputFormat = "resource-yaml"
	deviationOutputFormatResourceJSON deviationOutputFormat = "resource-json"
)

var deviationOutputFormats = []deviationOutputFormat{
	deviationOutputFormatText,
	deviationOutputFormatResourceYAML,
	deviationOutputFormatResourceJSON,
}

func deviationOutputFormatStrings() []string {
	formats := make([]string, len(deviationOutputFormats))
	for i, format := range deviationOutputFormats {
		formats[i] = string(format)
	}
	return formats
}

func deviationOutputFormatListString() string {
	return strings.Join(deviationOutputFormatStrings(), ", ")
}

func parseDeviationOutputFormat(format string) (deviationOutputFormat, error) {
	switch deviationOutputFormat(strings.ToLower(format)) {
	case deviationOutputFormatText:
		return deviationOutputFormatText, nil
	case deviationOutputFormatResourceYAML:
		return deviationOutputFormatResourceYAML, nil
	case deviationOutputFormatResourceJSON:
		return deviationOutputFormatResourceJSON, nil
	default:
		return "", fmt.Errorf("invalid format %q, must be one of: %s", format, deviationOutputFormatListString())
	}
}

func deviationFormatCompletionFunc() func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return deviationOutputFormatStrings(), cobra.ShellCompDirectiveNoFileComp
	}
}

func formatSelectedDeviations(devs types.Deviations, format deviationOutputFormat) (string, error) {
	if devs == nil || !devs.HasDeviations() {
		return "", nil
	}

	switch format {
	case deviationOutputFormatText:
		return strings.TrimSpace(devs.String()), nil
	case deviationOutputFormatResourceYAML:
		resource, err := selectedDeviationsResource(devs)
		if err != nil {
			return "", err
		}
		data, err := yaml.Marshal(resource)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(data)), nil
	case deviationOutputFormatResourceJSON:
		resource, err := selectedDeviationsResource(devs)
		if err != nil {
			return "", err
		}
		data, err := json.MarshalIndent(resource, "", "  ")
		if err != nil {
			return "", err
		}
		return string(data), nil
	default:
		return "", fmt.Errorf("unsupported output format %q", format)
	}
}

func selectedDeviationsResource(devs types.Deviations) (interface{}, error) {
	first := devs.First()
	if first == nil {
		return nil, fmt.Errorf("no deviations selected")
	}
	return client.NewTargetClearDeviation(first.Namespace(), first.Target(), devs), nil
}
