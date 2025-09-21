package conditions_test

import (
	"errors"
	"testing"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"

	. "github.com/shtsukada/cloudnative-observability-operator/internal/shared/conditions"
)

func TestClassifyApplyError(t *testing.T) {
	if got := ClassifyApplyError(nil); got != ReasonApplySucceeded {
		t.Fatalf("nil => %s", got)
	}
	if got := ClassifyApplyError(apierrors.NewForbidden(schema.GroupResource{Group: "", Resource: "deployments"}, "d", errors.New("x"))); got != ReasonErrForbidden {
		t.Fatalf("forbidden => %s", got)
	}
	if got := ClassifyApplyError(apierrors.NewInvalid(schema.GroupKind{Group: "", Kind: "Deployment"}, "d", nil)); got != ReasonErrInvalid {
		t.Fatalf("invalid => %s", got)
	}
	if got := ClassifyApplyError(errors.New("random")); got != ReasonErrUnknown {
		t.Fatalf("unknown => %s", got)
	}
}
