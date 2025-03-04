# Variables
LOCALSTACK_IMAGE = localstack/localstack:latest
LOCALSTACK_CONTAINER = localstack_s3
PORT_S3 = 4572 # Port par défaut pour S3 dans LocalStack

.PHONY: all
all: start

.PHONY: start
start:
	docker run --rm -d \
		--name $(LOCALSTACK_CONTAINER) \
		-p $(PORT_S3):4572 \
		-e DOCKER_HOST=unix:///var/run/docker.sock \
		-v /var/run/docker.sock:/var/run/docker.sock \
		$(LOCALSTACK_IMAGE)

.PHONY: stop
stop:
	docker stop $(LOCALSTACK_CONTAINER)
