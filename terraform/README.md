# Terraform — ECS Infrastructure

Provisions everything needed to run the Jobs API on AWS ECS Fargate in
`us-east-1`: networking, ALB, RDS Postgres, IAM, secrets, and the ECS
service itself. The configuration is split into two layers that are
applied in order.

---


### Why two layers

The remote state backend (S3 bucket + DynamoDB lock table) and the ECR
repo are
prerequisites for `terraform/main/`'s backend config and for ECS to have
an image to point at. Putting them in their own `bootstrap/` layer with
**local state** breaks the chicken-and-egg cleanly: bootstrap is applied
once, then the main layer takes over with state stored in the bucket
bootstrap created.

### What lives where

| File | Resources | Why it's separate |
|---|---|---|
| `bootstrap/main.tf` | S3 state bucket, DynamoDB lock, ECR repo | Prerequisites for everything else; rarely change |
| `main/vpc.tf` | VPC, subnets (2 public + 2 private), IGW, NAT GW, route tables | Networking foundation |
| `main/security_groups.tf` | `alb-sg`, `ecs-sg`, `rds-sg` | Least-privilege chain; one ingress rule per layer |
| `main/alb.tf` | ALB, target group, HTTP listener | Public entry point — health check on `/healthz` |
| `main/rds.tf` | DB subnet group, `pg16` parameter group (`rds.force_ssl=1`), Postgres instance | Private-subnet only, encrypted, single-AZ for cost |
| `main/secrets.tf` | `random_password`, Secrets Manager secret + version | Generates and stores the DB password |
| `main/iam.tf` | Task execution role, task role, GitHub OIDC provider + role | Every policy ARN-scoped; wildcards only where AWS demands |
| `main/ecs.tf` | Cluster, task definition, service | Fargate; ignores image/desired-count drift so CI can deploy |
| `main/logs.tf` | CloudWatch log group `/ecs/test-ecs/app` | 7-day retention |
| `main/outputs.tf` | ALB DNS, ECR URL, cluster/service names, GH role ARN, secret ARN | Consumed by CI workflow |

---

## Prerequisites

- **Terraform** ≥ 1.6 (`terraform -version`)
- **AWS CLI v2** configured for an account with permissions to create
  VPC / ECS / RDS / IAM / Secrets Manager / S3 / DynamoDB / ECR
  resources in `us-east-1`. Easiest is an admin role; least-privilege
  for running Terraform itself is left as an enhancement.
- An AWS account ID handy (used in the state bucket name).

Quick sanity check:

```bash
aws sts get-caller-identity
aws configure get region 
```

---

## Apply order

### 1. Bootstrap (one-time, local state)

```bash
cd terraform/bootstrap
terraform validate
terraform init
terraform apply
```

What this creates:

- `s3://test-ecs-tfstate-<your-account-id>`
- DynamoDB table — pay-per-request, PITR on.
- ECR repository — immutable tags, scan on push.

Save the `backend_config_hint` output — it tells you exactly what to
pass to the main layer's `terraform init`.

> The bootstrap state file (`terraform/bootstrap/terraform.tfstate`) lives
> on your machine and is gitignored. If you lose it, the resources stay
> alive but become unmanaged — re-importable with `terraform import`.

### 2. Main layer (remote state)

```bash
cd ../main
terraform validate
terraform init \
  -backend-config="bucket=test-ecs-tfstate-938413640052" \
  -backend-config="region=us-east-1" \
  -backend-config="dynamodb_table=test-ecs-tfstate-lock"

terraform plan
terraform apply
```

Replace `<your-account-id>` with the value from
`aws sts get-caller-identity --query Account --output text`.

### 3. First-time smoke test

```bash
curl "http://$(terraform output -raw alb_dns_name)/"
```

Until CI pushes the real image, the ALB returns the nginx welcome page.
That's the signal that the network path (Internet → ALB → ECS task →
container) is wired correctly. Once CI runs, the same URL will return
the Jobs API JSON.

---

## Outputs

After `terraform apply` in `main/`:

| Output | Used for |
|---|---|
| `alb_dns_name` | Smoke test, eventual DNS record target |
| `ecr_repository_url` | Tag and push images from CI: `<url>:<commit-sha>` |
| `ecs_cluster_name` | `aws ecs update-service --cluster <name>` |
| `ecs_service_name` | `aws ecs update-service --service <name>` |
| `ecs_task_family` | `aws ecs register-task-definition --family <family>` |
| `github_actions_role_arn` | Set as a GitHub Actions secret/var; used in `aws-actions/configure-aws-credentials` |
| `db_secret_arn` | Reference if a separate one-shot migration task is added |
| `log_group_name` | CloudWatch console / `aws logs tail` |

```bash
terraform output -raw alb_dns_name
terraform output -raw github_actions_role_arn
```

---

## Variables

Everything has a sensible default in `variables.tf`. The defaults reflect
the cost-optimised "test-ecs" environment:

| Variable | Default |
|---|---|
| `region` | `us-east-1` |
| `project` / `environment` | `test-ecs` / `test` |
| `vpc_cidr` | `10.0.0.0/16` |
| `task_cpu` / `task_memory` | `256` / `512` |
| `desired_count` | `1` |
| `db_instance_class` | `db.t4g.micro` |
| `db_allocated_storage` | `20` |
| `placeholder_image` | `public.ecr.aws/nginx/nginx:stable` |
| `image_tag` | `""` (use placeholder; CI overrides) |
| `github_org` / `github_repo` / `github_branch` | `GithinjiC` / `mpesa-devops-assessment` / `main` |
| `log_retention_days` | `7` |

Override by copying `terraform.tfvars.example` to `terraform.tfvars` and
editing only what you need (the `.tfvars` file is gitignored).

---

## Tearing it all down

```bash
cd terraform/main
terraform destroy

cd ../bootstrap
terraform destroy
```
---

