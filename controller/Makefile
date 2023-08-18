REPO_BASE := us-central1-docker.pkg.dev/otel-demo-2/otel-demo
PROGRAM := controller
COMMIT            ?= $(shell git rev-parse HEAD)

LDFLAGS_VARS	  = -X github.com/kostyay/otel-demo/common/version.ServiceName=${PROGRAM} -X github.com/kostyay/otel-demo/common/version.Version=${COMMIT}
LDFLAGS           = -ldflags "-s -w ${LDFLAGS_VARS}"
TAG := v1.0

.PHONY: build
build:
	@echo "Building controller..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o app ./cmd/...

.PHONY: container
container:
	@echo "Building container..."
	@docker buildx build --platform linux/amd64 -f build/Dockerfile . -t ${REPO_BASE}/controller:${TAG}

.PHONY: push
push:
	docker push ${REPO_BASE}/controller:${TAG}

.PHONY: deploy
deploy: build container push
	gcloud run deploy controller \
    --image=${REPO_BASE}/controller:${TAG} \
    --set-cloudsql-instances=otel-demo-2:us-central1:otel-demo \
    --region=us-central1 \
    --project=otel-demo-2 \
     && gcloud run services update-traffic controller --to-latest