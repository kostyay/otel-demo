REPO_BASE := us-central1-docker.pkg.dev/otel-demo-387216/otel-demo
TAG := v1.0

.PHONY: build
build:
	docker build -f build/Dockerfile . -t ${REPO_BASE}/controller:${TAG}


.PHONY: push
push:
	docker push ${REPO_BASE}/controller:${TAG}