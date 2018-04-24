package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	opentracing "github.com/opentracing/opentracing-go"
)

func serve(t opentracing.Tracer) error {
	http.HandleFunc("/echo", echoer)
	http.HandleFunc("/", recurse(t))
	return http.ListenAndServe(":8080", nethttp.Middleware(t, http.DefaultServeMux))
}

func echoer(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)
	log.Printf("--> %s", headerAsJSON(r.Header))
	io.WriteString(w, fmt.Sprintf(r.URL.RawQuery)+"\n")
}

func recurse(t opentracing.Tracer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		log.Printf("--> %s", headerAsJSON(r.Header))

		if r.URL.Path == "/" {
			io.WriteString(w, time.Now().Format(time.RFC3339Nano)+"\n")
			return
		}

		next := nextURL(r.URL.Path)
		// verify, in case of wrong nextURL algorithm
		if len(next) >= len(r.URL.Path) {
			io.WriteString(w, fmt.Sprintf("Invalid: Request (%q) and next url (%q) match\n", r.URL.Path, next))
			return
		}

		body, err := upstreamRequest(r.Context(), t, next, r.URL.RawQuery)
		if err != nil {
			io.WriteString(w, fmt.Sprintf("Error requesting upstream: %v\n", err))
			return
		}

		io.WriteString(w, body)
	}
}

func nextURL(path string) string {
	if len(path) < 2 {
		return "/"
	}
	if idx := strings.Index(path[1:], "/"); idx >= 0 {
		return path[idx+1:]
	}
	return "/"
}

func headerAsJSON(h http.Header) string {
	s, _ := json.Marshal(h)
	return string(s)
}
