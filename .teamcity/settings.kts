import jetbrains.buildServer.configs.kotlin.*
import jetbrains.buildServer.configs.kotlin.buildSteps.script
import jetbrains.buildServer.configs.kotlin.triggers.vcs
import jetbrains.buildServer.configs.kotlin.buildFeatures.perfmon
import jetbrains.buildServer.configs.kotlin.buildFeatures.golang

version = "2024.03"

project {
    description = "Bland Mock API - Lightweight configurable REST and GraphQL mock server"

    buildType(UnitTests)
    buildType(IntegrationTests)
    buildType(BuildDocker)
    buildType(DeployLambda)
    buildType(DeployECS)
    buildType(DeployAppRunner)

    params {
        param("env.GO_VERSION", "1.25")
        param("env.AWS_REGION", "us-east-1")
        param("env.ENVIRONMENT", "dev")
    }
}

object UnitTests : BuildType({
    name = "Unit Tests"
    description = "Run unit tests with coverage"

    vcs {
        root(DslContext.settingsRoot)
    }

    steps {
        script {
            name = "Download Dependencies"
            scriptContent = "go mod download"
        }
        script {
            name = "Run Unit Tests"
            scriptContent = """
                go test ./internal/... -v -cover -coverprofile=coverage.out
                go tool cover -func=coverage.out
            """.trimIndent()
        }
    }

    triggers {
        vcs {
            branchFilter = "+:*"
        }
    }

    features {
        perfmon {
        }
        golang {
            testFormat = "json"
        }
    }

    artifactRules = """
        coverage.out => coverage-reports/
        coverage.html => coverage-reports/
    """.trimIndent()
})

object IntegrationTests : BuildType({
    name = "Integration Tests"
    description = "Run integration tests"

    vcs {
        root(DslContext.settingsRoot)
    }

    dependencies {
        snapshot(UnitTests) {
        }
    }

    steps {
        script {
            name = "Run Integration Tests"
            scriptContent = """
                go test -tags=integration ./test/integration/... -v -timeout 60s
            """.trimIndent()
        }
    }

    triggers {
        vcs {
            branchFilter = "+:main"
        }
    }
})

object BuildDocker : BuildType({
    name = "Build Docker Images"
    description = "Build Docker images for all deployment targets"

    vcs {
        root(DslContext.settingsRoot)
    }

    dependencies {
        snapshot(UnitTests) {
        }
        snapshot(IntegrationTests) {
        }
    }

    steps {
        script {
            name = "Build Standard Docker Image"
            scriptContent = """
                docker build -t blandmockapi:latest .
                docker tag blandmockapi:latest %env.AWS_ACCOUNT_ID%.dkr.ecr.%env.AWS_REGION%.amazonaws.com/blandmockapi:latest
            """.trimIndent()
        }
        script {
            name = "Build Lambda Docker Image"
            scriptContent = """
                docker build -f Dockerfile.lambda -t blandmockapi:lambda .
                docker tag blandmockapi:lambda %env.AWS_ACCOUNT_ID%.dkr.ecr.%env.AWS_REGION%.amazonaws.com/blandmockapi:lambda
            """.trimIndent()
        }
        script {
            name = "Push to ECR"
            scriptContent = """
                aws ecr get-login-password --region %env.AWS_REGION% | docker login --username AWS --password-stdin %env.AWS_ACCOUNT_ID%.dkr.ecr.%env.AWS_REGION%.amazonaws.com
                docker push %env.AWS_ACCOUNT_ID%.dkr.ecr.%env.AWS_REGION%.amazonaws.com/blandmockapi:latest
                docker push %env.AWS_ACCOUNT_ID%.dkr.ecr.%env.AWS_REGION%.amazonaws.com/blandmockapi:lambda
            """.trimIndent()
        }
    }
})

object DeployLambda : BuildType({
    name = "Deploy to Lambda"
    description = "Deploy to AWS Lambda using CDK"

    vcs {
        root(DslContext.settingsRoot)
    }

    dependencies {
        snapshot(BuildDocker) {
        }
    }

    steps {
        script {
            name = "Install CDK Dependencies"
            scriptContent = """
                cd cdk
                npm install
            """.trimIndent()
        }
        script {
            name = "Deploy Lambda Stack"
            scriptContent = """
                cd cdk
                npm run cdk deploy BlandMockApiLambdaStack-%env.ENVIRONMENT% -- --require-approval never
            """.trimIndent()
        }
        script {
            name = "Test Deployment"
            scriptContent = """
                FUNCTION_URL=${'$'}(aws lambda get-function-url-config --function-name blandmockapi-%env.ENVIRONMENT% --query 'FunctionUrl' --output text)
                curl -f "${'$'}FUNCTION_URL/health" || exit 1
            """.trimIndent()
        }
    }

    params {
        param("env.STACK_NAME", "BlandMockApiLambdaStack-%env.ENVIRONMENT%")
    }
})

object DeployECS : BuildType({
    name = "Deploy to ECS"
    description = "Deploy to AWS ECS using CDK"

    vcs {
        root(DslContext.settingsRoot)
    }

    dependencies {
        snapshot(BuildDocker) {
        }
    }

    steps {
        script {
            name = "Install CDK Dependencies"
            scriptContent = """
                cd cdk
                npm install
            """.trimIndent()
        }
        script {
            name = "Deploy ECS Stack"
            scriptContent = """
                cd cdk
                npm run cdk deploy BlandMockApiEcsStack-%env.ENVIRONMENT% -- --require-approval never
            """.trimIndent()
        }
        script {
            name = "Wait for Service Stability"
            scriptContent = """
                aws ecs wait services-stable --cluster blandmockapi-%env.ENVIRONMENT% --services blandmockapi-%env.ENVIRONMENT%
            """.trimIndent()
        }
        script {
            name = "Test Deployment"
            scriptContent = """
                ALB_URL=${'$'}(aws cloudformation describe-stacks --stack-name BlandMockApiEcsStack-%env.ENVIRONMENT% --query 'Stacks[0].Outputs[?OutputKey==`LoadBalancerUrl`].OutputValue' --output text)
                curl -f "${'$'}ALB_URL/health" || exit 1
            """.trimIndent()
        }
    }
})

object DeployAppRunner : BuildType({
    name = "Deploy to App Runner"
    description = "Deploy to AWS App Runner using CDK"

    vcs {
        root(DslContext.settingsRoot)
    }

    dependencies {
        snapshot(BuildDocker) {
        }
    }

    steps {
        script {
            name = "Install CDK Dependencies"
            scriptContent = """
                cd cdk
                npm install
            """.trimIndent()
        }
        script {
            name = "Deploy App Runner Stack"
            scriptContent = """
                cd cdk
                npm run cdk deploy BlandMockApiAppRunnerStack-%env.ENVIRONMENT% -- --require-approval never
            """.trimIndent()
        }
        script {
            name = "Wait for Service"
            scriptContent = """
                SERVICE_ARN=${'$'}(aws cloudformation describe-stacks --stack-name BlandMockApiAppRunnerStack-%env.ENVIRONMENT% --query 'Stacks[0].Outputs[?OutputKey==`ServiceArn`].OutputValue' --output text)

                # Wait for service to be running
                for i in {1..30}; do
                    STATUS=${'$'}(aws apprunner describe-service --service-arn ${'$'}SERVICE_ARN --query 'Service.Status' --output text)
                    if [ "${'$'}STATUS" == "RUNNING" ]; then
                        break
                    fi
                    echo "Waiting for service to be running... (attempt ${'$'}i/30)"
                    sleep 10
                done
            """.trimIndent()
        }
        script {
            name = "Test Deployment"
            scriptContent = """
                SERVICE_URL=${'$'}(aws cloudformation describe-stacks --stack-name BlandMockApiAppRunnerStack-%env.ENVIRONMENT% --query 'Stacks[0].Outputs[?OutputKey==`ServiceUrl`].OutputValue' --output text)
                curl -f "${'$'}SERVICE_URL/health" || exit 1
            """.trimIndent()
        }
    }
})
