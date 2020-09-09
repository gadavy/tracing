package tracing

import (
	"context"
	"testing"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestTrace_ContextWithAction(t *testing.T) {
	tests := []struct {
		name     string
		ctx      func() context.Context
		expected string
	}{
		{
			name: "#1",
			ctx: func() context.Context {
				return opentracing.ContextWithSpan(context.TODO(), mocktracer.New().StartSpan("test"))
			},
			expected: "43",
		},
		{
			name: "#2",
			ctx: func() context.Context {
				return context.Background()
			},
			expected: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := &TraceLogger{
				traceContextName: "mockpfx-ids-traceid",
			}

			assert.Equal(t, tt.expected, logger.GetTraceID(tt.ctx()))
		})
	}
}

func TestTraceLogger_withErrorTag(t *testing.T) {
	tests := []struct {
		name    string
		span    opentracing.Span
		withTag bool
	}{
		{
			name:    "WithSpan",
			span:    mocktracer.New().StartSpan("test"),
			withTag: true,
		},
		{
			name:    "WithoutSpan",
			span:    nil,
			withTag: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := opentracing.ContextWithSpan(context.TODO(), tt.span)

			withErrorTag(ctx)

			var result bool

			if tt.withTag {
				mockSpan, ok := tt.span.(*mocktracer.MockSpan)
				if !ok {
					t.Error("expected mocktracer.MockSpan")
				}

				result, ok = mockSpan.Tag("error").(bool)
				if !ok {
					t.Error("expected bool tag value")
				}
			}

			assert.Equal(t, tt.withTag, result)
		})
	}
}

var js = []byte(
	`
	{
		"key": "value",
		"v2": {
			"k2": 123,
			"bb": ["cc", "dd"]
		}
	}
	`,
)

func TestTraceLogger(t *testing.T) {
	l, err := zap.NewDevelopment()
	if err != nil {
		t.Error(err)
	}

	s := struct {
		Key string
	}{
		Key: "AAAAA!!!!",
	}

	logger := NewTraceLogger("key", "mockpfx-ids-traceid", l.Sugar())

	logger.With("kkkk", "vvvv").Error(context.Background(), "123")

	ctx := context.Background()

	logger.WithJSON("json", js).Error(ctx, "123333")

	logger.With("kkkk2", "vvvv2", "json-v2", s).Error(ctx, "123")

	logger.Error(ctx, "non kkkk")
}
