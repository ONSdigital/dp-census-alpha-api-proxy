BINPATH ?= build
AUTH_TOKEN ?=
FTB_URL ?= http://localhost:10100/v6
PORT=8080

.PHONY: build
build:
	go build -tags 'production' -o $(BINPATH)/dp-census-alpha-api-proxy

.PHONY: debug
debug:
	go build -tags 'debug' -o $(BINPATH)/dp-census-alpha-api-proxy
	HUMAN_LOG=1 DEBUG=1 AUTH_TOKEN=$(AUTH_TOKEN) FTB_URL=$(FTB_URL) $(BINPATH)/dp-census-alpha-api-proxy

.PHONY: test
test:
	go test -race -cover ./...

ping:
	curl -i -H "Authorization: Bearer abc123" "http://localhost:${PORT}/v6/datasets"

.PHONY: convey
convey:
	goconvey ./...

