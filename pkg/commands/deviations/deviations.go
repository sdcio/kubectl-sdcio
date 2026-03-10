package deviations

import (
	"context"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	fuzzyfinder "github.com/ktr0731/go-fuzzyfinder"
	"github.com/sdcio/kubectl-sdcio/pkg/types"
)

// DeviationClient defines the interface for deviation operations
type DeviationClient interface {
	GetDeviations(ctx context.Context, namespace string, deviationName string) (*types.Deviations, error)
	ClearTargetDeviations(ctx context.Context, namespace, targetName, configName string, paths []string) error
}

// reasonInitial returns the first uppercase character of the reason in brackets
func reasonInitial(reason string) string {
	if reason == "" {
		return "[?]"
	}
	r, _ := utf8.DecodeRuneInString(reason)
	if r == utf8.RuneError {
		return "[?]"
	}
	return fmt.Sprintf("[%s]", string(unicode.ToUpper(r)))
}

// alignLabel pads a label with spaces to the specified width
func alignLabel(label string, width int) string {
	if width <= len(label) {
		return label
	}
	return label + strings.Repeat(" ", width-len(label))
}

// addPreviewOpt adds a preview window option to the fuzzy finder
func addPreviewOpt(opts []fuzzyfinder.Option, deviations []types.Deviation) []fuzzyfinder.Option {
	opts = append(opts, fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
		if i == -1 {
			return ""
		}
		// figure out indentation for each label by finding the longest label and adding 1 space
		labels := []string{"Path:", "Actual:", "Desired:", "Reason:"}
		maxLabel := 0
		for _, label := range labels {
			if len(label) > maxLabel {
				maxLabel = len(label)
			}
		}

		// build preview string with aligned labels
		preview := fmt.Sprintf(
			"%s %s\n%s %s\n%s %s\n%s %s\n",
			alignLabel("Path:", maxLabel), deviations[i].Path(),
			alignLabel("Actual:", maxLabel), deviations[i].ActualValue(),
			alignLabel("Desired:", maxLabel), deviations[i].DesiredValue(),
			alignLabel("Reason:", maxLabel), deviations[i].Reason(),
		)
		return preview
	}))
	return opts
}

// Run executes the deviation selection and returns the selected paths
func Run(ctx context.Context, cl DeviationClient, do *DeviationOptions) ([]string, error) {
	dev, err := cl.GetDeviations(ctx, do.namespace, do.deviationName)
	if err != nil {
		return nil, err
	}

	if dev.Length() == 0 {
		return nil, fmt.Errorf("no deviations found")
	}

	deviations := dev.Deviations()

	opts := []fuzzyfinder.Option{
		fuzzyfinder.WithHeader(fmt.Sprintf("Namespace: %s, Deviation: %s [%s]", dev.Namespace(), dev.Name(), dev.Type())),
		fuzzyfinder.WithSearchItemFunc(func(i int) string {
			return fmt.Sprintf("%s%s%s", deviations[i].Reason(), deviations[i].DesiredValue(), deviations[i].ActualValue())
		}),
	}

	// add preview as an option if the flag is set
	if do.preview {
		opts = addPreviewOpt(opts, dev.Deviations())
	}

	// Use fuzzy finder with multi-select to choose deviations to display
	idxs, err := fuzzyfinder.FindMulti(
		deviations,
		func(i int) string {
			return fmt.Sprintf("%s %s", reasonInitial(deviations[i].Reason()), deviations[i].Path())
		},
		opts...,
	)
	if err != nil {
		return nil, err
	}

	// Collect selected deviation paths
	paths := make([]string, 0, len(idxs))
	for _, idx := range idxs {
		paths = append(paths, deviations[idx].Path())
	}

	if len(paths) == 0 {
		return nil, nil
	}

	// If revert is requested, clear the deviations
	if do.revert {
		return nil, revert(ctx, cl, do.namespace, dev.Target(), do.deviationName, paths)
	}

	return paths, nil
}

// revert clears the specified paths on a target
func revert(ctx context.Context, cl DeviationClient, namespace, targetName, configName string, paths []string) error {
	if len(paths) == 0 {
		return fmt.Errorf("no paths provided to clear")
	}

	return cl.ClearTargetDeviations(ctx, namespace, targetName, configName, paths)
}
