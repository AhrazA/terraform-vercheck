# *WARNING* - Do not modify `VERSION` manually. It is managed bu `bumpversion`.
VERSION = 1.0.1

BUILD_DIR = build
DOCKER_IMAGE_NAME = ahraza/terraform-vercheck
DOCKER_IMAGE_VERSION_TAG = ${VERSION}
DOCKER_IMAGE_LATEST_TAG = latest

DOCKER_IMAGE_VERSION_REF = $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_VERSION_TAG)
DOCKER_IMAGE_LATEST_REF = $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_LATEST_TAG)

clean:
	rm -Rf ./$(BUILD_DIR)

.PHONY: lint test docker-build docker-push

lint:
	golint ./...
	go vet ./...
	staticcheck ./...

test: lint
	go test ./...

build: test
	mkdir -p $(BUILD_DIR)
	go build -o "$(BUILD_DIR)/vercheck" ./

alpine-build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-s' -o "$(BUILD_DIR)/vercheck" ./

docker-build:
	docker build -t $(DOCKER_IMAGE_VERSION_REF) -t $(DOCKER_IMAGE_LATEST_REF) ./

docker-push: docker-build
	docker push $(DOCKER_IMAGE_VERSION_REF)
