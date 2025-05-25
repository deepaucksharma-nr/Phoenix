# Phoenix Cloud Deployment Guide

## Overview

Phoenix can be deployed to major cloud providers (AWS, Azure) using container services. This guide covers deployment to AWS ECS and Azure Container Instances with full production configurations.

## Table of Contents

- [Prerequisites](#prerequisites)
- [AWS Deployment](#aws-deployment)
- [Azure Deployment](#azure-deployment)
- [Configuration Options](#configuration-options)
- [Monitoring & Operations](#monitoring--operations)
- [Cost Optimization](#cost-optimization)
- [Troubleshooting](#troubleshooting)

## Prerequisites

### Required Tools
- **Cloud CLI**: AWS CLI or Azure CLI
- **Terraform**: >= 1.3.0
- **Docker**: For building and managing containers
- **Docker Context**: For cloud deployments

### Cloud Permissions
Ensure you have sufficient permissions to create:
- VPCs/VNets and subnets
- Container instances and services
- Load balancers
- Storage accounts/S3 buckets
- IAM roles and policies

## AWS Deployment

### Quick Start
```bash
# Set environment variables
export AWS_REGION=us-east-1
export CLUSTER_NAME=phoenix-ecs
export ENVIRONMENT=dev

# Deploy Phoenix to AWS
./deploy-aws.sh
```

### Architecture on AWS

```
┌─────────────────────────────────────────────────────────────┐
│                        AWS Account                          │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────┐   │
│  │                    VPC (10.0.0.0/16)                 │   │
│  ├─────────────────────────────────────────────────────┤   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐ │   │
│  │  │  Public      │  │  Public     │  │  Public     │ │   │
│  │  │  Subnet 1    │  │  Subnet 2   │  │  Subnet 3   │ │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘ │   │
│  │         │                │                │          │   │
│  │    ┌────┴───────────┬────┴───────────┬────┴──────┐  │   │
│  │    │          NAT Gateway            │           │  │   │
│  │    └────┬───────────┴────┬───────────┴────┬──────┘  │   │
│  │         │                │                │          │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐ │   │
│  │  │  Private     │  │  Private    │  │  Private    │ │   │
│  │  │  Subnet 1    │  │  Subnet 2   │  │  Subnet 3   │ │   │
│  │  │              │  │             │  │             │ │   │
│  │  │  ECS Tasks   │  │  ECS Tasks  │  │  ECS Tasks  │ │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘ │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  ┌─────────────────┐  ┌─────────────┐  ┌──────────────┐   │
│  │   S3 Bucket     │  │     EFS     │  │  CloudWatch  │   │
│  │  (Phoenix Data) │  │  (Storage)  │  │  (Logs)      │   │
│  └─────────────────┘  └─────────────┘  └──────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### AWS-Specific Features

1. **Network Load Balancer (NLB)**
   - High-performance OTLP ingestion
   - Cross-zone load balancing
   - Static IP addresses available

2. **EBS Storage**
   - GP3 volumes for better performance
   - Automatic volume expansion
   - Snapshot support

3. **IAM Task Roles**
   - Fine-grained permissions
   - No credential management
   - Automatic rotation

4. **CloudWatch Integration**
   - Native metrics export
   - Log aggregation
   - Alarms and notifications

### Advanced AWS Configuration

```hcl
# terraform.tfvars for production
aws_region = "us-east-1"
environment = "prod"
cluster_name = "phoenix-ecs-prod"

# High availability across 3 AZs
private_subnet_cidrs = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
public_subnet_cidrs = ["10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"]

# Production task configuration
task_instance_types = ["m5.xlarge", "m5.2xlarge"]
task_capacity_type = "ON_DEMAND"
task_min_count = 3
task_max_count = 20
task_desired_count = 6

# Enable all monitoring
enable_monitoring = true
```

## Azure Deployment

### Quick Start
```bash
# Set environment variables
export AZURE_LOCATION=eastus
export RESOURCE_GROUP=phoenix-vnext-rg
export CLUSTER_NAME=phoenix-aci
export ENVIRONMENT=dev

# Deploy Phoenix to Azure
./deploy-azure.sh
```

### Architecture on Azure

```
┌─────────────────────────────────────────────────────────────┐
│                    Azure Subscription                       │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────┐   │
│  │              Resource Group (phoenix-rg)             │   │
│  ├─────────────────────────────────────────────────────┤   │
│  │  ┌─────────────────────────────────────────────┐    │   │
│  │  │          Virtual Network (10.0.0.0/16)      │    │   │
│  │  ├─────────────────────────────────────────────┤    │   │
│  │  │  ┌─────────────────┐  ┌─────────────────┐  │    │   │
│  │  │  │   ACI Subnet    │  │ Ingress Subnet  │  │    │   │
│  │  │  │   10.0.1.0/24   │  │  10.0.2.0/24    │  │    │   │
│  │  │  │                 │  │                 │  │    │   │
│  │  │  │  ACI Containers │  │  Load Balancer  │  │    │   │
│  │  │  └─────────────────┘  └─────────────────┘  │    │   │
│  │  └─────────────────────────────────────────────┘    │   │
│  │                                                      │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────┐  │   │
│  │  │ Storage Acct │  │  Container   │  │   Log    │  │   │
│  │  │  (Metrics)   │  │   Registry   │  │Analytics │  │   │
│  │  └──────────────┘  └──────────────┘  └──────────┘  │   │
│  │                                                      │   │
│  │  ┌─────────────────────────────────────────────┐    │   │
│  │  │          ACI Container Group                │    │   │
│  │  │  ┌────────────┐  ┌────────────┐            │    │   │
│  │  │  │  General   │  │ Monitoring │            │    │   │
│  │  │  │ Containers │  │ Containers │            │    │   │
│  │  │  └────────────┘  └────────────┘            │    │   │
│  │  └─────────────────────────────────────────────┘    │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### Azure-Specific Features

1. **Azure Load Balancer**
   - Standard SKU with zone redundancy
   - Health probes for reliability
   - Azure Private Link support

2. **Azure Files Storage**
   - Shared storage for control signals
   - Premium and Standard tiers
   - SMB and NFS support

3. **Managed Identity**
   - Managed identity for containers
   - Azure RBAC integration
   - Key Vault access

4. **Azure Monitor Integration**
   - Container insights
   - Log Analytics workspace
   - Application Insights

### Advanced Azure Configuration

```hcl
# terraform.tfvars for production
azure_location = "eastus"
environment = "prod"
cluster_name = "phoenix-aci-prod"

# Production container configuration
container_vm_size = "Standard_D8s_v3"
container_count = 6
container_min_count = 3
container_max_count = 20

# Enable monitoring containers
enable_monitoring_containers = true

# Azure AD integration
aci_admin_group_ids = ["your-aad-group-id"]
```

## Configuration Options

### Docker Compose Override

Create a custom override file for your deployment:

```yaml
# docker-compose.override.yaml
version: '3.8'
services:
  otelcol-main:
    deploy:
      replicas: 5
      resources:
        limits:
          memory: 4G
          cpus: '4'
        reservations:
          memory: 1G
          cpus: '1'
  
  prometheus:
    command:
      - '--storage.tsdb.retention.time=90d'
      - '--storage.tsdb.path=/prometheus'
      - '--config.file=/etc/prometheus/prometheus.yml'
    volumes:
      - prometheus-data:/prometheus:rw

  grafana:
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=your-secure-password
      - GF_SERVER_DOMAIN=grafana.your-domain.com

volumes:
  prometheus-data:
    driver: local
    driver_opts:
      type: none
      device: /data/prometheus
      o: bind
```

Deploy with custom configuration:
```bash
docker-compose -f docker-compose.yaml -f docker-compose.override.yaml up -d
```

### Environment Variables

Key environment variables for cloud deployments:

```bash
# AWS
export AWS_REGION=us-east-1
export AWS_PROFILE=production
export CLUSTER_NAME=phoenix-ecs-prod

# Azure
export AZURE_SUBSCRIPTION_ID=your-subscription-id
export AZURE_LOCATION=eastus
export RESOURCE_GROUP=phoenix-prod-rg

# Common
export ENVIRONMENT=prod
export ENABLE_MONITORING=true
export ENABLE_BACKUPS=true
```

## Monitoring & Operations

### Accessing Dashboards

#### AWS
```bash
# Access via Load Balancer
aws elbv2 describe-load-balancers --names phoenix-nlb

# Get service endpoints
docker context use aws-ecs
docker ps
```

#### Azure
```bash
# Get Container Group IP
az container show --resource-group phoenix-rg --name phoenix-containers --query ipAddress.ip

# Access via public IP
open http://<CONTAINER_IP>:3000
```

### Operational Tasks

#### Scaling
```bash
# Scale collector instances (AWS)
aws ecs update-service --cluster phoenix-ecs --service phoenix-collector --desired-count 10

# Scale container group (Azure)
az container-instances update --resource-group phoenix-rg --name phoenix-containers --cpu 4 --memory 8
```

#### Backup Control Signals
```bash
# AWS - Backup to S3
docker exec phoenix-actuator \
  aws s3 cp /etc/phoenix/control/optimization_mode.yaml \
  s3://phoenix-backups/control/$(date +%Y%m%d-%H%M%S).yaml

# Azure - Backup to Blob Storage
docker exec phoenix-actuator \
  az storage blob upload \
    --account-name $STORAGE_ACCOUNT \
    --container-name backups \
    --name control/$(date +%Y%m%d-%H%M%S).yaml \
    --file /etc/phoenix/control/optimization_mode.yaml
```

## Cost Optimization

### AWS Cost Savings
1. **Use Spot Instances** for non-critical tasks
2. **Reserved Capacity** for production workloads
3. **S3 Lifecycle Policies** for old metrics
4. **GP3 volumes** for storage
5. **Single NAT Gateway** for dev environments

### Azure Cost Savings
1. **Reserved Instances** for container instances
2. **Spot Instances** for batch workloads
3. **Standard tier** storage for archives
4. **Auto-scaling** to match demand
5. **Azure Hybrid Benefit** if applicable

### Resource Recommendations

| Environment | Tasks | Instance Type | Storage | Cost/Month |
|-------------|-------|---------------|---------|------------|
| Dev | 3 | t3.large / D2s_v3 | 100GB | ~$300 |
| Staging | 6 | t3.xlarge / D4s_v3 | 500GB | ~$800 |
| Production | 12 | m5.2xlarge / D8s_v3 | 2TB | ~$3000 |

## Troubleshooting

### Common Issues

#### Containers Not Starting
```bash
# Check container status
docker ps -a
docker logs <container-name>

# Check service logs
docker-compose logs -f <service-name>
```

#### Load Balancer Issues
```bash
# AWS - Check load balancer
aws elbv2 describe-load-balancers --names phoenix-nlb
aws ecs describe-services --cluster phoenix-ecs --services phoenix-collector

# Azure - Check container group
az container show --resource-group phoenix-rg --name phoenix-containers
```

#### Storage Issues
```bash
# Check volume mounts
docker volume ls
docker volume inspect <volume-name>

# Check disk space
df -h
docker system df
```

### Debug Commands

```bash
# Get all Phoenix containers
docker-compose ps

# Check collector logs
docker-compose logs -f otelcol-main

# Check control loop
docker-compose logs -f control-loop-actuator

# Exec into container
docker exec -it phoenix-collector sh
```

## Security Best Practices

1. **Network Isolation**: Restrict container communication
2. **IAM Roles**: Use least-privilege access
3. **Secrets Management**: Use cloud KMS for sensitive data
4. **Image Scanning**: Scan containers for vulnerabilities
5. **Audit Logging**: Enable container audit logs
6. **Encryption**: Enable encryption at rest and in transit

## Disaster Recovery

### Backup Strategy
- Control signals: Every 15 minutes
- Prometheus data: Daily snapshots
- Grafana dashboards: Version controlled
- Configuration: Stored in Git

### Recovery Procedures
1. Restore infrastructure with Terraform
2. Deploy Phoenix with Docker Compose
3. Restore control signals from backup
4. Import Prometheus snapshots
5. Verify system functionality

## Support

For issues or questions:
- GitHub Issues: https://github.com/deepaucksharma/Phoenix/issues
- Documentation: https://github.com/deepaucksharma/Phoenix/docs
- Community Slack: #phoenix-support