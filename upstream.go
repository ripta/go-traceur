package main

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/opentracing-contrib/go-stdlib/nethttp"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// GenerateSpan creates a new span with a nice name based on the path of the request.
func generateSpan(t opentracing.Tracer, path string) opentracing.Span {
	name := strings.Replace(path, "/", "-", -1)
	if path == "/" {
		name = "ROOT"
	}

	span := t.StartSpan("upstream-" + name)
	ext.SpanKindRPCClient.Set(span)
	return span
}

// UpstreamRequest performs a traced request to localhost:8080 with a specific
// path and rawQuery string. The body from the upstream is passed along as-is.
func upstreamRequest(ctx context.Context, t opentracing.Tracer, path, rawQuery string) (string, error) {
	span := generateSpan(t, path)
	defer span.Finish()

	url := "http://localhost:8080" + path
	if rawQuery != "" {
		url += "?" + rawQuery
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		ext.Error.Set(span, true)
		return "", err
	}

	// sheesh...
	req, nht := nethttp.TraceRequest(t, req.WithContext(opentracing.ContextWithSpan(ctx, span)))
	defer nht.Finish()

	c := &http.Client{Transport: &nethttp.Transport{}}
	res, err := c.Do(req)
	if err != nil {
		ext.Error.Set(span, true)
		return "", err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		ext.Error.Set(span, true)
		return "", err
	}
	return string(body), nil
}
