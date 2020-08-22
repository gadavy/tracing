package main

import (
	"context"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/gadavy/tracing"
)

var (
	service   = ""
	jaegerURL = ""
)

func main() {
	dev, err := zap.NewProduction(zap.AddStacktrace(zap.FatalLevel))
	if err != nil {
		panic(err)
	}

	logger := tracing.DefaultTraceLogger(dev.Sugar())

	tracer, err := tracing.NewTracer(tracing.DefaultConfiguration(service, jaegerURL))
	if err != nil {
		panic(err)
	}

	example := NewExample(tracer, logger)

	fmt.Println("=============== example 1 ===============")
	example.ExampleCreateSpanBase()

	fmt.Println("\n=============== example 2 ===============")
	example.ExampleCreateSpanWithHTTP(nil, &http.Request{})
}

type Example struct {
	tracer *tracing.Tracer
	logger *tracing.TraceLogger
}

func NewExample(tracer *tracing.Tracer, logger *tracing.TraceLogger) *Example {
	return &Example{
		tracer: tracer,
		logger: logger,
	}
}

func (e *Example) ExampleCreateSpanBase() {
	span, ctx := e.tracer.NewSpan().
		WithName("ExampleCreateSpanBase").
		BuildWithContext(context.Background())
	defer span.Finish()

	e.logger.Info(ctx, "ExampleCreateSpanBase info message")

	e.ExampleChildSpanFromContext(ctx)
}

func (e *Example) ExampleCreateSpanWithHTTP(w http.ResponseWriter, r *http.Request) {
	span, ctx := e.tracer.NewSpan().
		WithName("ExampleCreateSpanWithHTTP").
		ExtractHeaders(r.Header).
		BuildWithContext(context.Background())
	defer span.Finish()

	e.logger.Info(ctx, "ExampleCreateSpanWithHTTP info message")

	e.ExampleChildSpanFromContext(ctx)

	e.ExampleChildSpanFromContext(ctx)
}

func (e *Example) ExampleChildSpanFromContext(ctx context.Context) {
	defer tracing.ChildSpan(&ctx).WithTag("key", "val").Finish()

	e.logger.Debug(ctx, "ExampleChildSpanFromContext debug message")
	e.logger.Info(ctx, "ExampleChildSpanFromContext info message")
	e.logger.Warn(ctx, "ExampleChildSpanFromContext warn message")
	e.logger.Error(ctx, "ExampleChildSpanFromContext error message")
}
