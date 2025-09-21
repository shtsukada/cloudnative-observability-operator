package conditions_test

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"

	. "github.com/shtsukada/cloudnative-observability-operator/internal/shared/conditions"
)

func TestEmit(t *testing.T) {
	rec := record.NewFakeRecorder(10)
	pod := &corev1.Pod{}

	var obj runtime.Object = pod

	Emit(rec, obj, EventTypeWarning, ReasonApplyFailed, "oops %d", 42)

	select {
	case e := <-rec.Events:
		if len(e) == 0 {
			t.Fatal("empty event")
		}
	default:
		t.Fatal("no event emitted")
	}
}
