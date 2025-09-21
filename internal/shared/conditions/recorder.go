package conditions

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
)

func Emit(rec record.EventRecorder, obj runtime.Object, eventType, reason, msg string, args ...any) {
	if rec == nil || obj == nil {
		return
	}
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	rec.Event(obj, eventType, reason, msg)
}
