# ecs-cicd

A basic CICD pipeline that targets ECR/ECS. Just a hobby project, nothing serious.

It pulls a GitHub repo, builds a docker image, and pushes it to ECR. Deployment to ECS is a feature in progress.

Currently the application relies on you having local AWS CLI access set up, but for portability I will expand on this in the future.

Required environment variables:

|Variable name|Description|Required|
|-------------|-----------|--------|
|PROJECT|The name of the GitHub project, this is [organisation/username]/repository format. For example: `CharonWare/investment-calculator`|Yes|
|BRANCH|The branch to build, defailts to `main` if not provided|No|
|PAT_TOKEN|Your GitHub PAT token, required if your target repository is private. It requires read access to the project repository at minimum|No|
|ECR|The ECR URI to push to, for example: `12345678.dkr.ecr.eu-west-1.amazonaws.com/my-repo`|Yes|