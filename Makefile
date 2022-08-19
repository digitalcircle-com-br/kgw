#DOCKER = nerdctl -n k8s.io
DOCKER = docker

.PHONY: kgw
kgw:
	echo $$(date) > cmd/kgw/version
	CGO_ENABLED=false GOARCH=amd64 GOOS=linux go build -o ./deploy/kgw ./cmd/kgw
	$(DOCKER) build -t digitalcircle/kgw -f deploy/Dockerfile .
	$(DOCKER) push digitalcircle/kgw
#	docker build -t router -f deploy/router/Dockerfile .
