import * as cdk from 'aws-cdk-lib';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as apigateway from 'aws-cdk-lib/aws-apigateway';
import * as logs from 'aws-cdk-lib/aws-logs';
import * as iam from 'aws-cdk-lib/aws-iam';
import { Construct } from 'constructs';

export interface LambdaStackProps extends cdk.StackProps {
  readonly environment?: string;
  readonly configPath?: string;
  readonly memorySize?: number;
  readonly timeout?: number;
}

export class BlandMockApiLambdaStack extends cdk.Stack {
  public readonly functionUrl: cdk.CfnOutput;
  public readonly apiEndpoint: cdk.CfnOutput;

  constructor(scope: Construct, id: string, props?: LambdaStackProps) {
    super(scope, id, props);

    const env = props?.environment || 'dev';
    const configPath = props?.configPath || '/var/task/config';

    // Lambda function from Docker image
    const mockApiFunction = new lambda.DockerImageFunction(this, 'BlandMockApiFunction', {
      functionName: `blandmockapi-${env}`,
      description: 'Bland Mock API - Lightweight configurable REST and GraphQL mock server',
      code: lambda.DockerImageCode.fromImageAsset('../', {
        file: 'Dockerfile.lambda',
        buildArgs: {
          BUILDPLATFORM: 'linux/amd64',
        },
      }),
      memorySize: props?.memorySize || 512,
      timeout: cdk.Duration.seconds(props?.timeout || 30),
      environment: {
        CONFIG_PATH: configPath,
        ENVIRONMENT: env,
      },
      logRetention: logs.RetentionDays.ONE_WEEK,
      architecture: lambda.Architecture.X86_64,
    });

    // Lambda Function URL (simplest option)
    const functionUrl = mockApiFunction.addFunctionUrl({
      authType: lambda.FunctionUrlAuthType.NONE,
      cors: {
        allowedOrigins: ['*'],
        allowedMethods: [lambda.HttpMethod.ALL],
        allowedHeaders: ['*'],
      },
    });

    // API Gateway REST API (for more control)
    const api = new apigateway.LambdaRestApi(this, 'BlandMockApiGateway', {
      handler: mockApiFunction,
      restApiName: `blandmockapi-${env}`,
      description: 'API Gateway for Bland Mock API',
      deployOptions: {
        stageName: env,
        loggingLevel: apigateway.MethodLoggingLevel.INFO,
        dataTraceEnabled: true,
        metricsEnabled: true,
      },
      proxy: true,
      defaultCorsPreflightOptions: {
        allowOrigins: apigateway.Cors.ALL_ORIGINS,
        allowMethods: apigateway.Cors.ALL_METHODS,
        allowHeaders: ['*'],
      },
      endpointConfiguration: {
        types: [apigateway.EndpointType.REGIONAL],
      },
    });

    // CloudWatch Logs for API Gateway
    const logGroup = new logs.LogGroup(this, 'ApiGatewayLogs', {
      logGroupName: `/aws/apigateway/blandmockapi-${env}`,
      retention: logs.RetentionDays.ONE_WEEK,
      removalPolicy: cdk.RemovalPolicy.DESTROY,
    });

    // Outputs
    this.functionUrl = new cdk.CfnOutput(this, 'FunctionUrl', {
      value: functionUrl.url,
      description: 'Lambda Function URL',
      exportName: `BlandMockApi-${env}-FunctionUrl`,
    });

    this.apiEndpoint = new cdk.CfnOutput(this, 'ApiEndpoint', {
      value: api.url,
      description: 'API Gateway endpoint URL',
      exportName: `BlandMockApi-${env}-ApiUrl`,
    });

    new cdk.CfnOutput(this, 'FunctionName', {
      value: mockApiFunction.functionName,
      description: 'Lambda function name',
      exportName: `BlandMockApi-${env}-FunctionName`,
    });

    new cdk.CfnOutput(this, 'FunctionArn', {
      value: mockApiFunction.functionArn,
      description: 'Lambda function ARN',
      exportName: `BlandMockApi-${env}-FunctionArn`,
    });

    // Tags
    cdk.Tags.of(this).add('Project', 'BlandMockApi');
    cdk.Tags.of(this).add('Environment', env);
    cdk.Tags.of(this).add('ManagedBy', 'CDK');
  }
}
