package tracer

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

type JaegerTracerOptions struct {
	Sampler sdktrace.Sampler
	Host    string
	Port    string
}

func SetupJaegerTracerProvider(ctx context.Context, o JaegerTracerOptions, serviceName string) error {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return err
	}

	// Set up a trace exporter
	traceExporter, err := jaeger.New(
		jaeger.WithAgentEndpoint(jaeger.WithAgentHost(o.Host), jaeger.WithAgentPort(o.Port)),
	)
	if err != nil {
		return errors.New("failed to create trace exporter")
	}

	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.
	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(o.Sampler),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)

	// set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	// setup global tracer for service
	SetupGlobalTracer(otel.Tracer(serviceName), serviceName)

	return nil
}
