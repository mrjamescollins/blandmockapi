# Deployment Guide

Complete guide for deploying Bland Mock API to AWS using CDK, with TeamCity CI/CD integration.

## Table of Contents

- [Prerequisites](#prerequisites)
- [AWS CDK Deployment](#aws-cdk-deployment)
- [Lambda Deployment](#lambda-deployment)
- [ECS Deployment](#ecs-deployment)
- [App Runner Deployment](#app-runner-deployment)
- [TeamCity Integration](#teamcity-integration)
- [Configuration Management](#configuration-management)
- [Monitoring & Logging](#monitoring--logging)

## Prerequisites

### Required Tools

```bash
# AWS CLI
aws --version

# AWS CDK
npm install -g aws-cdk
cdk --version

# Node.js & TypeScript
node --version
npm --version

# Go
go version

# Docker
docker --version
```

### AWS Configuration

```bash
# Configure AWS credentials
aws configure

# Set environment variables
export AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
export AWS_REGION=us-east-1
export ENVIRONMENT=dev
```

### CDK Bootstrap

```bash
# Bootstrap CDK (one-time per account/region)
cd cdk
npm install
cdk bootstrap aws://$AWS_ACCOUNT_ID/$AWS_REGION
```

## AWS CDK Deployment

### Project Structure

```
cdk/
├── bin/
│   └── app.ts              # CDK app entry point
├── lib/
│   ├── lambda-stack.ts     # Lambda infrastructure
│   ├── ecs-stack.ts        # ECS infrastructure
│   └── apprunner-stack.ts  # App Runner infrastructure
├── package.json
├── tsconfig.json
└── cdk.json
```

### Stack Overview

| Stack | Resources | Best For |
|-------|-----------|----------|
| Lambda | Lambda Function, API Gateway, Function URL | Serverless, low cost, variable traffic |
| ECS | Fargate, ALB, Auto-scaling | Consistent traffic, container orchestration |
| App Runner | Managed service, auto-scaling | Simple deployment, managed infrastructure |

### Common CDK Commands

```bash
cd cdk

# List all stacks
cdk list

# Synthesize CloudFormation
cdk synth

# Show differences
cdk diff

# Deploy specific stack
cdk deploy BlandMockApiLambdaStack-dev

# Deploy all stacks
cdk deploy --all

# Destroy stack
cdk destroy BlandMockApiLambdaStack-dev
```

## Lambda Deployment

### Architecture

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
┌──────▼──────────────────┐
│  Lambda Function URL    │
│  or API Gateway         │
└──────┬──────────────────┘
       │
┌──────▼──────────────────┐
│  Lambda Function        │
│  (Docker Image)         │
│  - Go Runtime           │
│  - Mock API Server      │
└─────────────────────────┘
```

### Deploy Lambda Stack

```bash
# Build and deploy
cd cdk
npm install
cdk deploy BlandMockApiLambdaStack-dev --require-approval never

# Or using make
cd ..
make quick-lambda LAMBDA_FUNCTION_NAME=blandmockapi-dev
```

### Lambda Configuration

Customize in `cdk/bin/app.ts`:

```typescript
new BlandMockApiLambdaStack(app, `BlandMockApiLambdaStack-${environment}`, {
  environment,
  memorySize: 512,      // 128 MB - 10240 MB
  timeout: 30,          // Max 900 seconds
  configPath: '/var/task/config',
});
```

### Testing Lambda Deployment

```bash
# Get function URL
FUNCTION_URL=$(aws cloudformation describe-stacks \
  --stack-name BlandMockApiLambdaStack-dev \
  --query 'Stacks[0].Outputs[?OutputKey==`FunctionUrl`].OutputValue' \
  --output text)

# Test health endpoint
curl "$FUNCTION_URL/health"

# Test API endpoint
curl "$FUNCTION_URL/api/users"

# Test GraphQL
curl -X POST "$FUNCTION_URL/graphql" \
  -H "Content-Type: application/json" \
  -d '{"query":"{ users { id name } }"}'
```

### Lambda Pricing

- **Requests**: $0.20 per 1M requests
- **Duration**: $0.0000166667 per GB-second
- **Example**: 1M requests at 512MB, 100ms avg = ~$5/month

## ECS Deployment

### Architecture

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
┌──────▼──────────────────┐
│  Application Load       │
│  Balancer (ALB)         │
└──────┬──────────────────┘
       │
┌──────▼──────────────────┐
│  ECS Service            │
│  ┌──────────────────┐   │
│  │  Fargate Task 1  │   │
│  │  (Container)     │   │
│  └──────────────────┘   │
│  ┌──────────────────┐   │
│  │  Fargate Task 2  │   │
│  │  (Container)     │   │
│  └──────────────────┘   │
└─────────────────────────┘
```

### Deploy ECS Stack

```bash
# Deploy with default VPC
cd cdk
cdk deploy BlandMockApiEcsStack-dev

# Deploy with existing VPC
cdk deploy BlandMockApiEcsStack-dev \
  -c vpcId=vpc-xxxxx

# Or using make
cd ..
make quick-ecs \
  ECS_CLUSTER=blandmockapi-dev \
  ECS_SERVICE=blandmockapi-dev
```

### ECS Configuration

Customize in `cdk/bin/app.ts`:

```typescript
new BlandMockApiEcsStack(app, `BlandMockApiEcsStack-${environment}`, {
  environment,
  desiredCount: 2,      // Number of tasks
  cpu: 256,            // 256, 512, 1024, 2048, 4096
  memoryLimitMiB: 512, // Must pair with CPU
  vpcId: 'vpc-xxxxx',  // Optional: use existing VPC
});
```

### Auto-Scaling Configuration

The stack includes auto-scaling based on:

- **CPU**: Target 70% utilization
- **Memory**: Target 80% utilization
- **Min tasks**: 2
- **Max tasks**: 10

### Testing ECS Deployment

```bash
# Get ALB URL
ALB_URL=$(aws cloudformation describe-stacks \
  --stack-name BlandMockApiEcsStack-dev \
  --query 'Stacks[0].Outputs[?OutputKey==`LoadBalancerUrl`].OutputValue' \
  --output text)

# Wait for service to be stable
aws ecs wait services-stable \
  --cluster blandmockapi-dev \
  --services blandmockapi-dev

# Test endpoint
curl "$ALB_URL/health"
curl "$ALB_URL/api/users"
```

### ECS Pricing

- **Fargate**: $0.04048 per vCPU per hour + $0.004445 per GB per hour
- **ALB**: $0.0225 per hour + $0.008 per LCU-hour
- **Example**: 2 tasks (0.25 vCPU, 0.5 GB) = ~$25/month

## App Runner Deployment

### Architecture

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
┌──────▼──────────────────┐
│  App Runner Service     │
│  (Managed)              │
│                         │
│  ┌──────────────────┐   │
│  │  Auto-scaled     │   │
│  │  Instances       │   │
│  └──────────────────┘   │
└─────────────────────────┘
```

### Deploy App Runner Stack

```bash
# Deploy
cd cdk
cdk deploy BlandMockApiAppRunnerStack-dev

# Or using make
cd ..
make quick-apprunner \
  APPRUNNER_SERVICE_ARN=arn:aws:apprunner:...
```

### App Runner Configuration

Customize in `cdk/bin/app.ts`:

```typescript
new BlandMockApiAppRunnerStack(app, `BlandMockApiAppRunnerStack-${environment}`, {
  environment,
  cpu: '0.25 vCPU',              // 0.25, 0.5, 1, 2, 4 vCPU
  memory: '0.5 GB',               // 0.5, 1, 2, 3, 4 GB
  port: 8080,
  autoScalingMaxConcurrency: 100, // Requests per instance
  autoScalingMaxSize: 10,         // Max instances
  autoScalingMinSize: 1,          // Min instances
});
```

### Testing App Runner Deployment

```bash
# Get service URL
SERVICE_URL=$(aws cloudformation describe-stacks \
  --stack-name BlandMockApiAppRunnerStack-dev \
  --query 'Stacks[0].Outputs[?OutputKey==`ServiceUrl`].OutputValue' \
  --output text)

# Wait for service to be running
SERVICE_ARN=$(aws cloudformation describe-stacks \
  --stack-name BlandMockApiAppRunnerStack-dev \
  --query 'Stacks[0].Outputs[?OutputKey==`ServiceArn`].OutputValue' \
  --output text)

aws apprunner describe-service \
  --service-arn $SERVICE_ARN \
  --query 'Service.Status'

# Test endpoint
curl "$SERVICE_URL/health"
curl "$SERVICE_URL/api/users"
```

### App Runner Pricing

- **Compute**: $0.064 per vCPU per hour + $0.007 per GB per hour
- **Active**: Charged when processing requests
- **Idle**: $0.007 per hour when no requests
- **Example**: 0.25 vCPU, 0.5 GB, 50% active = ~$15/month

## TeamCity Integration

### Configuration

TeamCity configuration is in `.teamcity/settings.kts` (Kotlin DSL).

### Build Configurations

1. **Unit Tests**: Run on every commit
2. **Integration Tests**: Run after unit tests pass
3. **Build Docker**: Build and push images to ECR
4. **Deploy Lambda**: Deploy Lambda stack
5. **Deploy ECS**: Deploy ECS stack
6. **Deploy App Runner**: Deploy App Runner stack

### Setting Up TeamCity

1. **Create Project**:
   - VCS Root: Point to GitHub repository
   - Enable versioned settings
   - Use Kotlin DSL format

2. **Configure Parameters**:
   ```
   env.GO_VERSION = 1.25
   env.AWS_REGION = us-east-1
   env.AWS_ACCOUNT_ID = <your-account-id>
   env.ENVIRONMENT = dev
   ```

3. **AWS Credentials**:
   - Add AWS credentials as TeamCity parameters
   - Use AWS Parameter Store or Secrets Manager
   - Configure IAM role for TeamCity agents

### Triggering Builds

```bash
# Via TeamCity UI
# Or via REST API:

curl -X POST https://teamcity.example.com/app/rest/buildQueue \
  -H "Authorization: Bearer $TEAMCITY_TOKEN" \
  -H "Content-Type: application/xml" \
  -d '<build><buildType id="BlandMockApi_DeployLambda"/></build>'
```

### Build Dependencies

```
UnitTests
    ↓
IntegrationTests
    ↓
BuildDocker
    ↓
┌───────┬─────────┬──────────┐
│       │         │          │
Deploy  Deploy    Deploy
Lambda  ECS       AppRunner
```

## Configuration Management

### Environment-Specific Config

```bash
# Store configs in S3
aws s3 cp examples/ s3://blandmockapi-config-dev/config/ --recursive

# Download in Lambda
aws s3 sync s3://blandmockapi-config-dev/config/ /tmp/config/
```

### Using Parameter Store

```typescript
// In CDK stack
import * as ssm from 'aws-cdk-lib/aws-ssm';

const configPath = ssm.StringParameter.valueFromLookup(
  this,
  '/blandmockapi/dev/config-path'
);
```

```bash
# Store parameter
aws ssm put-parameter \
  --name /blandmockapi/dev/config-path \
  --value /var/task/config \
  --type String
```

### Secrets Management

```bash
# Store secret
aws secretsmanager create-secret \
  --name blandmockapi/dev/api-key \
  --secret-string "your-secret-key"

# Access in code
export API_KEY=$(aws secretsmanager get-secret-value \
  --secret-id blandmockapi/dev/api-key \
  --query SecretString --output text)
```

## Monitoring & Logging

### CloudWatch Logs

```bash
# Lambda logs
aws logs tail /aws/lambda/blandmockapi-dev --follow

# ECS logs
aws logs tail /aws/ecs/blandmockapi/blandmockapi --follow

# App Runner logs
aws logs tail /aws/apprunner/blandmockapi-dev/application --follow
```

### CloudWatch Metrics

```bash
# Lambda metrics
aws cloudwatch get-metric-statistics \
  --namespace AWS/Lambda \
  --metric-name Invocations \
  --dimensions Name=FunctionName,Value=blandmockapi-dev \
  --start-time 2024-01-01T00:00:00Z \
  --end-time 2024-01-01T23:59:59Z \
  --period 3600 \
  --statistics Sum

# ECS metrics
aws cloudwatch get-metric-statistics \
  --namespace AWS/ECS \
  --metric-name CPUUtilization \
  --dimensions Name=ServiceName,Value=blandmockapi-dev \
  --start-time 2024-01-01T00:00:00Z \
  --end-time 2024-01-01T23:59:59Z \
  --period 300 \
  --statistics Average
```

### Alarms

```typescript
// In CDK stack
import * as cloudwatch from 'aws-cdk-lib/aws-cloudwatch';

const errorAlarm = new cloudwatch.Alarm(this, 'ErrorAlarm', {
  metric: fn.metricErrors(),
  threshold: 10,
  evaluationPeriods: 2,
  alarmDescription: 'Alert when function errors exceed threshold',
});
```

### X-Ray Tracing

Enable in CDK:

```typescript
const mockApiFunction = new lambda.DockerImageFunction(this, 'Function', {
  // ... other props
  tracing: lambda.Tracing.ACTIVE,
});
```

View traces:

```bash
aws xray get-trace-summaries \
  --start-time 2024-01-01T00:00:00Z \
  --end-time 2024-01-01T23:59:59Z
```

## Rollback & Disaster Recovery

### Lambda Rollback

```bash
# List versions
aws lambda list-versions-by-function \
  --function-name blandmockapi-dev

# Update alias to previous version
aws lambda update-alias \
  --function-name blandmockapi-dev \
  --name live \
  --function-version 3
```

### ECS Rollback

```bash
# Describe previous task definition
aws ecs describe-task-definition \
  --task-definition blandmockapi-dev:2

# Update service to use previous version
aws ecs update-service \
  --cluster blandmockapi-dev \
  --service blandmockapi-dev \
  --task-definition blandmockapi-dev:2 \
  --force-new-deployment
```

### CDK Rollback

```bash
# Destroy and redeploy
cdk destroy BlandMockApiLambdaStack-dev
cdk deploy BlandMockApiLambdaStack-dev

# Or use CloudFormation
aws cloudformation rollback-stack \
  --stack-name BlandMockApiLambdaStack-dev
```

## Best Practices

### 1. Use Environment Variables

```typescript
environment: {
  ENVIRONMENT: env,
  LOG_LEVEL: 'info',
  CONFIG_PATH: '/var/task/config',
}
```

### 2. Enable Encryption

```typescript
import * as kms from 'aws-cdk-lib/aws-kms';

const key = new kms.Key(this, 'EncryptionKey', {
  enableKeyRotation: true,
});
```

### 3. Implement Health Checks

All stacks include health check configuration pointing to `/health`.

### 4. Use Blue/Green Deployments

```typescript
deploymentConfiguration: {
  type: ecs.DeploymentControllerType.CODE_DEPLOY,
}
```

### 5. Tag Resources

```typescript
cdk.Tags.of(this).add('Project', 'BlandMockApi');
cdk.Tags.of(this).add('Environment', env);
cdk.Tags.of(this).add('CostCenter', 'Engineering');
cdk.Tags.of(this).add('ManagedBy', 'CDK');
```

## Troubleshooting

### CDK Deployment Failures

```bash
# Check CDK diff
cdk diff BlandMockApiLambdaStack-dev

# Synthesize and inspect
cdk synth BlandMockApiLambdaStack-dev > template.yaml

# Check CloudFormation events
aws cloudformation describe-stack-events \
  --stack-name BlandMockApiLambdaStack-dev
```

### Lambda Issues

```bash
# Check function configuration
aws lambda get-function-configuration \
  --function-name blandmockapi-dev

# Invoke function directly
aws lambda invoke \
  --function-name blandmockapi-dev \
  --payload '{}' \
  response.json
```

### ECS Issues

```bash
# Check service events
aws ecs describe-services \
  --cluster blandmockapi-dev \
  --services blandmockapi-dev

# Check task status
aws ecs list-tasks \
  --cluster blandmockapi-dev \
  --service-name blandmockapi-dev
```

## Resources

- [AWS CDK Documentation](https://docs.aws.amazon.com/cdk/)
- [TeamCity Kotlin DSL](https://www.jetbrains.com/help/teamcity/kotlin-dsl.html)
- [AWS Lambda Best Practices](https://docs.aws.amazon.com/lambda/latest/dg/best-practices.html)
- [ECS Best Practices](https://docs.aws.amazon.com/AmazonECS/latest/bestpracticesguide/)
- [App Runner Documentation](https://docs.aws.amazon.com/apprunner/)
