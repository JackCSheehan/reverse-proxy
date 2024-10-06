$(shell mkdir -p build)

SRCS=$(wildcard src/*.go src/common/*.go src/proxy/*.go)

format:
	cd src && gofmt -s -w .

.PHONY: build
build: build/reverse-proxy
build/reverse-proxy: $(SRCS)
	cd src && go build -o ../build/reverse-proxy

# Helper target to build and run with a known-good test config
run: build
	./build/reverse-proxy e2e/configs/test_bad_gateway.yaml

# Run unit tests
.PHONY: unit
unit:
	cd src && go test -v -coverprofile coverage.out ./...

# Open unit test code coverage report
.PHONY: coverage
coverage:
	cd src && go tool cover -html=coverage.out

.PHONY: venv
venv: build/venv/bin/activate

# Create a Python venv and install libs
build/venv/bin/activate: e2e/requirements.txt
	python3 -m venv build/venv && build/venv/bin/pip install -r e2e/requirements.txt

ACTIVATE_VENV=. build/venv/bin/activate

# Run e2e tests
.PHONY: e2e
e2e: build venv
	$(ACTIVATE_VENV) && pytest e2e

.PHONY: clean
clean:
	rm -rf build

