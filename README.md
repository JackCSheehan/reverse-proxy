# Reverse Proxy
A toy config-driven reverse proxy server with simple round-robin load balancing.

The reverse proxy is written in Go and uses Python for end-to-end tests.

## Dependencies
- Go 1.19
- Python 3.11
- GNU Make

## Building and Running
To build:
```shell
$ make build
```

The binary will be written to a `build` directory along with other build and test artifacts.

To build and run the reverse proxy with a demo config:
```shell
$ make run
```

To run with a custom config YAML, simply pass the path to the config as the first argument to the binary:
```shell
./build/reverse-proxy path/to/config.yaml
```

## Development and Testing
To run Gofmt:
```shell
$ make format
```

To run the unit test suite:
```shell
$ make unit
```

Note that `make unit` generates coverage information. You can view a coverage HTML report by running:
```shell
$ make coverage
```

There is also a suite of end-to-end tests written in Python that will actually run the reverse proxy and test it with requests. Run this suite using:
```shell
$ make e2e
```

## Configuration
The reverse proxy accepts YAML-based configs describing the endpoints to create and which URLs the requests should be forwarded to. The following is an example of a config file:
```yaml
port: 8000
endpoints:
    - from: /index
      pool:
        - http://localhost:5000/index-proxied
        - http://localhost:5002/index-proxied
    - from: /home
      pool:
        - http://localhost:5001/home-proxied
```

The top-level `port` config describes which port the reverse proxy will listen for HTTP requests on.

Each endpoint has a `from` field. Each `from` path will be created as an endpoint on the reverse proxy at startup. The `pool` field contains a list of URLs to forward requests to the corresponding `from` path to. Round-robin load balancing is used to balance forwarded requests to each URL in the pool.

For example, with the above config, the reverse proxy will be started on port 8000 and will have two endpoints: `/index` and `/home`. The first request to `/index` will be forwarded to the first URL in the pool and the second request will be forwarded to the second URL in the pool. Since there are only two endpoints in the pool, the load balancer wraps back to the beginning and will forward the third request to the first URL in the pool, and so on. For endpoints with only one URL in the pool, such as `/home`, requests are simply forwarded without any load balancing applied since there are no other URLs to forward requests to.

## Metrics
Prometheus is used to gather metrics about the application at runtime. The current metrics can be viewed by requesting the `/metrics` endpoint of the reverse proxy.

