REGION := us-central1
TOPIC := math-topic
ENTRY_POINT := calculateExpression
GOOGLE_CLOUD_PROJECT := otel-demo-387216

.PHONY: deploy
deploy:
	@echo "Deploying function..."
	@go mod vendor
	@gcloud functions deploy go-pubsub-function \
	--set-env-vars GOOGLE_CLOUD_PROJECT=${GOOGLE_CLOUD_PROJECT} \
	--gen2 \
	--runtime=go120 \
	--region=${REGION} \
	--source=. \
	--entry-point=${ENTRY_POINT} \
	--trigger-topic=${TOPIC}

.PHONY: clean
clean:
	@echo "Cleaning up..."
	@rm -rf vendor
	@echo "Done."
