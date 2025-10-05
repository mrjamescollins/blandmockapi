import * as cdk from 'aws-cdk-lib';
import * as ec2 from 'aws-cdk-lib/aws-ec2';
import * as ecs from 'aws-cdk-lib/aws-ecs';
import * as ecsPatterns from 'aws-cdk-lib/aws-ecs-patterns';
import * as logs from 'aws-cdk-lib/aws-logs';
import * as elbv2 from 'aws-cdk-lib/aws-elasticloadbalancingv2';
import { Construct } from 'constructs';

export interface EcsStackProps extends cdk.StackProps {
  readonly environment?: string;
  readonly vpcId?: string;
  readonly desiredCount?: number;
  readonly cpu?: number;
  readonly memoryLimitMiB?: number;
}

export class BlandMockApiEcsStack extends cdk.Stack {
  public readonly serviceUrl: cdk.CfnOutput;
  public readonly cluster: ecs.ICluster;

  constructor(scope: Construct, id: string, props?: EcsStackProps) {
    super(scope, id, props);

    const env = props?.environment || 'dev';

    // VPC - use existing or create new
    const vpc = props?.vpcId
      ? ec2.Vpc.fromLookup(this, 'Vpc', { vpcId: props.vpcId })
      : new ec2.Vpc(this, 'BlandMockApiVpc', {
          maxAzs: 2,
          natGateways: 1,
          subnetConfiguration: [
            {
              cidrMask: 24,
              name: 'public',
              subnetType: ec2.SubnetType.PUBLIC,
            },
            {
              cidrMask: 24,
              name: 'private',
              subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS,
            },
          ],
        });

    // ECS Cluster
    const cluster = new ecs.Cluster(this, 'BlandMockApiCluster', {
      clusterName: `blandmockapi-${env}`,
      vpc,
      containerInsights: true,
    });

    this.cluster = cluster;

    // Fargate Service with ALB
    const fargateService = new ecsPatterns.ApplicationLoadBalancedFargateService(
      this,
      'BlandMockApiFargateService',
      {
        cluster,
        serviceName: `blandmockapi-${env}`,
        desiredCount: props?.desiredCount || 2,
        cpu: props?.cpu || 256,
        memoryLimitMiB: props?.memoryLimitMiB || 512,
        taskImageOptions: {
          image: ecs.ContainerImage.fromAsset('../', {
            file: 'Dockerfile',
          }),
          containerName: 'blandmockapi',
          containerPort: 8080,
          environment: {
            ENVIRONMENT: env,
          },
          logDriver: ecs.LogDrivers.awsLogs({
            streamPrefix: 'blandmockapi',
            logRetention: logs.RetentionDays.ONE_WEEK,
          }),
        },
        publicLoadBalancer: true,
        healthCheckGracePeriod: cdk.Duration.seconds(60),
        circuitBreaker: {
          rollback: true,
        },
      }
    );

    // Configure health check
    fargateService.targetGroup.configureHealthCheck({
      path: '/health',
      interval: cdk.Duration.seconds(30),
      timeout: cdk.Duration.seconds(5),
      healthyThresholdCount: 2,
      unhealthyThresholdCount: 3,
      healthyHttpCodes: '200',
    });

    // Auto-scaling based on CPU
    const scaling = fargateService.service.autoScaleTaskCount({
      minCapacity: props?.desiredCount || 2,
      maxCapacity: 10,
    });

    scaling.scaleOnCpuUtilization('CpuScaling', {
      targetUtilizationPercent: 70,
      scaleInCooldown: cdk.Duration.seconds(60),
      scaleOutCooldown: cdk.Duration.seconds(60),
    });

    // Auto-scaling based on memory
    scaling.scaleOnMemoryUtilization('MemoryScaling', {
      targetUtilizationPercent: 80,
      scaleInCooldown: cdk.Duration.seconds(60),
      scaleOutCooldown: cdk.Duration.seconds(60),
    });

    // Outputs
    this.serviceUrl = new cdk.CfnOutput(this, 'LoadBalancerUrl', {
      value: `http://${fargateService.loadBalancer.loadBalancerDnsName}`,
      description: 'Application Load Balancer URL',
      exportName: `BlandMockApi-${env}-AlbUrl`,
    });

    new cdk.CfnOutput(this, 'ClusterName', {
      value: cluster.clusterName,
      description: 'ECS Cluster name',
      exportName: `BlandMockApi-${env}-ClusterName`,
    });

    new cdk.CfnOutput(this, 'ServiceName', {
      value: fargateService.service.serviceName,
      description: 'ECS Service name',
      exportName: `BlandMockApi-${env}-ServiceName`,
    });

    new cdk.CfnOutput(this, 'ServiceArn', {
      value: fargateService.service.serviceArn,
      description: 'ECS Service ARN',
      exportName: `BlandMockApi-${env}-ServiceArn`,
    });

    // Tags
    cdk.Tags.of(this).add('Project', 'BlandMockApi');
    cdk.Tags.of(this).add('Environment', env);
    cdk.Tags.of(this).add('ManagedBy', 'CDK');
  }
}
