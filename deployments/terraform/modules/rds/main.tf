# Terraform module: Phoenix RDS Database
# Created by Abhinav as part of Infrastructure as Code tasks

variable "identifier" {
  description = "Identifier for the RDS instance"
  type        = string
  default     = "phoenix-db"
}

variable "engine" {
  description = "Database engine type"
  type        = string
  default     = "postgres"
}

variable "engine_version" {
  description = "Database engine version"
  type        = string
  default     = "14.6"
}

variable "instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.t3.medium"
}

variable "allocated_storage" {
  description = "Allocated storage in GB"
  type        = number
  default     = 20
}

variable "max_allocated_storage" {
  description = "Maximum storage allocation in GB for autoscaling"
  type        = number
  default     = 100
}

variable "storage_type" {
  description = "Storage type"
  type        = string
  default     = "gp2"
}

variable "storage_encrypted" {
  description = "Enable storage encryption"
  type        = bool
  default     = true
}

variable "db_name" {
  description = "Database name"
  type        = string
  default     = "phoenix"
}

variable "username" {
  description = "Database admin username"
  type        = string
  default     = "postgres"
}

variable "password" {
  description = "Database admin password"
  type        = string
  sensitive   = true
}

variable "port" {
  description = "Database port"
  type        = number
  default     = 5432
}

variable "vpc_id" {
  description = "VPC ID where the DB will be deployed"
  type        = string
}

variable "subnet_ids" {
  description = "Subnet IDs for the DB subnet group"
  type        = list(string)
}

variable "parameter_group_family" {
  description = "DB parameter group family"
  type        = string
  default     = "postgres14"
}

variable "backup_retention_period" {
  description = "Backup retention period in days"
  type        = number
  default     = 7
}

variable "backup_window" {
  description = "Preferred backup window"
  type        = string
  default     = "03:00-04:00" # UTC
}

variable "maintenance_window" {
  description = "Preferred maintenance window"
  type        = string
  default     = "sun:04:00-sun:05:00" # UTC
}

variable "multi_az" {
  description = "Enable Multi-AZ deployment"
  type        = bool
  default     = false
}

variable "skip_final_snapshot" {
  description = "Skip final snapshot when deleting DB"
  type        = bool
  default     = false
}

variable "deletion_protection" {
  description = "Enable deletion protection"
  type        = bool
  default     = true
}

variable "apply_immediately" {
  description = "Apply changes immediately"
  type        = bool
  default     = false
}

variable "monitoring_interval" {
  description = "Enhanced monitoring interval in seconds"
  type        = number
  default     = 60
}

variable "enabled_cloudwatch_logs_exports" {
  description = "List of log types to export to CloudWatch"
  type        = list(string)
  default     = ["postgresql", "upgrade"]
}

variable "tags" {
  description = "Map of tags to apply to resources"
  type        = map(string)
  default     = {}
}

variable "allowed_cidr_blocks" {
  description = "CIDR blocks that can access the database"
  type        = list(string)
  default     = []
}

# DB Subnet Group
resource "aws_db_subnet_group" "phoenix" {
  name       = "${var.identifier}-subnet-group"
  subnet_ids = var.subnet_ids

  tags = merge(
    var.tags,
    {
      "Name" = "${var.identifier}-subnet-group"
    }
  )
}

# Security Group
resource "aws_security_group" "db_sg" {
  name        = "${var.identifier}-sg"
  description = "Security group for ${var.identifier} RDS instance"
  vpc_id      = var.vpc_id

  ingress {
    description = "Database port"
    from_port   = var.port
    to_port     = var.port
    protocol    = "tcp"
    cidr_blocks = var.allowed_cidr_blocks
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = merge(
    var.tags,
    {
      "Name" = "${var.identifier}-sg"
    }
  )
}

# Parameter Group
resource "aws_db_parameter_group" "phoenix" {
  name        = "${var.identifier}-pg"
  family      = var.parameter_group_family
  description = "Parameter group for ${var.identifier} RDS instance"

  parameter {
    name  = "log_connections"
    value = "1"
  }

  parameter {
    name  = "log_disconnections"
    value = "1"
  }

  parameter {
    name  = "log_checkpoints"
    value = "1"
  }

  parameter {
    name  = "log_min_duration_statement"
    value = "1000" # milliseconds
  }

  tags = merge(
    var.tags,
    {
      "Name" = "${var.identifier}-pg"
    }
  )
}

# Option Group (Postgres doesn't use option groups)

# RDS Instance
resource "aws_db_instance" "phoenix" {
  identifier                  = var.identifier
  engine                      = var.engine
  engine_version              = var.engine_version
  instance_class              = var.instance_class
  allocated_storage           = var.allocated_storage
  max_allocated_storage       = var.max_allocated_storage
  storage_type                = var.storage_type
  storage_encrypted           = var.storage_encrypted
  db_name                     = var.db_name
  username                    = var.username
  password                    = var.password
  port                        = var.port
  vpc_security_group_ids      = [aws_security_group.db_sg.id]
  db_subnet_group_name        = aws_db_subnet_group.phoenix.name
  parameter_group_name        = aws_db_parameter_group.phoenix.name
  backup_retention_period     = var.backup_retention_period
  backup_window               = var.backup_window
  maintenance_window          = var.maintenance_window
  multi_az                    = var.multi_az
  skip_final_snapshot         = var.skip_final_snapshot
  final_snapshot_identifier   = var.skip_final_snapshot ? null : "${var.identifier}-final-snapshot"
  deletion_protection         = var.deletion_protection
  apply_immediately           = var.apply_immediately
  monitoring_interval         = var.monitoring_interval
  enabled_cloudwatch_logs_exports = var.enabled_cloudwatch_logs_exports
  
  performance_insights_enabled = true
  performance_insights_retention_period = 7

  tags = merge(
    var.tags,
    {
      "Name" = var.identifier
    }
  )

  lifecycle {
    prevent_destroy = true
  }
}

# Outputs
output "db_instance_endpoint" {
  description = "The connection endpoint"
  value       = aws_db_instance.phoenix.endpoint
}

output "db_instance_id" {
  description = "The RDS instance ID"
  value       = aws_db_instance.phoenix.id
}

output "db_instance_arn" {
  description = "The ARN of the RDS instance"
  value       = aws_db_instance.phoenix.arn
}

output "db_instance_name" {
  description = "The database name"
  value       = aws_db_instance.phoenix.db_name
}

output "db_instance_username" {
  description = "The master username for the database"
  value       = aws_db_instance.phoenix.username
  sensitive   = true
}

output "db_instance_port" {
  description = "The database port"
  value       = aws_db_instance.phoenix.port
}

output "db_subnet_group_id" {
  description = "The DB subnet group name"
  value       = aws_db_subnet_group.phoenix.id
}

output "db_security_group_id" {
  description = "The security group ID"
  value       = aws_security_group.db_sg.id
}
