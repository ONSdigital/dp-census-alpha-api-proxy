BINPATH ?= build
AUTH_TOKEN ?=
FTB_URL ?= http://localhost:8491/v6
PORT=10100

.PHONY: build
build:
	go build -tags 'production' -o $(BINPATH)/dp-census-alpha-api-proxy

.PHONY: debug
debug:
	go build -tags 'debug' -o $(BINPATH)/dp-census-alpha-api-proxy
	HUMAN_LOG=1 DEBUG=1 PORT=$(PORT) AUTH_TOKEN=$(AUTH_TOKEN) FTB_URL=$(FTB_URL) $(BINPATH)/dp-census-alpha-api-proxy

.PHONY: test
test:
	go test -race -cover ./...

ping:
	curl -i -H "Authorization: Bearer ${AUTH_TOKEN}" "http://localhost:${PORT}/v6/datasets"

.PHONY: convey
convey:
	goconvey ./...

