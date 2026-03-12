package deviations

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	fuzzyfinder "github.com/ktr0731/go-fuzzyfinder"
	"github.com/sdcio/config-server/apis/config/v1alpha1"
	"github.com/sdcio/kubectl-sdc/pkg/client"
	"github.com/sdcio/kubectl-sdc/pkg/types"
)

// findDeviationIndexes wraps fuzzyfinder multi-select so tests can inject selections deterministically.
var findDeviationIndexes = func(deviations []*types.Deviation, display func(i int) string, opts ...fuzzyfinder.Option) ([]int, error) {
	return fuzzyfinder.FindMulti(deviations, display, opts...)
}

var (
	ErrDeviationOrTargetNotSet        = errors.New("deviation or target not set")
	ErrNoDeviationsFound              = errors.New("no deviations found")
	ErrNoDeviationsAfterPathFiltering = errors.New("no deviations found after path filtering")
)

// DeviationClient defines the interface for deviation operations
type DeviationClient interface {
	GetDeviationByName(ctx context.Context, namespace string, deviationName string) (*types.IntentDeviations, error)
	GetDeviationsByTarget(ctx context.Context, namespace string, targetName string) (types.Deviations, error)
	ClearTargetDeviations(ctx context.Context, resource *v1alpha1.TargetClearDeviation) error
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

// getPreviewOpt adds a preview window option to the fuzzy finder
func getPreviewOpt(deviations []*types.Deviation) fuzzyfinder.Option {
	return fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
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
			alignLabel("Path:", maxLabel), deviations[i].Path,
			alignLabel("Actual:", maxLabel), deviations[i].ActualValue,
			alignLabel("Desired:", maxLabel), deviations[i].DesiredValue,
			alignLabel("Reason:", maxLabel), deviations[i].Reason,
		)
		return preview
	})
}

type deviationViewConfig struct {
	header     string
	searchItem func(i int) string
	display    func(i int) string
}

func buildDeviationViewConfig(devs types.Deviations, deviations []*types.Deviation, do *DeviationOptions, maxNameLength int) deviationViewConfig {
	if devs.MultipleIntents() {
		return deviationViewConfig{
			header: fmt.Sprintf("Namespace: %s, Target: %s", do.namespace, do.target),
			searchItem: func(i int) string {
				return fmt.Sprintf("%s%s", deviations[i].DesiredValue, deviations[i].ActualValue)
			},
			display: func(i int) string {
				return fmt.Sprintf("%-*s %s %s", maxNameLength, deviations[i].Name(), reasonInitial(deviations[i].Reason), deviations[i].Path)
			},
		}
	}

	return deviationViewConfig{
		header: fmt.Sprintf("Namespace: %s, Deviation: %s [%s]", do.namespace, devs.First().Name(), devs.First().Type()),
		searchItem: func(i int) string {
			return fmt.Sprintf("%s%s%s%s", deviations[i].Name(), deviations[i].Reason, deviations[i].DesiredValue, deviations[i].ActualValue)
		},
		display: func(i int) string {
			return fmt.Sprintf("%s %s", reasonInitial(deviations[i].Reason), deviations[i].Path)
		},
	}
}

func loadDeviations(ctx context.Context, cl DeviationClient, do *DeviationOptions) (types.Deviations, error) {
	switch {
	case do.deviationName != "":
		dev, err := cl.GetDeviationByName(ctx, do.namespace, do.deviationName)
		if err != nil {
			return nil, err
		}
		devs := types.Deviations{}
		devs.AddDeviation(dev)
		return devs, nil
	case do.target != "":
		return cl.GetDeviationsByTarget(ctx, do.namespace, do.target)
	default:
		return nil, ErrDeviationOrTargetNotSet
	}
}

func buildInteractiveFinderOptions(deviations types.DeviationSlice, allDeviations types.Deviations, do *DeviationOptions) []fuzzyfinder.Option {
	maxNameLength := allDeviations.MaxDeviationNameLength()
	viewCfg := buildDeviationViewConfig(allDeviations, deviations, do, maxNameLength)

	opts := []fuzzyfinder.Option{
		fuzzyfinder.WithHeader(viewCfg.header),
		fuzzyfinder.WithSearchItemFunc(viewCfg.searchItem),
		fuzzyfinder.WithPreviewVisible(do.Preview()),
		getPreviewOpt(deviations),
	}

	if do.InitialQuery() != "" {
		opts = append(opts, fuzzyfinder.WithQuery(do.InitialQuery()))
	}

	if len(do.SelectPathPrefix()) > 0 {
		opts = append(opts, fuzzyfinder.WithPreselected(func(i int) bool {
			for _, prefix := range do.SelectPathPrefix() {
				if prefix == "" {
					continue
				}
				if strings.HasPrefix(deviations[i].Path, prefix) {
					return true
				}
			}
			return false
		}))
		if do.AutoAcceptSelectPathPrefix() {
			opts = append(opts, fuzzyfinder.WithAutoAcceptPreselected())
		}
	}

	return opts
}

func selectDeviationsInteractive(deviations types.DeviationSlice, allDeviations types.Deviations, do *DeviationOptions) (types.Deviations, error) {
	viewCfg := buildDeviationViewConfig(allDeviations, deviations, do, allDeviations.MaxDeviationNameLength())
	opts := buildInteractiveFinderOptions(deviations, allDeviations, do)
	idxs, err := findDeviationIndexes(deviations, viewCfg.display, opts...)
	if err != nil {
		return nil, err
	}

	if len(idxs) == 0 {
		return nil, nil
	}

	return deviations.FilterByIndexes(idxs), nil
}

// Run executes the deviation selection and returns the selected deviations
func Run(ctx context.Context, cl DeviationClient, do *DeviationOptions) (types.Deviations, error) {
	var err error
	devs, err := loadDeviations(ctx, cl, do)
	if err != nil {
		return nil, err
	}

	if !devs.HasDeviations() {
		return nil, ErrNoDeviationsFound
	}

	// collect all the deviations into a single slice for fuzzy finding
	deviations := devs.Items()
	deviations = deviations.FilterByPathPrefixes(do.FilterPath())
	if len(deviations) == 0 {
		return nil, ErrNoDeviationsAfterPathFiltering
	}

	var selectedDeviations types.Deviations
	if do.Interactive() {
		selectedDeviations, err = selectDeviationsInteractive(deviations, devs, do)
		if err != nil {
			return nil, err
		}
	} else {
		selectedDeviations = deviations.ToDeviations()
	}

	if do.Revert() && selectedDeviations.HasDeviations() {
		return nil, revert(ctx, cl, do.namespace, selectedDeviations.First().Target(), selectedDeviations)
	}

	return selectedDeviations, nil
}

// revert clears the specified paths on a target
func revert(ctx context.Context, cl DeviationClient, namespace, targetName string, devs types.Deviations) error {
	if !devs.HasDeviations() {
		return fmt.Errorf("no deviations provided to clear")
	}

	return cl.ClearTargetDeviations(ctx, client.NewTargetClearDeviation(namespace, targetName, devs))
}
