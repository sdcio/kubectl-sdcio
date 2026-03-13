package deviations

import (
	"context"
	"errors"
	"testing"

	fuzzyfinder "github.com/ktr0731/go-fuzzyfinder"
	v1alpha1 "github.com/sdcio/config-server/apis/config/v1alpha1"
	mockdeviations "github.com/sdcio/kubectl-sdc/mocks/deviations"
	"github.com/sdcio/kubectl-sdc/pkg/types"
	"go.uber.org/mock/gomock"
)

func TestRun_ByTargetReturnsSelectedDeviations(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cl := mockdeviations.NewMockDeviationClient(ctrl)
	devs := newTestDeviations()

	cl.EXPECT().GetDeviationsByTarget(gomock.Any(), "default", "target-1").Return(devs, nil)

	restore := stubFindDeviationIndexes(func(deviations []*types.Deviation, display func(i int) string, opts ...interface{}) ([]int, error) {
		if len(deviations) != 2 {
			t.Fatalf("deviations length = %d, want 2", len(deviations))
		}
		if got := display(1); got == "" {
			t.Fatal("display output is empty")
		}
		return []int{1}, nil
	})
	defer restore()

	selected, err := Run(context.Background(), cl, NewDeviationOptions("default", WithTarget("target-1")))
	if err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}
	if selected == nil || !selected.HasDeviations() {
		t.Fatal("Run() returned no selected deviations")
	}
	if got := selected.First().DeviationPaths(); len(got) != 2 {
		t.Fatalf("selected paths length = %d, want 2", len(got))
	}
}

func TestRun_ByTargetInteractiveReturnsSelectedDeviations(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cl := mockdeviations.NewMockDeviationClient(ctrl)
	devs := newTestDeviations()

	cl.EXPECT().GetDeviationsByTarget(gomock.Any(), "default", "target-1").Return(devs, nil)

	restore := stubFindDeviationIndexes(func(deviations []*types.Deviation, display func(i int) string, opts ...interface{}) ([]int, error) {
		if len(deviations) != 2 {
			t.Fatalf("deviations length = %d, want 2", len(deviations))
		}
		if got := display(1); got == "" {
			t.Fatal("display output is empty")
		}
		return []int{1}, nil
	})
	defer restore()

	selected, err := Run(context.Background(), cl, NewDeviationOptions("default", WithTarget("target-1"), WithInteractive(true)))
	if err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}
	if selected == nil || !selected.HasDeviations() {
		t.Fatal("Run() returned no selected deviations")
	}
	if got := selected.First().DeviationPaths(); len(got) != 1 || got[0] != "/system/location" {
		t.Fatalf("selected paths = %v, want [/system/location]", got)
	}
}

func TestRun_ByDeviationNameReturnsSelectedDeviations(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cl := mockdeviations.NewMockDeviationClient(ctrl)
	intent := newTestIntentDeviations()

	cl.EXPECT().GetDeviationByName(gomock.Any(), "default", "dev-1").Return(intent, nil)

	restore := stubFindDeviationIndexes(func(deviations []*types.Deviation, display func(i int) string, opts ...interface{}) ([]int, error) {
		return []int{0}, nil
	})
	defer restore()

	selected, err := Run(context.Background(), cl, NewDeviationOptions("default", WithDeviationName("dev-1"), WithInteractive(true)))
	if err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}
	if got := selected.First().DeviationPaths(); len(got) != 1 || got[0] != "/system/name" {
		t.Fatalf("selected paths = %v, want [/system/name]", got)
	}
}

func TestRun_NoSelectionReturnsNilInteractive(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cl := mockdeviations.NewMockDeviationClient(ctrl)
	cl.EXPECT().GetDeviationsByTarget(gomock.Any(), "default", "target-1").Return(newTestDeviations(), nil)

	restore := stubFindDeviationIndexes(func(deviations []*types.Deviation, display func(i int) string, opts ...interface{}) ([]int, error) {
		return []int{}, nil
	})
	defer restore()

	selected, err := Run(context.Background(), cl, NewDeviationOptions("default", WithTarget("target-1"), WithInteractive(true)))
	if err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}
	if selected != nil {
		t.Fatalf("selected = %v, want nil", selected)
	}
}

func TestNewDeviationOptions_AutoAcceptSelectPathPrefix(t *testing.T) {
	do := NewDeviationOptions("default", WithAutoAcceptSelectPathPrefix(true))
	if !do.AutoAcceptSelectPathPrefix() {
		t.Fatal("AutoAcceptSelectPathPrefix() = false, want true")
	}
}

func TestNewDeviationOptions_PreselectSlice(t *testing.T) {
	want := []string{"/system/name", "/system/location"}
	do := NewDeviationOptions("default", WithSelectPathPrefix(want))
	if len(do.SelectPathPrefix()) != len(want) {
		t.Fatalf("SelectPathPrefix length = %d, want %d", len(do.SelectPathPrefix()), len(want))
	}
	for i := range want {
		if do.SelectPathPrefix()[i] != want[i] {
			t.Fatalf("SelectPathPrefix[%d] = %q, want %q", i, do.SelectPathPrefix()[i], want[i])
		}
	}
}

func TestNewDeviationOptions_FilterPathSlice(t *testing.T) {
	want := []string{"/system/name", "/system/location"}
	do := NewDeviationOptions("default", WithFilterPath(want))
	if len(do.FilterPath()) != len(want) {
		t.Fatalf("FilterPath length = %d, want %d", len(do.FilterPath()), len(want))
	}
}

func TestRun_AutoAcceptSelectPathPrefixOptionWiring(t *testing.T) {
	tests := []struct {
		name       string
		options    []DeviationOptionSetter
		wantOptLen int
	}{
		{
			name: "disabled with preselect",
			options: []DeviationOptionSetter{
				WithTarget("target-1"),
				WithInteractive(true),
				WithSelectPathPrefix([]string{"/system"}),
			},
			wantOptLen: 5,
		},
		{
			name: "enabled without preselect",
			options: []DeviationOptionSetter{
				WithTarget("target-1"),
				WithInteractive(true),
				WithAutoAcceptSelectPathPrefix(true),
			},
			wantOptLen: 4,
		},
		{
			name: "enabled with preselect",
			options: []DeviationOptionSetter{
				WithTarget("target-1"),
				WithInteractive(true),
				WithSelectPathPrefix([]string{"/system"}),
				WithAutoAcceptSelectPathPrefix(true),
			},
			wantOptLen: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			cl := mockdeviations.NewMockDeviationClient(ctrl)
			cl.EXPECT().GetDeviationsByTarget(gomock.Any(), "default", "target-1").Return(newTestDeviations(), nil)

			restore := stubFindDeviationIndexes(func(deviations []*types.Deviation, display func(i int) string, opts ...interface{}) ([]int, error) {
				if got := len(opts); got != tt.wantOptLen {
					t.Fatalf("options length = %d, want %d", got, tt.wantOptLen)
				}
				return []int{0}, nil
			})
			defer restore()

			selected, err := Run(context.Background(), cl, NewDeviationOptions("default", tt.options...))
			if err != nil {
				t.Fatalf("Run() unexpected error: %v", err)
			}
			if selected == nil || !selected.HasDeviations() {
				t.Fatal("Run() returned no selected deviations")
			}
		})
	}
}

func TestRun_NonInteractiveFilterPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cl := mockdeviations.NewMockDeviationClient(ctrl)
	cl.EXPECT().GetDeviationsByTarget(gomock.Any(), "default", "target-1").Return(newTestDeviations(), nil)

	selected, err := Run(context.Background(), cl, NewDeviationOptions("default", WithTarget("target-1"), WithFilterPath([]string{"/system/name"})))
	if err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}
	if selected == nil || !selected.HasDeviations() {
		t.Fatal("Run() returned no selected deviations")
	}
	if got := selected.First().DeviationPaths(); len(got) != 1 || got[0] != "/system/name" {
		t.Fatalf("selected paths = %v, want [/system/name]", got)
	}
}

func TestRun_FilterPathNoMatchesReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cl := mockdeviations.NewMockDeviationClient(ctrl)
	cl.EXPECT().GetDeviationsByTarget(gomock.Any(), "default", "target-1").Return(newTestDeviations(), nil)

	selected, err := Run(context.Background(), cl, NewDeviationOptions("default", WithTarget("target-1"), WithFilterPath([]string{"/interfaces"})))
	if !errors.Is(err, ErrNoDeviationsAfterPathFiltering) {
		t.Fatalf("Run() error = %v, want errors.Is(..., ErrNoDeviationsAfterPathFiltering)", err)
	}
	if selected != nil {
		t.Fatalf("selected = %v, want nil", selected)
	}
}

func TestRun_RevertClearsSelectedDeviations(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cl := mockdeviations.NewMockDeviationClient(ctrl)
	cl.EXPECT().GetDeviationsByTarget(gomock.Any(), "default", "target-1").Return(newTestDeviations(), nil)
	cl.EXPECT().ClearTargetDeviations(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, resource *v1alpha1.TargetClearDeviation) error {
		if resource.Name != "target-1" {
			return errors.New("unexpected target name")
		}
		if resource.Namespace != "default" {
			return errors.New("unexpected namespace")
		}
		if len(resource.Spec.Config) != 1 || len(resource.Spec.Config[0].Paths) != 1 || resource.Spec.Config[0].Paths[0] != "/system/name" {
			return errors.New("unexpected resource payload")
		}
		return nil
	})

	restore := stubFindDeviationIndexes(func(deviations []*types.Deviation, display func(i int) string, opts ...interface{}) ([]int, error) {
		return []int{0}, nil
	})
	defer restore()

	selected, err := Run(context.Background(), cl, NewDeviationOptions("default", WithTarget("target-1"), WithInteractive(true), WithRevert(true)))
	if err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}
	if selected != nil {
		t.Fatalf("selected = %v, want nil after revert", selected)
	}
}

func TestRun_NoTargetOrDeviationReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cl := mockdeviations.NewMockDeviationClient(ctrl)
	selected, err := Run(context.Background(), cl, NewDeviationOptions("default"))
	if !errors.Is(err, ErrDeviationOrTargetNotSet) {
		t.Fatalf("Run() error = %v, want errors.Is(..., ErrDeviationOrTargetNotSet)", err)
	}
	if selected != nil {
		t.Fatalf("selected = %v, want nil", selected)
	}
}

func TestRun_NoDeviationsFoundReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cl := mockdeviations.NewMockDeviationClient(ctrl)
	cl.EXPECT().GetDeviationsByTarget(gomock.Any(), "default", "target-1").Return(types.Deviations{}, nil)

	selected, err := Run(context.Background(), cl, NewDeviationOptions("default", WithTarget("target-1")))
	if !errors.Is(err, ErrNoDeviationsFound) {
		t.Fatalf("Run() error = %v, want errors.Is(..., ErrNoDeviationsFound)", err)
	}
	if selected != nil {
		t.Fatalf("selected = %v, want nil", selected)
	}
}

func TestRevert_WithoutDeviationsReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cl := mockdeviations.NewMockDeviationClient(ctrl)
	err := revert(context.Background(), cl, "default", "target-1", types.Deviations{})
	if err == nil || err.Error() != "no deviations provided to clear" {
		t.Fatalf("revert() error = %v, want %q", err, "no deviations provided to clear")
	}
}

func stubFindDeviationIndexes(stub func(deviations []*types.Deviation, display func(i int) string, opts ...interface{}) ([]int, error)) func() {
	original := findDeviationIndexes
	findDeviationIndexes = func(deviations []*types.Deviation, display func(i int) string, opts ...fuzzyfinder.Option) ([]int, error) {
		genericOpts := make([]interface{}, len(opts))
		for i, opt := range opts {
			genericOpts[i] = opt
		}
		return stub(deviations, display, genericOpts...)
	}
	return func() {
		findDeviationIndexes = original
	}
}

func newTestDeviations() types.Deviations {
	devs := types.Deviations{}
	devs.AddDeviation(newTestIntentDeviations())
	return devs
}

func newTestIntentDeviations() *types.IntentDeviations {
	intent := types.NewDeviations("target-1", "dev-1", types.DeviationTypeConfig, 2).SetNamespace("default")
	intent.AddDeviation(types.NewDeviation("/system/name", "router-1", "router-2", "mismatch"))
	intent.AddDeviation(types.NewDeviation("/system/location", "lab-1", "lab-2", "mismatch"))
	return intent
}
