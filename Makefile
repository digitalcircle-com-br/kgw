#DOCKER = nerdctl -n k8s.io
DOCKER = docker

.PHONY: kgw
kgw:
	CGO_ENABLED=false GOARCH=amd64 GOOS=linux go build -o ./deploy/kgw/main ./cmd/kgw
	$(DOCKER) build -t digitalcircle/kgw -f deploy/kgw/Dockerfile .
	$(DOCKER) push digitalcircle/kgw
#	docker build -t router -f deploy/router/Dockerfile .
