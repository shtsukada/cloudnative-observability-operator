package telemetry

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type tracingReconciler struct {
	kind   string
	inner  reconcile.Reconciler
	tracer trace.Tracer
}

func WrapReconciler(kind string, inner reconcile.Reconciler) reconcile.Reconciler {
	return &tracingReconciler{
		kind:   kind,
		inner:  inner,
		tracer: otel.Tracer("cno-operator"),
	}
}

func (t *tracingReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	start := time.Now()
	ctx, span := t.tracer.Start(ctx, "Reconcile "+t.kind)
	defer span.End()

	res, err := t.inner.Reconcile(ctx, req)

	result := "success"
	if err != nil {
		result = "error"
	} else if res.RequeueAfter > 0 {
		result = "requeue"
	}

	ObserveReconcile(t.kind, result, time.Since(start))
	span.SetAttributes(
		attribute.String("k8s.kind", t.kind),
		attribute.String("k8s.namespace", req.Namespace),
		attribute.String("k8s.name", req.Name),
		attribute.String("result", result),
	)
	return res, err
}
