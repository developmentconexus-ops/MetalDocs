package taxonomy

import (
	"context"
	"testing"
)

type stubTemplateChecker struct{}

func (stubTemplateChecker) IsPublished(_ context.Context, _ string) (bool, string, error) {
	return false, "", nil
}

func TestNew_UsesDefaultTemplateCheckerWhenDependencyIsNil(t *testing.T) {
	mod := New(Dependencies{TplChecker: stubTemplateChecker{}})
	if mod == nil || mod.Handler == nil {
		t.Fatal("expected module handler to be initialized")
	}
}
