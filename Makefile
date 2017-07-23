TARGET = pgm
IMAGE_TARGET = $(TARGET)-linux-64
SOURCES = $(shell find . -type f -name '*.go')
VERSION ?= latest
REPO = walle/$(TARGET)

.PHONY: help test docker-login push

help: ## Prints this help.
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
	| awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-10s\033[0m %s\n", $$1, $$2}'

test: ## Run the test suite
	go test -cover -bench=. -benchmem

$(TARGET): ## Build binary for current arch/os
	go build -o $(TARGET)

$(IMAGE_TARGET): $(SOURCES) ## Build binary for docker image
	env GOOS=linux GOARCH=amd64 go build -tags netgo -ldflags "-w" -o $(IMAGE_TARGET)

image: $(IMAGE_TARGET) ## Builds the docker image
	docker build -t $(REPO):$(VERSION) .

docker-login: ## [ci-only] Login to docker registry
	@docker login -u=$(DOCKER_USERNAME) -p=$(DOCKER_PASSWORD)

push: docker-login ## [ci-only] Push Docker container
	docker tag $(REPO):$(COMMIT) $(REPO):$(TAG)
	docker tag $(REPO):$(COMMIT) $(REPO):$(TRAVIS_BUILD_NUMBER)
	docker push $(REPO)
