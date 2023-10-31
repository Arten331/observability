package tracer

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

type JaegerTransport int

const (
	TransportHTTP JaegerTransport = iota + 1
	TransportAgentUDP
)

type JaegerTracerOptions struct {
	Sampler   sdktrace.Sampler
	Transport JaegerTransport
	Host      string
	Port      string
}

func SetupJaegerTracerProviderHTTP(
	ctx context.Context,
	o JaegerTracerOptions,
	namespace, serviceName, env string,
) error {
	if namespace != "" {
		serviceName = namespace + "." + serviceName
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return err
	}

	var exporter sdktrace.SpanExporter

	// Set up a trace exporter
	switch o.Transport {
	case TransportHTTP:
		exporter, err = jaeger.New(
			jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(fmt.Sprintf("http://%s:%s/api/traces", o.Host, o.Port))),
		)
	case TransportAgentUDP:
		exporter, err = jaeger.New(
			jaeger.WithAgentEndpoint(jaeger.WithAgentHost(o.Host), jaeger.WithAgentPort(o.Port)),
		)
	}

	if err != nil {
		return errors.New("failed to create trace exporter")
	}

	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.
	bsp := sdktrace.NewBatchSpanProcessor(exporter)
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
