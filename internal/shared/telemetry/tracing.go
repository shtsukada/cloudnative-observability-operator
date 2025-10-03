package telemetry

import (
	"context"
	"fmt"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

func envBool(key string, def bool) bool {
	if v, ok := os.LookupEnv(key); ok {
		v = strings.ToLower(strings.TrimSpace(v))
		return v == "1" || v == "true" || v == "yes"
	}
	return def
}

func envStr(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && strings.TrimSpace(v) != "" {
		return v
	}
	return def
}

func envFloat(key string, def float64) float64 {
	if v, ok := os.LookupEnv(key); ok {
		var f float64
		if _, err := fmt.Sscan(v, &f); err == nil {
			return f
		}
	}
	return def
}

func stripScheme(u string) string {
	u = strings.TrimPrefix(u, "http://")
	u = strings.TrimPrefix(u, "https://")
	return u
}

func InitTracer(ctx context.Context) (func(context.Context) error, error) {
	if !envBool("OTEL_TRACES_ENABLED", false) {
		otel.SetTracerProvider(sdktrace.NewTracerProvider())
		otel.SetTextMapPropagator(propagation.TraceContext{})
		return func(ctx context.Context) error { return nil }, nil
	}

	endpoint := envStr("OTEL_EXPORTER_OTLP_ENDPOINT", "otel-collector.monitoring:4317")
	exp, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpoint(stripScheme(endpoint)),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	sampler := envStr("OTEL_TRACES_SAMPLER", "parentbased_traceidratio")
	arg := envFloat("OTEL_TRACES_SAMPLER_ARG", 0.1)

	var s sdktrace.Sampler = sdktrace.ParentBased(sdktrace.TraceIDRatioBased(arg))
	switch sampler {
	case "always_on":
		s = sdktrace.AlwaysSample()
	case "always_off":
		s = sdktrace.NeverSample()
	case "traceidratio":
		s = sdktrace.TraceIDRatioBased(arg)
	case "parentbased_traceidratio":
	}

	res, _ := sdkresource.Merge(
		sdkresource.Default(),
		sdkresource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(envStr("OTEL_SERVICE_NAME", "cloudnative-observability-operator")),
		),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(s),
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	return tp.Shutdown, nil
}
