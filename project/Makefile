AUTH_BINARY=authApp
POST_BINARY=postApp
IMAGE_BINARY=imageApp
IDENTITY_BINARY=identityApp
MESSAGING_BINARY=messagingApp
FRIEND_BINARY=friendApp

## up: starts all containers in the background without forcing build
up:
	@echo "Starting Docker images..."
	docker-compose up -d
	@echo "Docker images started!"

## up_build: stops docker-compose (if running), builds all projects and starts docker compose
up_build: build_auth build_post build_image build_identity build_messaging build_image_proccessing_service_worker build_friend build_java
	@echo "Stopping docker images (if running...)"
	docker-compose down
	@echo "Building (when required) and starting docker images..."
	docker-compose up --build
	@echo "Docker images built and started!"

## down: stop docker compose
down:
	@echo "Stopping docker compose..."
	docker-compose down
	@echo "Done!"

## build_auth: builds the auth binary as a linux executable
build_auth:
	@echo "Building auth binary..."
	cd ../authentication-service && env GOOS=linux CGO_ENABLED=0 go build -o ${AUTH_BINARY} ./cmd/api
	@echo "Done!"

## build_post: builds the pos binary as a linux executable
build_post:
	@echo "Building post-service binary..."
	cd ../post-service && env GOOS=linux CGO_ENABLED=0 go build -o ${POST_BINARY} ./cmd/api
	@echo "Done!"

## build_post: builds the pos binary as a linux executable
build_image:
	@echo "Building image-service binary..."
	cd ../image-service && env GOOS=linux CGO_ENABLED=0 go build -o ${IMAGE_BINARY} ./cmd/api
	@echo "Done!"

## build_identity: builds the pos binary as a linux executable
build_identity:
	@echo "Building identity-service binary..."
	cd ../identity-service && env GOOS=linux CGO_ENABLED=0 go build -o ${IDENTITY_BINARY} ./cmd/api
	@echo "Done!"

## build_messaging: builds the pos binary as a linux executable
build_messaging:
	@echo "Building identity-service binary..."
	cd ../messaging-service && env GOOS=linux CGO_ENABLED=0 go build -o ${MESSAGING_BINARY} ./cmd/api
	@echo "Done!"
## build_friend: builds the pos binary as a linux executable
build_friend:
	@echo "Building friend-service binary..."
	cd ../friend-service && env GOOS=linux CGO_ENABLED=0 go build -o ${FRIEND_BINARY} ./cmd/api
	@echo "Done!"

build_image_proccessing_service_worker:
	@echo "Building image_proccessing binary..."
	cd ../image-proccessing-service-worker && docker build -t image-proccessing-service-worker -f image-proccessing-service-worker.dockerfile .
	@echo "Done!"

build_java:
	@echo "Building image_proccessing binary..."
	cd ../notificationService && mvn clean package
	@echo "Done!"