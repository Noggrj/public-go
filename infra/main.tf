# ============================================================
# Auto Repair Shop API — AWS Infrastructure
# Provisions: VPC, EKS Cluster, RDS PostgreSQL, ECR Repository
# ============================================================

data "aws_availability_zones" "available" {
  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

locals {
  azs = slice(data.aws_availability_zones.available.names, 0, 3)
}

# ---------------------------------------------------------------
# VPC
# ---------------------------------------------------------------
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 5.5"

  name = "${var.cluster_name}-vpc"
  cidr = var.vpc_cidr

  azs              = local.azs
  public_subnets   = [for k, v in local.azs : cidrsubnet(var.vpc_cidr, 8, k)]
  private_subnets  = [for k, v in local.azs : cidrsubnet(var.vpc_cidr, 8, k + 10)]
  database_subnets = [for k, v in local.azs : cidrsubnet(var.vpc_cidr, 8, k + 20)]

  enable_nat_gateway   = true
  single_nat_gateway   = true
  enable_dns_hostnames = true
  enable_dns_support   = true

  create_database_subnet_group = true

  public_subnet_tags = {
    "kubernetes.io/role/elb"                      = 1
    "kubernetes.io/cluster/${var.cluster_name}"    = "owned"
  }

  private_subnet_tags = {
    "kubernetes.io/role/internal-elb"              = 1
    "kubernetes.io/cluster/${var.cluster_name}"    = "owned"
  }
}

# ---------------------------------------------------------------
# EKS Cluster (raw resources — compatible with AWS Academy)
# ---------------------------------------------------------------
data "aws_caller_identity" "current" {}

locals {
  lab_role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/LabRole"
}

resource "aws_eks_cluster" "this" {
  name     = var.cluster_name
  role_arn = local.lab_role_arn
  version  = var.cluster_version

  vpc_config {
    subnet_ids              = module.vpc.private_subnets
    endpoint_public_access  = true
    endpoint_private_access = false
  }

  depends_on = [module.vpc]
}

resource "aws_eks_node_group" "default" {
  cluster_name    = aws_eks_cluster.this.name
  node_group_name = "default"
  node_role_arn   = local.lab_role_arn
  subnet_ids      = module.vpc.private_subnets

  instance_types = var.node_instance_types

  scaling_config {
    desired_size = var.node_desired_size
    max_size     = var.node_max_size
    min_size     = var.node_min_size
  }

  labels = {
    role = "general"
  }

  depends_on = [aws_eks_cluster.this]
}

# ---------------------------------------------------------------
# ECR Repository (Docker image registry)
# ---------------------------------------------------------------
resource "aws_ecr_repository" "autorepair" {
  name                 = "autorepair"
  image_tag_mutability = "MUTABLE"
  force_delete         = true

  image_scanning_configuration {
    scan_on_push = true
  }

  encryption_configuration {
    encryption_type = "AES256"
  }
}

resource "aws_ecr_lifecycle_policy" "autorepair" {
  repository = aws_ecr_repository.autorepair.name

  policy = jsonencode({
    rules = [
      {
        rulePriority = 1
        description  = "Keep last 10 images"
        selection = {
          tagStatus   = "any"
          countType   = "imageCountMoreThan"
          countNumber = 10
        }
        action = {
          type = "expire"
        }
      }
    ]
  })
}

# ---------------------------------------------------------------
# RDS PostgreSQL
# ---------------------------------------------------------------
resource "aws_security_group" "rds" {
  name_prefix = "autorepair-rds-"
  vpc_id      = module.vpc.vpc_id

  ingress {
    description     = "PostgreSQL from EKS"
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [aws_eks_cluster.this.vpc_config[0].cluster_security_group_id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_db_instance" "autorepair" {
  identifier             = "autorepair-db"
  engine                 = "postgres"
  engine_version         = "16.6"
  instance_class         = var.db_instance_class
  allocated_storage      = 20
  max_allocated_storage  = 20
  storage_encrypted      = false

  db_name  = var.db_name
  username = var.db_username
  password = var.db_password

  db_subnet_group_name   = module.vpc.database_subnet_group_name
  vpc_security_group_ids = [aws_security_group.rds.id]

  multi_az               = false
  publicly_accessible    = false
  skip_final_snapshot    = true
  deletion_protection    = false

  backup_retention_period = 0
  maintenance_window      = "Mon:04:00-Mon:05:00"

  tags = {
    Name = "autorepair-db"
  }
}

# ---------------------------------------------------------------
# Kubernetes Namespace + Secrets (applied via K8s provider)
# ---------------------------------------------------------------
resource "kubernetes_namespace" "autorepair" {
  metadata {
    name = "autorepair"
    labels = {
      app         = "autorepair"
      environment = var.environment
    }
  }

  depends_on = [aws_eks_node_group.default]
}

resource "kubernetes_secret" "autorepair" {
  metadata {
    name      = "autorepair-secret"
    namespace = kubernetes_namespace.autorepair.metadata[0].name
  }

  data = {
    DB_PASSWORD = var.db_password
    JWT_SECRET  = var.jwt_secret
  }

  depends_on = [kubernetes_namespace.autorepair]
}

resource "kubernetes_config_map" "autorepair" {
  metadata {
    name      = "autorepair-config"
    namespace = kubernetes_namespace.autorepair.metadata[0].name
  }

  data = {
    PORT       = "8080"
    DB_HOST    = aws_db_instance.autorepair.address
    DB_PORT    = "5432"
    DB_NAME    = var.db_name
    DB_USER    = var.db_username
    DB_SSLMODE = "require"
  }

  depends_on = [kubernetes_namespace.autorepair]
}
