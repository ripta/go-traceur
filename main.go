package main

import (
	"io"
	"log"

	"github.com/opentracing/opentracing-go"

	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/transport/zipkin"
)

func main() {
	run() // so defer works
}

func newTracer() (opentracing.Tracer, io.Closer, error) {
	tp, err := zipkin.NewHTTPTransport(
		"http://localhost:9411/api/v1/spans", // v2 doesn't seem to work?
		zipkin.HTTPBatchSize(1),
		zipkin.HTTPLogger(jaeger.StdLogger),
	)
	if err != nil {
		return nil, nil, err
	}

	t, c := jaeger.NewTracer(
		"traceur-server",
		jaeger.NewConstSampler(true), // sample 100%
		jaeger.NewRemoteReporter(tp),
	)
	return t, c, nil
}

func run() {
	t, c, err := newTracer()
	if err != nil {
		log.Fatalf("Error initializing tracer: %v", err)
	}

	opentracing.SetGlobalTracer(t)
	defer c.Close()

	if err := serve(); err != nil {
		log.Fatalf("Error serving: %v", err)
	}
}
