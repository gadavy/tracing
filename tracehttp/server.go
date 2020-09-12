package tracehttp

import (
	"net/http"
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

type Option func(options *Options)

func WithNameFunc(f func(r *http.Request) string) Option {
	return func(options *Options) {
		options.NameFunc = f
	}
}

func WithFilterFunc(f func(r *http.Request) bool) Option {
	return func(options *Options) {
		options.Filter = f
	}
}

type Options struct {
	NameFunc func(r *http.Request) string
	Filter   func(r *http.Request) bool
}

type Middleware struct {
	tracer  opentracing.Tracer
	options *Options
}

func NewMiddleware(tracer opentracing.Tracer, options ...Option) *Middleware {
	opts := &Options{
		NameFunc: defaultServerNameFunc,
		Filter:   func(r *http.Request) bool { return true },
	}

	for _, option := range options {
		option(opts)
	}

	return &Middleware{tracer: tracer, options: opts}
}

func (m *Middleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.options.Filter(r) {
			next.ServeHTTP(w, r)

			return
		}

		spanCtx, _ := m.tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))

		span := m.tracer.StartSpan(m.options.NameFunc(r), ext.RPCServerOption(spanCtx))
		defer span.Finish()

		ext.Component.Set(span, "net/http")
		ext.HTTPMethod.Set(span, r.Method)
		ext.HTTPUrl.Set(span, r.URL.String())

		tracker := NewStatusCodeTracker(w)

		next.ServeHTTP(tracker, r.WithContext(opentracing.ContextWithSpan(r.Context(), span)))

		ext.HTTPStatusCode.Set(span, tracker.OpentracingCode())

		if tracker.code >= http.StatusInternalServerError {
			ext.Error.Set(span, true)
		}
	})
}

func defaultServerNameFunc(r *http.Request) string {
	return strings.Join([]string{"HTTP", r.Method, r.RequestURI}, " ")
}

type StatusCodeTracker struct {
	http.ResponseWriter
	code int
}

func NewStatusCodeTracker(w http.ResponseWriter) *StatusCodeTracker {
	return &StatusCodeTracker{ResponseWriter: w, code: http.StatusOK}
}

func (t *StatusCodeTracker) WriteHeader(statusCode int) {
	t.code = statusCode
	t.ResponseWriter.WriteHeader(statusCode)
}

func (t *StatusCodeTracker) OpentracingCode() uint16 {
	return uint16(t.code)
}
