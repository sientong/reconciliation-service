IMAGE_NAME=reconciliation-service
CSV_DIR=$(PWD)/csv

build:
	docker build -t $(IMAGE_NAME) .

run:
	docker run --rm \
		-v $(CSV_DIR):/app/csv \
		$(IMAGE_NAME) \
		csv/system_transactions.csv \
		csv/bankA_20250605_large.csv,csv/bankB_20250605_large.csv \
		20250604 \
		20250610

run-custom:
	docker run --rm \
		-v $(CSV_DIR):/app/csv \
		$(IMAGE_NAME) \
		$(ARGS)

logs:
	docker logs -f $(CONTAINER_ID)

.PHONY: build run run-custom logs