
# Variables
APP_NAME = event-receiver
DOCKER_IMAGE = $(APP_NAME):latest
ECR_REGISTRY = your-ecr-repo
GO_VERSION = 1.24
TERRAFORM_DIR = terraform
TEST_DIR = ./...
BINARY = ./bin/$(APP_NAME)
CONFIG_FILE = config.yaml

# Default target
.PHONY: all
all: build

# Build the Go binary
.PHONY: build
build: clean deps
	@echo "Building $(APP_NAME)..."
	@mkdir -p bin/
	@go build -o $(BINARY) ./cmd/main.go

# Run the application locally
.PHONY: run
run:
	@echo "Running $(APP_NAME)..."
	@go run ./cmd/main.go

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	@go test -v $(TEST_DIR)/...

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Vet code for potential issues
.PHONY: vet
vet:
	@echo "Vetting code..."
	@go vet ./...

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning..."
	@rm -f $(BINARY)
	@go clean

# Build Docker image
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE) .

# Run Docker container locally
.PHONY: docker-run
docker-run: docker-build
	@echo "Running Docker container..."
	@docker run -p 8080:8000 -v $(PWD)/$(CONFIG_FILE):/app/$(CONFIG_FILE) $(DOCKER_IMAGE)

# Push Docker image to ECR
.PHONY: docker-push
docker-push: docker-build
	@echo "Pushing Docker image to ECR..."
	@docker tag $(DOCKER_IMAGE) $(ECR_REGISTRY)/$(DOCKER_IMAGE)
	@docker push $(ECR_REGISTRY)/$(DOCKER_IMAGE)

# Initialize Terraform
.PHONY: terraform-init
terraform-init:
	@echo "Initializing Terraform..."
	@cd $(TERRAFORM_DIR) && terraform init

# Plan Terraform changes
.PHONY: terraform-plan
terraform-plan: terraform-init
	@echo "Planning Terraform changes..."
	@cd $(TERRAFORM_DIR) && terraform plan

# Apply Terraform configuration
.PHONY: terraform-apply
terraform-apply: terraform-init
	@echo "Applying Terraform configuration..."
	@cd $(TERRAFORM_DIR) && terraform apply -auto-approve

# Destroy Terraform resources
.PHONY: terraform-destroy
terraform-destroy:
	@echo "Destroying Terraform resources..."
	@cd $(TERRAFORM_DIR) && terraform destroy -auto-approve

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	@go mod tidy
	@go mod download


# Full development setup
.PHONY: setup
setup: deps fmt vet test build

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'