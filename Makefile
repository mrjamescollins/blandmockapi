.PHONY: help build run test clean docker-build docker-run lambda-build lambda-deploy deploy-ecs deploy-apprunner

# Variables
APP_NAME := blandmockapi
DOCKER_IMAGE := $(APP_NAME):latest
DOCKER_LAMBDA_IMAGE := $(APP_NAME):lambda
AWS_REGION ?= us-east-1
AWS_ACCOUNT_ID ?= $(shell aws sts get-caller-identity --query Account --output text)
ECR_REPO ?= $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com/$(APP_NAME)

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

build: ## Build the application binary
	@echo "Building $(APP_NAME)..."
	@go build -ldflags="-w -s" -o bin/$(APP_NAME) ./cmd/server

build-lambda: ## Build the application for AWS Lambda
	@echo "Building $(APP_NAME) for Lambda..."
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build \
		-tags lambda \
		-ldflags="-w -s" \
		-o bin/bootstrap \
		./cmd/server

run: ## Run the application locally
	@echo "Running $(APP_NAME)..."
	@go run ./cmd/server -config ./examples

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@docker rmi -f $(DOCKER_IMAGE) $(DOCKER_LAMBDA_IMAGE) 2>/dev/null || true

# Docker targets
docker-build: ## Build Docker image for ECS/AppRunner
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE) .

docker-run: ## Run Docker container locally
	@echo "Running Docker container..."
	@docker run --rm -p 8080:8080 -v $(PWD)/examples:/app/examples $(DOCKER_IMAGE)

docker-lambda-build: ## Build Docker image for Lambda
	@echo "Building Lambda Docker image..."
	@docker build -f Dockerfile.lambda -t $(DOCKER_LAMBDA_IMAGE) .

# AWS ECR targets
ecr-login: ## Login to AWS ECR
	@echo "Logging into ECR..."
	@aws ecr get-login-password --region $(AWS_REGION) | \
		docker login --username AWS --password-stdin $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com

ecr-create: ## Create ECR repository
	@echo "Creating ECR repository..."
	@aws ecr create-repository \
		--repository-name $(APP_NAME) \
		--region $(AWS_REGION) \
		--image-scanning-configuration scanOnPush=true \
		|| echo "Repository might already exist"

ecr-push: ecr-login ## Push Docker image to ECR
	@echo "Pushing to ECR..."
	@docker tag $(DOCKER_IMAGE) $(ECR_REPO):latest
	@docker push $(ECR_REPO):latest

ecr-push-lambda: ecr-login ## Push Lambda Docker image to ECR
	@echo "Pushing Lambda image to ECR..."
	@docker tag $(DOCKER_LAMBDA_IMAGE) $(ECR_REPO):lambda
	@docker push $(ECR_REPO):lambda

# Lambda deployment
lambda-package: build-lambda ## Package Lambda function for deployment
	@echo "Packaging Lambda function..."
	@mkdir -p bin/lambda-package
	@cp bin/bootstrap bin/lambda-package/
	@cp -r examples bin/lambda-package/config
	@cd bin/lambda-package && zip -r ../lambda.zip .

lambda-deploy: lambda-package ## Deploy to AWS Lambda (requires LAMBDA_FUNCTION_NAME)
	@echo "Deploying to Lambda..."
	@aws lambda update-function-code \
		--function-name $(LAMBDA_FUNCTION_NAME) \
		--zip-file fileb://bin/lambda.zip \
		--region $(AWS_REGION)

lambda-deploy-docker: docker-lambda-build ecr-create ecr-push-lambda ## Deploy Lambda using Docker image
	@echo "Deploying Lambda with Docker image..."
	@aws lambda update-function-code \
		--function-name $(LAMBDA_FUNCTION_NAME) \
		--image-uri $(ECR_REPO):lambda \
		--region $(AWS_REGION)

# ECS deployment
deploy-ecs: docker-build ecr-create ecr-push ## Deploy to AWS ECS (requires ECS_CLUSTER, ECS_SERVICE)
	@echo "Deploying to ECS..."
	@aws ecs update-service \
		--cluster $(ECS_CLUSTER) \
		--service $(ECS_SERVICE) \
		--force-new-deployment \
		--region $(AWS_REGION)
	@echo "ECS deployment initiated. Monitor with: aws ecs describe-services --cluster $(ECS_CLUSTER) --services $(ECS_SERVICE)"

# App Runner deployment
deploy-apprunner: docker-build ecr-create ecr-push ## Deploy to AWS App Runner (requires APPRUNNER_SERVICE_ARN)
	@echo "Deploying to App Runner..."
	@aws apprunner start-deployment \
		--service-arn $(APPRUNNER_SERVICE_ARN) \
		--region $(AWS_REGION)
	@echo "App Runner deployment initiated. Check status in AWS Console."

# Development targets
fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

lint: fmt vet ## Run formatters and linters
	@echo "Linting complete"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Quick deploy targets
quick-lambda: build-lambda lambda-package lambda-deploy ## Quick build and deploy to Lambda
	@echo "Lambda deployment complete!"

quick-ecs: docker-build ecr-create ecr-push deploy-ecs ## Quick build and deploy to ECS
	@echo "ECS deployment complete!"

quick-apprunner: docker-build ecr-create ecr-push deploy-apprunner ## Quick build and deploy to App Runner
	@echo "App Runner deployment complete!"
