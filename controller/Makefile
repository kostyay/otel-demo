REPO_BASE := us-central1-docker.pkg.dev/otel-demo-387216/otel-demo
TAG := v1.0

.PHONY: build
build:
	@echo "Building controller..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app ./cmd/...

.PHONY: container
container:
	@echo "Building container..."
	@docker buildx build --platform linux/amd64 -f build/Dockerfile . -t ${REPO_BASE}/controller:${TAG}

.PHONY: push
push:
	docker push ${REPO_BASE}/controller:${TAG}

.PHONY: deploy
deploy:
	gcloud run deploy controller \
    --image=us-central1-docker.pkg.dev/otel-demo-387216/otel-demo/controller:${TAG} \
    --set-cloudsql-instances=otel-demo-387216:us-central1:otel-demo \
    --region=us-central1 \
    --project=otel-demo-387216 \
     && gcloud run services update-traffic controller --to-latest