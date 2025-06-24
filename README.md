# High-Throughput Event Receiver Service

## Overview
This is a high-throughput event receiver service built in Go using the Gin framework. It receives events via an HTTP POST endpoint (`/ingest`), filters them based on the `X-Customer-Tier` header, batches them, and stores them in AWS S3 with low latency and high throughput (>100 req/s). The service can be deployed on AWS ECS Fargate or any compute engine.

## Project Structure

event_receiver_service/
   - cmd/                 # Main application entry point
   - internal/            # Core application logic (config, handler, batch, metrics, model)
   - terraform/           # Terraform configuration for ECS deployment
   - tests                # load testing
   - config.yaml          # Application configuration
   - Dockerfile           # Docker image definition
   - Makefile             # Build and deployment tasks
   - README.md            # This file



## Prerequisites
- Go 1.24.4
- Docker
- AWS CLI configured with credentials
- Terraform (optional for infrastructure deployment)
- - python 3.x (for load test)

- **AWS CLI**: Configured with credentials having permissions for ECR, ECS, S3, DynamoDB, and CloudFormation.
- **Docker**: Installed locally for building images.
- **Go**: Version 1.24 for local builds and tests.
- **Terraform**: Version 1.12.2 for infrastructure deployment.
- **Git**: For managing the codebase.
- **AWS Account**: Access to `ap-south-1` region with permissions for:
   - `ecr:CreateRepository`, `ecr:PutImage`, etc.
   - `ecs:*`, `s3:*`, `dynamodb:*`, `iam:*`, `cloudwatch:*`
   - `cloudformation:*`

- Update s3_bucket to match your S3 bucket name (e.g., ravi-portal26-events).

## Configuration
The application uses a `config.yaml` file:
```yaml
---
valid_tiers:
  - free
  - pro
  - enterprise
s3_bucket: ravi-portal26-events
batch_size_mb: 5
batch_timeout_seconds: 5
aws_region: ap-south-1
port: 8000
```
## Build and Run Locally

1. Place config.yaml in the project root.

2. abuild the service:
   ```bash
   make build
   ```

3. Run the service:
   ```bash
   ./bin/event-receiver 
   ```

4. Test the endpoint:
   ```bash
   curl -X POST http://localhost:8000/ingest \
       -H "X-Customer-Tier: pro" \
       -d '{"event_timestamp":"2024-01-01T00:00:00Z","body":"test event"}'
   ```

## Build Docker Image
```bash
make docker-build
```

## Run Docker Image Locally:
```bash
make docker-run
```
- The service runs on http://localhost:8080.
- Test the endpoint:
```bash
curl -X POST http://localhost:8080/ingest \
  -H "X-Customer-Tier: pro" \
  -d '{"event_timestamp":"2025-06-24T18:00:00Z","body":"test event"}'
```
- Expected response: ```{"status":"success"}```

## Create Amazon ECR Repository
1. Create the ECR Repository:
```bash
aws ecr create-repository --repository-name ravi-event-receiver --region ap-south-1
```
- Note the repository URI from the output, e.g., <account-id>.dkr.ecr.ap-south-1.amazonaws.com/ravi-event-receiver.
2. Set the environment variables.
```bash
export ECR_REGISTRY=<account-id>.dkr.ecr.ap-south-1.amazonaws.com/ravi-event-receiver
export AWS_REGION=ap-south-1
```
## Push Docker Image to ECR
1. Log in to ECR:
```bash
aws ecr get-login-password --region ap-south-1 | docker login --username AWS --password-stdin $ECR_REGISTRY
```
2. Tag the docker image:
```bash
docker tag event-receiver:latest $ECR_REGISTRY:latest
```
3. Push the Image to ECR:
```bash
docker push $ECR_REGISTRY:latest
```
## Deploy to AWS
1. Configure AWS credentials:
   ```bash
   export AWS_ACCESS_KEY_ID=<your_access_key>
   export AWS_SECRET_ACCESS_KEY=<your_secret_key>
   ```

2. Initialize Terraform:
   ```bash
   cd terraform
   terraform init
   ```

3. Update terraform variables
   - Edit terraform/variables.tf to include your VPC subnets and security groups
```hcl
variable "region" {
default = "ap-south-1"
}

variable "s3_bucket_name" {
default = "ravi-portal26-events"
}

variable "image" {
default = "<account-id>.dkr.ecr.ap-south-1.amazonaws.com/ravi-event-receiver"
}

variable "subnets" {
type    = list(string)
default = ["subnet-xxxx", "subnet-yyyy"] # Replace with your VPC subnet IDs
}

variable "security_groups" {
type    = list(string)
default = ["sg-zzzz"] # Replace with your security group ID
}
```
4. Apply Terraform configuration:
   ```bash
   terraform apply
   ```
   - This creates:
   - ECS cluster (event-receiver-cluster)
   - IAM role for ECS tasks
   - ECS task definition (event-receiver-task)
   - ECS service (event-receiver-service) with 2 tasks
   - CloudWatch log group (/ecs/event-receiver)
5. Get the Public Endpoint:
    - Find the public IP or DNS name of the ECS service:
```bash
aws ecs describe-services \
  --cluster event-receiver-cluster \
  --services event-receiver-service \
  --region ap-south-1 \
  --query 'services[0].networkConfiguration.awsvpcConfiguration.publicIp'
```
    - Alternatively, check the ECS console for the task's public IP or attach an Application Load Balancer (ALB) for a stable DNS name (requires additional Terraform configuration).

## Testing
Run unit tests:
```bash
   go test ./...
```


## Metrics
Prometheus metrics are available at `http://<hostname>:8000`

/metrics`. Key metrics include:
- `requests_total`: Total HTTP requests
- `filtered_requests_total`: Filtered requests due to invalid tier
- `events_processed_total`: Successfully processed events
- `s3_writes_total`: Total S3 writes
- `s3_errors_total`: S3 write errors

## Scalability Considerations
- Uses Go goroutines for concurrent request handling
- Buffered channels for event processing to prevent blocking
- Batch processing to reduce S3 write operations
- ECS Fargate for horizontal scaling
- Configurable batch size (5MB) and timeout (5s) for optimal throughput
```

## Observability
- Structured logging for success/failure cases
- Prometheus metrics for monitoring
- AWS CloudWatch logs for ECS tasks
```

## Notes
- Ensure AWS credentials have permissions for S3 `PutObject` and ECS operations.
- The service handles 1KB to 10MB request bodies via Gin's default body size limit.
- For production, configure auto-scaling based on CPU/memory metrics in ECS.


This implementation:
- Uses Go's concurrency model with goroutines and channels for high throughput (>100 req/s).
- Implements batching with a 5MB size limit or 5-second timeout to minimize S3 writes while maintaining low latency.
- Provides Prometheus metrics for observability (requests/s, S3 writes/s, filtered requests).
- Includes structured logging for debugging.
- Uses ECS Fargate for scalable deployment with Terraform for infrastructure.
- Validates `X-Customer-Tier` header and handles invalid requests gracefully.
- Includes unit tests for handler and batch processor.

To deploy, follow the README instructions, ensuring AWS credentials are provided and Terraform variables are updated with your VPC subnets and security groups.

## 🚀 Load Testing

### 📦 Install Dependencies
```bash
python -m pip install -r requirements.txt
```

🧪 Run Load Test
```bash
python tests/load_test.py \
  --endpoint http://localhost:8000/ingest \
  --tier pro \
  --target-rps 150 \
  --duration 30 \
  --min-size-kb 100 \
  --max-size-kb 5000
```
### Sample results
Payload Size: 100KB
✅ Total Requests: 2970
✅ Success: 2970
❌ Failed: 0
⚡ Actual RPS: 99.00
⏱️ Avg Response Time: 44.74ms

Payload Size: 1000KB
✅ Total Requests: 1065
✅ Success: 1065
❌ Failed: 0
⚡ Actual RPS: 35.50
⏱️ Avg Response Time: 182.84ms

Payload Size: 5000KB
✅ Total Requests: 225
✅ Success: 225
❌ Failed: 0
⚡ Actual RPS: 7.50
⏱️ Avg Response Time: 959.55ms

### Server output logs
```bash
[GIN] 2025/06/24 - 23:14:52 | 200 |  657.057ms | ::1 | POST "/ingest"
Flushed batch to S3: events/batch_1750787092.jsonl (6.1 MB)
Flushed batch to S3: events/batch_1750787093.jsonl (6.1 MB)
...
Flushed batch to S3: events/batch_1750787298.jsonl (10.2 MB)
```
All payloads were successfully flushed to S3 in .jsonl format with no data loss or errors observed during the test.