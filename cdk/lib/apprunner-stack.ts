import * as cdk from 'aws-cdk-lib';
import * as apprunner from 'aws-cdk-lib/aws-apprunner';
import * as ecr from 'aws-cdk-lib/aws-ecr';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as ecrassets from 'aws-cdk-lib/aws-ecr-assets';
import { Construct } from 'constructs';

export interface AppRunnerStackProps extends cdk.StackProps {
  readonly environment?: string;
  readonly cpu?: string;
  readonly memory?: string;
  readonly port?: number;
  readonly autoScalingMaxConcurrency?: number;
  readonly autoScalingMaxSize?: number;
  readonly autoScalingMinSize?: number;
}

export class BlandMockApiAppRunnerStack extends cdk.Stack {
  public readonly serviceUrl: cdk.CfnOutput;

  constructor(scope: Construct, id: string, props?: AppRunnerStackProps) {
    super(scope, id, props);

    const env = props?.environment || 'dev';
    const port = props?.port || 8080;

    // Build and push Docker image to ECR
    const imageAsset = new ecrassets.DockerImageAsset(this, 'BlandMockApiImage', {
      directory: '../',
      file: 'Dockerfile',
    });

    // IAM role for App Runner to access ECR
    const accessRole = new iam.Role(this, 'AppRunnerEcrAccessRole', {
      assumedBy: new iam.ServicePrincipal('build.apprunner.amazonaws.com'),
      managedPolicies: [
        iam.ManagedPolicy.fromAwsManagedPolicyName('service-role/AWSAppRunnerServicePolicyForECRAccess'),
      ],
    });

    // IAM role for the App Runner service instance
    const instanceRole = new iam.Role(this, 'AppRunnerInstanceRole', {
      assumedBy: new iam.ServicePrincipal('tasks.apprunner.amazonaws.com'),
      description: 'Instance role for Bland Mock API App Runner service',
    });

    // App Runner service
    const service = new apprunner.CfnService(this, 'BlandMockApiService', {
      serviceName: `blandmockapi-${env}`,
      sourceConfiguration: {
        authenticationConfiguration: {
          accessRoleArn: accessRole.roleArn,
        },
        imageRepository: {
          imageIdentifier: imageAsset.imageUri,
          imageRepositoryType: 'ECR',
          imageConfiguration: {
            port: port.toString(),
            runtimeEnvironmentVariables: [
              {
                name: 'ENVIRONMENT',
                value: env,
              },
            ],
          },
        },
        autoDeploymentsEnabled: false,
      },
      instanceConfiguration: {
        cpu: props?.cpu || '0.25 vCPU',
        memory: props?.memory || '0.5 GB',
        instanceRoleArn: instanceRole.roleArn,
      },
      healthCheckConfiguration: {
        protocol: 'HTTP',
        path: '/health',
        interval: 10,
        timeout: 5,
        healthyThreshold: 1,
        unhealthyThreshold: 5,
      },
      autoScalingConfigurationArn: this.createAutoScalingConfiguration(
        env,
        props?.autoScalingMaxConcurrency,
        props?.autoScalingMaxSize,
        props?.autoScalingMinSize
      ).attrAutoScalingConfigurationArn,
    });

    service.node.addDependency(accessRole);
    service.node.addDependency(instanceRole);

    // Outputs
    this.serviceUrl = new cdk.CfnOutput(this, 'ServiceUrl', {
      value: `https://${service.attrServiceUrl}`,
      description: 'App Runner service URL',
      exportName: `BlandMockApi-${env}-AppRunnerUrl`,
    });

    new cdk.CfnOutput(this, 'ServiceId', {
      value: service.attrServiceId,
      description: 'App Runner service ID',
      exportName: `BlandMockApi-${env}-AppRunnerServiceId`,
    });

    new cdk.CfnOutput(this, 'ServiceArn', {
      value: service.attrServiceArn,
      description: 'App Runner service ARN',
      exportName: `BlandMockApi-${env}-AppRunnerServiceArn`,
    });

    // Tags
    cdk.Tags.of(this).add('Project', 'BlandMockApi');
    cdk.Tags.of(this).add('Environment', env);
    cdk.Tags.of(this).add('ManagedBy', 'CDK');
  }

  private createAutoScalingConfiguration(
    env: string,
    maxConcurrency?: number,
    maxSize?: number,
    minSize?: number
  ): apprunner.CfnAutoScalingConfiguration {
    return new apprunner.CfnAutoScalingConfiguration(this, 'AutoScalingConfig', {
      autoScalingConfigurationName: `blandmockapi-${env}-autoscaling`,
      maxConcurrency: maxConcurrency || 100,
      maxSize: maxSize || 10,
      minSize: minSize || 1,
    });
  }
}
