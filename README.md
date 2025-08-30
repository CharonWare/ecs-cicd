# ecs-cicd

A basic CICD pipeline that targets ECR/ECS. Just a hobby project, nothing serious.

It pulls a GitHub repo, builds a docker image, and pushes it to ECR. Deployment to ECS is a feature in progress.

Required environment variables:

|Variable name|Description|Required|
|-------------|-----------|--------|
|PROJECT|The name of the GitHub project, this is [organisation/username]/repository format. For example: `CharonWare/investment-calculator`|Yes|
|BRANCH|The branch to build, defailts to `main` if not provided|No|
|PAT_TOKEN|Your GitHub PAT token, it requires read access to the project repository at minimum|Yes|
|ECR|The ECR URI to push to, for example: `12345678.dkr.ecr.eu-west-1.amazonaws.com/my-repo`|Yes|
|AWS_DEFAULT_REGION|The AWS region for your ECR repo|Yes|