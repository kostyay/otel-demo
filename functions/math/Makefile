REGION := us-central1
TOPIC := math-topic
ENTRY_POINT := calculateExpression

.PHONY: deploy
deploy:
	@echo "Deploying function..."
	@go mod vendor
	@gcloud functions deploy go-pubsub-function \
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
