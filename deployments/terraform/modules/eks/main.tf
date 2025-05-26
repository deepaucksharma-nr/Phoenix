# Terraform module: Phoenix Kubernetes Cluster
# Created by Abhinav as part of Infrastructure as Code tasks

variable "cluster_name" {
  description = "Name of the Kubernetes cluster"
  type        = string
  default     = "phoenix-cluster"
}

variable "cluster_version" {
  description = "Kubernetes version"
  type        = string
  default     = "1.25"
}

variable "region" {
  description = "Cloud provider region"
  type        = string
}

variable "node_groups" {
  description = "Map of node group configurations"
  type = map(object({
    name         = string
    min_size     = number
    max_size     = number
    desired_size = number
    instance_type = string
    labels       = map(string)
    taints       = list(object({
      key    = string
      value  = string
      effect = string
    }))
  }))
  default = {
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
}

variable "vpc_id" {
  description = "VPC ID where Kubernetes cluster will be created"
  type        = string
}

variable "subnet_ids" {
  description = "List of subnet IDs to use for Kubernetes nodes"
  type        = list(string)
}

variable "tags" {
  description = "Tags to apply to resources"
  type        = map(string)
  default     = {}
}

# EKS Cluster
resource "aws_eks_cluster" "phoenix" {
  name     = var.cluster_name
  role_arn = aws_iam_role.cluster_role.arn
  version  = var.cluster_version

  vpc_config {
    subnet_ids         = var.subnet_ids
    security_group_ids = [aws_security_group.cluster_sg.id]
  }

  tags = merge(
    var.tags,
    {
      "Name" = var.cluster_name
      "Platform" = "Phoenix"
    }
  )

  depends_on = [
    aws_iam_role_policy_attachment.cluster_AmazonEKSClusterPolicy,
    aws_iam_role_policy_attachment.cluster_AmazonEKSVPCResourceController,
  ]
}

# IAM Role for EKS Cluster
resource "aws_iam_role" "cluster_role" {
  name = "${var.cluster_name}-cluster-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "eks.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      }
    ]
  })

  tags = var.tags
}

resource "aws_iam_role_policy_attachment" "cluster_AmazonEKSClusterPolicy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.cluster_role.name
}

resource "aws_iam_role_policy_attachment" "cluster_AmazonEKSVPCResourceController" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSVPCResourceController"
  role       = aws_iam_role.cluster_role.name
}

# Security Group for EKS Cluster
resource "aws_security_group" "cluster_sg" {
  name        = "${var.cluster_name}-cluster-sg"
  description = "Security group for Kubernetes cluster"
  vpc_id      = var.vpc_id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = merge(
    var.tags,
    {
      "Name" = "${var.cluster_name}-cluster-sg"
    }
  )
}

# Node Groups
resource "aws_eks_node_group" "phoenix" {
  for_each = var.node_groups

  cluster_name    = aws_eks_cluster.phoenix.name
  node_group_name = "${var.cluster_name}-${each.value.name}"
  node_role_arn   = aws_iam_role.node_role.arn
  subnet_ids      = var.subnet_ids

  scaling_config {
    desired_size = each.value.desired_size
    max_size     = each.value.max_size
    min_size     = each.value.min_size
  }

  instance_types = [each.value.instance_type]

  labels = merge(
    each.value.labels,
    {
      "nodegroup" = each.value.name
    }
  )

  dynamic "taint" {
    for_each = each.value.taints
    content {
      key    = taint.value.key
      value  = taint.value.value
      effect = taint.value.effect
    }
  }

  tags = merge(
    var.tags,
    {
      "Name" = "${var.cluster_name}-${each.value.name}"
    }
  )

  depends_on = [
    aws_iam_role_policy_attachment.node_AmazonEKSWorkerNodePolicy,
    aws_iam_role_policy_attachment.node_AmazonEKS_CNI_Policy,
    aws_iam_role_policy_attachment.node_AmazonEC2ContainerRegistryReadOnly,
  ]
}

# IAM Role for Node Groups
resource "aws_iam_role" "node_role" {
  name = "${var.cluster_name}-node-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      }
    ]
  })

  tags = var.tags
}

resource "aws_iam_role_policy_attachment" "node_AmazonEKSWorkerNodePolicy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy"
  role       = aws_iam_role.node_role.name
}

resource "aws_iam_role_policy_attachment" "node_AmazonEKS_CNI_Policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"
  role       = aws_iam_role.node_role.name
}

resource "aws_iam_role_policy_attachment" "node_AmazonEC2ContainerRegistryReadOnly" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
  role       = aws_iam_role.node_role.name
}

# Outputs
output "cluster_id" {
  description = "EKS cluster ID"
  value       = aws_eks_cluster.phoenix.id
}

output "cluster_endpoint" {
  description = "Endpoint for EKS control plane"
  value       = aws_eks_cluster.phoenix.endpoint
}

output "cluster_security_group_id" {
  description = "Security group ID attached to the EKS cluster"
  value       = aws_security_group.cluster_sg.id
}

output "cluster_arn" {
  description = "ARN of the EKS cluster"
  value       = aws_eks_cluster.phoenix.arn
}

output "node_groups" {
  description = "Node group resources"
  value       = aws_eks_node_group.phoenix
}
