#DOCKER = nerdctl -n k8s.io
DOCKER = docker

.PHONY: kgw
kgw-local:
	echo $$(date) > cmd/kgw/version
	CGO_ENABLED=false GOOS=linux go build -o ./deploy/amd64/kgw ./cmd/kgw
	$(DOCKER) build -t digitalcircle/kgw:latest -f deploy/amd64/Dockerfile .
	$(DOCKER) push digitalcircle/kgw:latest
#	docker build -t router -f deploy/router/Dockerfile .
kgw-amd64:
	echo $$(date) > cmd/kgw/version
	CGO_ENABLED=false GOARCH=amd64 GOOS=linux go build -o ./deploy/amd64/kgw ./cmd/kgw
	$(DOCKER) build -t digitalcircle/kgw:amd64 -f deploy/amd64/Dockerfile .
	$(DOCKER) push digitalcircle/kgw:amd64
#	docker build -t router -f deploy/router/Dockerfile .
kgw-arm64:
	echo $$(date) > cmd/kgw/version
	CGO_ENABLED=false GOARCH=arm64 GOOS=linux go build -o ./deploy/arm64/kgw ./cmd/kgw
	$(DOCKER) build -t digitalcircle/kgw:arm64 -f deploy/arm64/Dockerfile .
	$(DOCKER) push digitalcircle/kgw:arm64
#	docker build -t router -f deploy/router/Dockerfile .
build: kgw-amd64 kgw-arm64