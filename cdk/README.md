# Bland Mock API - CDK Infrastructure

AWS CDK infrastructure code for deploying Bland Mock API to Lambda, ECS, and App Runner.

## Quick Start

```bash
# Install dependencies
npm install

# Configure AWS
export AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
export AWS_REGION=us-east-1

# Bootstrap CDK (one-time)
cdk bootstrap aws://$AWS_ACCOUNT_ID/$AWS_REGION

# Deploy Lambda
npm run deploy:lambda

# Deploy ECS
npm run deploy:ecs

# Deploy App Runner
npm run deploy:apprunner

# Deploy all
npm run deploy:all
```

## Stacks

### Lambda Stack (`BlandMockApiLambdaStack`)

Deploys the mock API as a Lambda function with:
- Docker image from ECR
- Lambda Function URL
- API Gateway (optional)
- CloudWatch Logs

**Resources**:
- Lambda Function (512 MB, 30s timeout)
- Function URL with CORS
- API Gateway REST API
- CloudWatch Log Groups

### ECS Stack (`BlandMockApiEcsStack`)

Deploys the mock API on ECS Fargate with:
- Application Load Balancer
- Auto-scaling (2-10 tasks)
- VPC with public/private subnets
- CloudWatch Container Insights

**Resources**:
- ECS Cluster
- Fargate Service
- Application Load Balancer
- VPC (or use existing)
- Auto-scaling policies

### App Runner Stack (`BlandMockApiAppRunnerStack`)

Deploys the mock API on App Runner with:
- Managed container service
- Auto-scaling (1-10 instances)
- HTTPS endpoint
- Health checks

**Resources**:
- App Runner Service
- ECR Image Asset
- IAM Roles
- Auto-scaling Configuration

## Configuration

### Environment Context

```bash
# Deploy to dev environment
cdk deploy --context environment=dev

# Deploy to prod environment
cdk deploy --context environment=prod
```

### Custom Parameters

Edit `bin/app.ts` to customize:

```typescript
// Lambda configuration
memorySize: 512,      // 128-10240 MB
timeout: 30,          // 1-900 seconds

// ECS configuration
desiredCount: 2,      // Number of tasks
cpu: 256,            // 256, 512, 1024, 2048, 4096
memoryLimitMiB: 512, // Depends on CPU

// App Runner configuration
cpu: '0.25 vCPU',     // 0.25, 0.5, 1, 2, 4
memory: '0.5 GB',     // 0.5-4 GB
```

## Commands

```bash
# List all stacks
cdk list

# Show CloudFormation template
cdk synth

# Compare deployed stack with current state
cdk diff

# Deploy specific stack
cdk deploy BlandMockApiLambdaStack-dev

# Destroy stack
cdk destroy BlandMockApiLambdaStack-dev
```

## Outputs

After deployment, CDK outputs important values:

### Lambda
- `FunctionUrl`: Lambda function URL
- `ApiEndpoint`: API Gateway endpoint
- `FunctionName`: Lambda function name
- `FunctionArn`: Lambda function ARN

### ECS
- `LoadBalancerUrl`: ALB DNS name
- `ClusterName`: ECS cluster name
- `ServiceName`: ECS service name
- `ServiceArn`: ECS service ARN

### App Runner
- `ServiceUrl`: HTTPS service URL
- `ServiceId`: App Runner service ID
- `ServiceArn`: App Runner service ARN

## Testing Deployments

```bash
# Test Lambda
FUNCTION_URL=$(aws cloudformation describe-stacks \
  --stack-name BlandMockApiLambdaStack-dev \
  --query 'Stacks[0].Outputs[?OutputKey==`FunctionUrl`].OutputValue' \
  --output text)
curl "$FUNCTION_URL/health"

# Test ECS
ALB_URL=$(aws cloudformation describe-stacks \
  --stack-name BlandMockApiEcsStack-dev \
  --query 'Stacks[0].Outputs[?OutputKey==`LoadBalancerUrl`].OutputValue' \
  --output text)
curl "$ALB_URL/health"

# Test App Runner
SERVICE_URL=$(aws cloudformation describe-stacks \
  --stack-name BlandMockApiAppRunnerStack-dev \
  --query 'Stacks[0].Outputs[?OutputKey==`ServiceUrl`].OutputValue' \
  --output text)
curl "$SERVICE_URL/health"
```

## Cost Estimation

### Lambda (estimated)
- **Compute**: ~$5-10/month for moderate usage
- **API Gateway**: ~$3.50 per million requests
- **Best for**: Variable/sporadic traffic

### ECS (estimated)
- **Fargate**: ~$25-30/month (2 tasks, 0.25 vCPU, 0.5 GB)
- **ALB**: ~$18/month
- **Total**: ~$45-50/month
- **Best for**: Consistent traffic

### App Runner (estimated)
- **Compute**: ~$15-20/month (0.25 vCPU, 0.5 GB, 50% active)
- **Best for**: Simple deployments, managed infrastructure

## Architecture Diagrams

### Lambda Architecture
```
Internet → API Gateway/Function URL → Lambda → Response
```

### ECS Architecture
```
Internet → ALB → Target Group → ECS Tasks (Fargate) → Response
```

### App Runner Architecture
```
Internet → App Runner (Managed) → Container Instances → Response
```

## CI/CD Integration

### TeamCity

The `.teamcity/settings.kts` file includes build configurations for:
- Unit tests
- Integration tests
- Docker build
- CDK deployments

### GitHub Actions

The `.github/workflows/ci.yml` includes:
- Automated testing
- Docker builds
- Deployment workflows

## Troubleshooting

### Bootstrap Issues

```bash
# Re-bootstrap if needed
cdk bootstrap aws://$AWS_ACCOUNT_ID/$AWS_REGION --force
```

### Permission Issues

Ensure your AWS credentials have:
- ECR access (push images)
- CloudFormation access
- Lambda/ECS/App Runner permissions
- IAM role creation

### Deployment Failures

```bash
# Check CloudFormation events
aws cloudformation describe-stack-events \
  --stack-name BlandMockApiLambdaStack-dev

# View detailed errors
cdk deploy --verbose
```

## Cleanup

```bash
# Delete all stacks
cdk destroy --all

# Or individual stacks
cdk destroy BlandMockApiLambdaStack-dev
cdk destroy BlandMockApiEcsStack-dev
cdk destroy BlandMockApiAppRunnerStack-dev
```

## Resources

- [AWS CDK TypeScript Reference](https://docs.aws.amazon.com/cdk/api/v2/docs/aws-construct-library.html)
- [CDK Best Practices](https://docs.aws.amazon.com/cdk/latest/guide/best-practices.html)
- [Parent Project README](../README.md)
- [Deployment Guide](../DEPLOYMENT.md)
