
# Experimenting with OpenTracing in Go

**These are implementation-specific notes.**

Run zipkin:

```
docker run --rm -p 9411:9411 -e JAVA_OPTS=-Dlogging.level.zipkin=DEBUG openzipkin/zipkin:2.7.2
```

Then build and run go-traceur:

```
go install github.com/ripta/go-traceur
go-traceur
```

And then do a request:

```
curl -i localhost:8080/hello/world/echo?foo=bar
```

which will generate recursive requests to:

```
/hello/world/echo?foo=bar
/world/echo?foo=bar
/echo?foo=bar
```

The output of `go-traceur` should include request logs and dumps of the request
headers. The headers should look like:

```
2018/04/24 00:31:04 --> {"Accept-Encoding":["gzip"],"Uber-Trace-Id":["79939e6527574fb:519a5fc30eb16a20:126014b42a7845c9:1"],"User-Agent":["Go-http-client/1.1"]}
```

Specifically, you'll want the `Uber-Trace-Id`, which consists of four colon-separated segments:

- the trace ID, which is globally unique;
- the span ID, which is unique within the trace and equal to the trace ID at the root;
- the span ID of the parent span, which is zero at the root; and
- a bitmap flag: sampled(1) or debug(2).

The trace ID is the ID that can be looked up in Zipkin.
