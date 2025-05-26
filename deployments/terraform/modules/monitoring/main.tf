# Terraform module: Phoenix Monitoring Stack
# Created by Abhinav as part of Infrastructure as Code tasks

variable "name_prefix" {
  description = "Prefix to use for resource names"
  type        = string
  default     = "phoenix"
}

variable "vpc_id" {
  description = "VPC ID where resources will be deployed"
  type        = string
}

variable "subnet_ids" {
  description = "Subnet IDs where resources will be deployed"
  type        = list(string)
}

variable "instance_type" {
  description = "EC2 instance type for Prometheus"
  type        = string
  default     = "t3.medium"
}

variable "prometheus_version" {
  description = "Prometheus version"
  type        = string
  default     = "2.41.0"
}

variable "grafana_version" {
  description = "Grafana version"
  type        = string
  default     = "9.3.6"
}

variable "grafana_admin_password" {
  description = "Grafana admin password"
  type        = string
  sensitive   = true
}

variable "prometheus_retention_days" {
  description = "Prometheus data retention in days"
  type        = number
  default     = 15
}

variable "monitoring_secrets" {
  description = "Map of secrets for monitoring components"
  type        = map(string)
  default     = {}
  sensitive   = true
}

variable "tags" {
  description = "Map of tags to apply to resources"
  type        = map(string)
  default     = {}
}

# Security group for monitoring instances
resource "aws_security_group" "monitoring" {
  name        = "${var.name_prefix}-monitoring-sg"
  description = "Security group for monitoring stack"
  vpc_id      = var.vpc_id

  ingress {
    description = "Prometheus UI"
    from_port   = 9090
    to_port     = 9090
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/8"]
  }

  ingress {
    description = "Prometheus metrics"
    from_port   = 9100
    to_port     = 9100
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/8"]
  }

  ingress {
    description = "Grafana UI"
    from_port   = 3000
    to_port     = 3000
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/8"]
  }

  ingress {
    description = "SSH"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/8"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = merge(var.tags, {
    Name = "${var.name_prefix}-monitoring-sg"
  })
}

# IAM role for monitoring instances
resource "aws_iam_role" "monitoring" {
  name = "${var.name_prefix}-monitoring-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      }
    ]
  })

  tags = var.tags
}

# IAM instance profile
resource "aws_iam_instance_profile" "monitoring" {
  name = "${var.name_prefix}-monitoring-profile"
  role = aws_iam_role.monitoring.name
}

# IAM policy for EC2 discovery and CloudWatch metrics
resource "aws_iam_policy" "monitoring" {
  name        = "${var.name_prefix}-monitoring-policy"
  description = "Policy for monitoring instances"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "ec2:DescribeInstances",
          "ec2:DescribeTags"
        ]
        Effect   = "Allow"
        Resource = "*"
      },
      {
        Action = [
          "cloudwatch:GetMetricData",
          "cloudwatch:ListMetrics"
        ]
        Effect   = "Allow"
        Resource = "*"
      }
    ]
  })
}

# Attach policy to role
resource "aws_iam_role_policy_attachment" "monitoring" {
  role       = aws_iam_role.monitoring.name
  policy_arn = aws_iam_policy.monitoring.arn
}

# EBS volume for Prometheus
resource "aws_ebs_volume" "prometheus_data" {
  availability_zone = data.aws_subnet.selected.availability_zone
  size              = 100
  type              = "gp3"
  
  tags = merge(var.tags, {
    Name = "${var.name_prefix}-prometheus-data"
  })
}

# Get info about the first subnet
data "aws_subnet" "selected" {
  id = var.subnet_ids[0]
}

# User data script
locals {
  user_data = <<-EOF
    #!/bin/bash
    set -e

    # Set up environment
    echo "Setting up monitoring environment..."
    mkdir -p /opt/prometheus /opt/grafana /opt/prometheus/data

    # Install Docker
    apt-get update
    apt-get install -y apt-transport-https ca-certificates curl software-properties-common
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
    add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
    apt-get update
    apt-get install -y docker-ce docker-ce-cli containerd.io
    systemctl enable docker
    systemctl start docker

    # Install Docker Compose
    curl -L "https://github.com/docker/compose/releases/download/1.29.2/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    chmod +x /usr/local/bin/docker-compose

    # Mount EBS volume for data
    device_name=$(ls -la /dev/disk/by-id | grep ${aws_ebs_volume.prometheus_data.id} | awk '{print $11}' | sed 's/\.\.\/\.\.\///' || echo "")
    if [ -n "$device_name" ]; then
      mkfs -t ext4 "$device_name"
      mount "$device_name" /opt/prometheus/data
      echo "$device_name /opt/prometheus/data ext4 defaults,nofail 0 2" >> /etc/fstab
    else
      echo "EBS volume not found, using instance storage."
    fi

    # Create Prometheus config
    cat > /opt/prometheus/prometheus.yml << 'EOFP'
    global:
      scrape_interval: 15s
      evaluation_interval: 15s

    scrape_configs:
      - job_name: 'prometheus'
        static_configs:
          - targets: ['localhost:9090']

      - job_name: 'node_exporter'
        ec2_sd_configs:
          - region: ${data.aws_region.current.name}
            port: 9100
        relabel_configs:
          - source_labels: [__meta_ec2_tag_Name]
            target_label: instance

      - job_name: 'phoenix_services'
        kubernetes_sd_configs:
          - role: pod
        relabel_configs:
          - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
            action: keep
            regex: true
          - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
            action: replace
            target_label: __metrics_path__
            regex: (.+)
          - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]
            action: replace
            regex: (.+):(?:\d+);(\d+)
            replacement: $${1}:$${2}
            target_label: __address__
          - source_labels: [__meta_kubernetes_namespace]
            action: replace
            target_label: kubernetes_namespace
          - source_labels: [__meta_kubernetes_pod_name]
            action: replace
            target_label: kubernetes_pod_name
    EOFP

    # Create docker-compose file
    cat > /opt/docker-compose.yml << 'EOFD'
    version: '3.8'

    services:
      prometheus:
        image: prom/prometheus:${var.prometheus_version}
        volumes:
          - /opt/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml
          - /opt/prometheus/data:/prometheus
        command:
          - '--config.file=/etc/prometheus/prometheus.yml'
          - '--storage.tsdb.path=/prometheus'
          - '--storage.tsdb.retention.time=${var.prometheus_retention_days}d'
          - '--web.console.libraries=/etc/prometheus/console_libraries'
          - '--web.console.templates=/etc/prometheus/consoles'
          - '--web.enable-lifecycle'
        ports:
          - 9090:9090
        restart: always

      grafana:
        image: grafana/grafana:${var.grafana_version}
        volumes:
          - /opt/grafana:/var/lib/grafana
        environment:
          - GF_SECURITY_ADMIN_PASSWORD=${var.grafana_admin_password}
          - GF_USERS_ALLOW_SIGN_UP=false
          - GF_SERVER_DOMAIN=monitoring.phoenix.local
        ports:
          - 3000:3000
        depends_on:
          - prometheus
        restart: always

      node_exporter:
        image: prom/node-exporter:latest
        command:
          - '--path.procfs=/host/proc'
          - '--path.sysfs=/host/sys'
          - '--collector.filesystem.ignored-mount-points="^/(sys|proc|dev|host|etc)($$|/)"'
        volumes:
          - /proc:/host/proc:ro
          - /sys:/host/sys:ro
          - /:/rootfs:ro
        ports:
          - 9100:9100
        restart: always
    EOFD

    # Start monitoring stack
    cd /opt
    docker-compose up -d

    # Set up auto-updates
    cat > /etc/cron.daily/monitoring-update << 'EOFU'
    #!/bin/bash
    cd /opt
    docker-compose pull
    docker-compose up -d
    EOFU
    chmod +x /etc/cron.daily/monitoring-update

    echo "Monitoring setup complete!"
  EOF
}

# Get current region
data "aws_region" "current" {}

# Monitoring EC2 instance
resource "aws_instance" "monitoring" {
  ami                    = data.aws_ami.ubuntu.id
  instance_type          = var.instance_type
  subnet_id              = var.subnet_ids[0]
  vpc_security_group_ids = [aws_security_group.monitoring.id]
  iam_instance_profile   = aws_iam_instance_profile.monitoring.name
  user_data              = local.user_data
  user_data_replace_on_change = true

  root_block_device {
    volume_size = 20
    volume_type = "gp3"
    encrypted   = true
  }

  tags = merge(var.tags, {
    Name = "${var.name_prefix}-monitoring"
  })

  volume_tags = merge(var.tags, {
    Name = "${var.name_prefix}-monitoring-root"
  })
}

# Volume attachment
resource "aws_volume_attachment" "monitoring_data" {
  device_name  = "/dev/sdf"
  volume_id    = aws_ebs_volume.prometheus_data.id
  instance_id  = aws_instance.monitoring.id
  skip_destroy = true
}

# Latest Ubuntu AMI
data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"] # Canonical

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

# Create Route53 record for monitoring tools (optional)
resource "aws_route53_record" "monitoring" {
  count = var.create_dns ? 1 : 0
  
  zone_id = var.dns_zone_id
  name    = "monitoring.${var.dns_domain}"
  type    = "A"
  ttl     = "300"
  records = [aws_instance.monitoring.private_ip]
}

variable "create_dns" {
  description = "Whether to create DNS records"
  type        = bool
  default     = false
}

variable "dns_zone_id" {
  description = "Route53 DNS zone ID"
  type        = string
  default     = ""
}

variable "dns_domain" {
  description = "DNS domain for monitoring"
  type        = string
  default     = ""
}

# Outputs
output "prometheus_url" {
  description = "Prometheus URL"
  value       = "http://${aws_instance.monitoring.private_ip}:9090"
}

output "grafana_url" {
  description = "Grafana URL"
  value       = "http://${aws_instance.monitoring.private_ip}:3000"
}

output "instance_id" {
  description = "ID of the monitoring EC2 instance"
  value       = aws_instance.monitoring.id
}

output "private_ip" {
  description = "Private IP address of the monitoring instance"
  value       = aws_instance.monitoring.private_ip
}

output "security_group_id" {
  description = "Security group ID for monitoring"
  value       = aws_security_group.monitoring.id
}
