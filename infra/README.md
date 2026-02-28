# Auto Repair Shop — AWS Infrastructure (Terraform)

## Architecture

This Terraform configuration provisions the following AWS resources:

| Resource | Description |
|:---|:---|
| **VPC** | 3 AZs, public/private/database subnets, NAT gateway |
| **EKS** | Managed Kubernetes cluster with auto-scaling node group |
| **ECR** | Docker container registry with lifecycle policy (keep last 10 images) |
| **RDS** | PostgreSQL 16 with encryption, automated backups, private subnet |

## Prerequisites

- [AWS CLI](https://aws.amazon.com/cli/) configured with appropriate credentials
- [Terraform](https://www.terraform.io/downloads) >= 1.0
- [kubectl](https://kubernetes.io/docs/tasks/tools/)

## Quick Start

```bash
cd infra

# 1. Copy and fill in your variables
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with real values (db_password, jwt_secret, etc.)

# 2. Initialize Terraform
terraform init

# 3. Preview changes
terraform plan

# 4. Apply infrastructure
terraform apply

# 5. Configure kubectl
aws eks update-kubeconfig --region us-east-1 --name autorepair-cluster

# 6. Apply K8s manifests (skip namespace/configmap/secret — already created by Terraform)
kubectl apply -f ../k8s/deployment.yaml
kubectl apply -f ../k8s/service.yaml
kubectl apply -f ../k8s/hpa.yaml

# 7. Verify
kubectl get pods -n autorepair
```

## Tear Down

```bash
terraform destroy
```

## Important Notes

- **Never commit `terraform.tfvars`** — it contains sensitive values
- The RDS instance is deployed in private subnets (not publicly accessible)
- EKS uses IRSA (IAM Roles for Service Accounts) for secure access
- ECR images are scanned on push for vulnerabilities
- When using AWS, the K8s `configmap.yaml` and `secret.yaml` from `/k8s` are **replaced** by Terraform-managed resources that inject the real RDS endpoint
