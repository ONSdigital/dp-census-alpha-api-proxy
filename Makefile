BINPATH ?= build

BIND_ADDR=10100
FTB_URL=http://localhost:8491

containerName=alpha-api-proxy
binary-name=dp-census-alpha-api-proxy

.PHONY: build
build:
	go build -tags 'production' -o $(BINPATH)/${binary-name}

.PHONY: debug
debug:
	go build -tags 'debug' -o $(BINPATH)/${binary-name}
	HUMAN_LOG=1 DEBUG=1 BIND_ADDR=:$(BIND_ADDR) AUTH_TOKEN=$(AUTH_PROXY_TOKEN) FTB_URL=$(FTB_URL) $(BINPATH)/${binary-name}

.PHONY: ping
ping:
	curl -i -H "Authorization: Bearer ${AUTH_PROXY_TOKEN}" "http://localhost:${BIND_ADDR}/v6/datasets"

.PHONY: container
container:
	@echo "stopping ${containerName} container"
	docker stop ${containerName} || true

	@echo "removing ${containerName} container"
	docker rm ${containerName} || true

	@echo "removing ${containerName} image"
	docker rmi ${containerName} || true

	@echo "building ${binary-name}-linux binary"
	env GOOS=linux GOARCH=amd64 go build -o $(BINPATH)/${binary-name}-linux

	@echo "building ${containerName}  container"
	docker build -t ${containerName} -f Dockerfile.ec2 \
		--build-arg BIND_ADDR=${BIND_ADDR} \
		--build-arg AUTH_TOKEN=${AUTH_PROXY_TOKEN} \
		--build-arg FTB_URL=${FTB_URL} .

.PHONY: docker
docker: container
	docker run -it \
		--name ${containerName} \
		-p 0.0.0.0:${BIND_ADDR}:${BIND_ADDR} \
		${containerName}

