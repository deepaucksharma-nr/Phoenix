# Development Environment Terraform Configuration
# Created by Abhinav as part of Infrastructure as Code task

provider "aws" {
  region = var.aws_region
}

locals {
  environment = "dev"
  name_prefix = "phoenix-${local.environment}"
  
  common_tags = {
    Environment = "Development"
    Project     = "Phoenix"
    ManagedBy   = "Terraform"
    Team        = "Platform"
  }
}

# Use for remote state (uncomment for production use)
# terraform {
#   backend "s3" {
#     bucket         = "phoenix-terraform-state"
#     key            = "dev/terraform.tfstate"
#     region         = "us-west-2"
#     dynamodb_table = "phoenix-terraform-locks"
#     encrypt        = true
#   }
# }

# Variables
variable "aws_region" {
  description = "AWS region to deploy resources"
  type        = string
  default     = "us-west-2"
}

variable "vpc_cidr" {
  description = "VPC CIDR block"
  type        = string
  default     = "10.0.0.0/16"
}

variable "availability_zones" {
  description = "List of availability zones"
  type        = list(string)
  default     = ["us-west-2a", "us-west-2b"]
}

variable "db_password" {
  description = "Database password"
  type        = string
  sensitive   = true
  default     = "CHANGE_ME_IN_TERRAFORM_TFVARS" # Should be set in terraform.tfvars
}

variable "grafana_password" {
  description = "Grafana admin password"
  type        = string
  sensitive   = true
  default     = "CHANGE_ME_IN_TERRAFORM_TFVARS" # Should be set in terraform.tfvars
}

# VPC and Networking
module "vpc" {
  source = "terraform-aws-modules/vpc/aws"
  version = "~> 3.0"

  name = "${local.name_prefix}-vpc"
  cidr = var.vpc_cidr

  azs             = var.availability_zones
  private_subnets = [for i, az in var.availability_zones : cidrsubnet(var.vpc_cidr, 8, i)]
  public_subnets  = [for i, az in var.availability_zones : cidrsubnet(var.vpc_cidr, 8, i + 100)]

  enable_nat_gateway = true
  single_nat_gateway = true
  enable_vpn_gateway = false

  private_subnet_tags = {
    "kubernetes.io/role/internal-elb" = 1
  }

  public_subnet_tags = {
    "kubernetes.io/role/elb" = 1
  }

  tags = local.common_tags
}

# EKS Cluster
module "eks" {
  source = "../../modules/eks"

  cluster_name    = "${local.name_prefix}-cluster"
  cluster_version = "1.24"
  region          = var.aws_region
  vpc_id          = module.vpc.vpc_id
  subnet_ids      = module.vpc.private_subnets

  node_groups = {
    default = {
      name         = "default"
      min_size     = 2
      max_size     = 4
      desired_size = 2
      instance_type = "t3.medium"
      labels       = { role = "worker" }
      taints       = []
    }
  }

  tags = local.common_tags
}

# RDS Database
module "rds" {
  source = "../../modules/rds"

  identifier        = "${local.name_prefix}-db"
  engine            = "postgres"
  engine_version    = "14.6"
  instance_class    = "db.t3.medium"
  allocated_storage = 20
  db_name           = "phoenix"
  username          = "postgres"
  password          = var.db_password
  port              = 5432

  vpc_id            = module.vpc.vpc_id
  subnet_ids        = module.vpc.private_subnets
  multi_az          = false
  skip_final_snapshot = true # For dev environment only
  deletion_protection = false # For dev environment only

  allowed_cidr_blocks = [var.vpc_cidr]

  tags = local.common_tags
}

# Monitoring
module "monitoring" {
  source = "../../modules/monitoring"

  name_prefix           = local.name_prefix
  vpc_id                = module.vpc.vpc_id
  subnet_ids            = module.vpc.private_subnets
  instance_type         = "t3.medium"
  grafana_admin_password = var.grafana_password

  tags = local.common_tags
}

# Outputs
output "cluster_endpoint" {
  description = "EKS cluster endpoint"
  value       = module.eks.cluster_endpoint
}

output "cluster_id" {
  description = "EKS cluster ID"
  value       = module.eks.cluster_id
}

output "db_endpoint" {
  description = "RDS database endpoint"
  value       = module.rds.db_instance_endpoint
  sensitive   = true
}

output "prometheus_url" {
  description = "Prometheus URL"
  value       = module.monitoring.prometheus_url
}

output "grafana_url" {
  description = "Grafana URL"
  value       = module.monitoring.grafana_url
}
