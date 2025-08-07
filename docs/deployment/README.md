# Deployment Guide

## Overview

Hướng dẫn này cung cấp thông tin chi tiết về cách triển khai Zplus SaaS Base lên các môi trường khác nhau.

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Environment Setup](#environment-setup)
- [Kubernetes Deployment](#kubernetes-deployment)
- [CI/CD Pipeline](#cicd-pipeline)
- [Monitoring & Observability](#monitoring--observability)
- [Security](#security)
- [Backup & Recovery](#backup--recovery)
- [Troubleshooting](#troubleshooting)

## Architecture Overview

### Production Architecture
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   CloudFlare    │    │   Load Balancer │    │   Kubernetes    │
│      CDN        │◄──►│     (ALB)       │◄──►│    Cluster      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │                        │
                       ┌─────────────────┐    ┌─────────────────┐
                       │   Traefik       │    │   Services      │
                       │   Gateway       │◄──►│   & Databases   │
                       └─────────────────┘    └─────────────────┘
```

### Service Topology
```
api.zplus.io          ──► API Gateway (Traefik)
app.zplus.io          ──► Frontend (Next.js)
auth.zplus.io         ──► Keycloak
grafana.zplus.io      ──► Grafana
prometheus.zplus.io   ──► Prometheus
```

## Environment Setup

### 1. Infrastructure Requirements

#### AWS EKS Cluster
```yaml
# eksctl cluster config
apiVersion: eksctl.io/v1alpha5
kind: ClusterConfig

metadata:
  name: zplus-production
  region: us-west-2
  version: "1.28"

nodeGroups:
  - name: worker-nodes
    instanceType: t3.large
    minSize: 3
    maxSize: 10
    desiredCapacity: 5
    ssh:
      enableSsm: true
    iam:
      withAddonPolicies:
        autoScaler: true
        cloudWatch: true
        ebs: true

addons:
  - name: vpc-cni
  - name: coredns
  - name: kube-proxy
  - name: aws-ebs-csi-driver
```

#### Terraform Infrastructure
```hcl
# terraform/main.tf
module "eks" {
  source = "terraform-aws-modules/eks/aws"
  
  cluster_name    = "zplus-production"
  cluster_version = "1.28"
  
  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets
  
  eks_managed_node_groups = {
    main = {
      instance_types = ["t3.large"]
      min_size       = 3
      max_size       = 10
      desired_size   = 5
    }
  }
}

module "rds" {
  source = "terraform-aws-modules/rds/aws"
  
  identifier = "zplus-postgres"
  engine     = "postgres"
  engine_version = "16"
  family     = "postgres16"
  
  allocated_storage     = 100
  max_allocated_storage = 1000
  storage_type          = "gp3"
  
  db_name  = "zplus"
  username = "postgres"
  
  backup_retention_period = 7
  backup_window          = "03:00-04:00"
  maintenance_window     = "Sun:04:00-Sun:05:00"
  
  multi_az = true
}
```

### 2. Prerequisites

```bash
# Required tools
kubectl version --client    # >= 1.28
helm version               # >= 3.12
argocd version             # >= 2.8
terraform version          # >= 1.5
```

### 3. Environment Variables

```bash
# Production environment
export ENVIRONMENT=production
export CLUSTER_NAME=zplus-production
export AWS_REGION=us-west-2
export DOMAIN=zplus.io

# Database URLs
export DATABASE_URL=postgres://user:pass@rds-endpoint:5432/zplus
export MONGODB_URL=mongodb://user:pass@docdb-endpoint:27017/zplus
export REDIS_URL=redis://elasticache-endpoint:6379

# External services
export KEYCLOAK_URL=https://auth.zplus.io
export S3_BUCKET=zplus-production-files
```

## Kubernetes Deployment

### 1. Namespace Setup

```yaml
# k8s/namespaces.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: zplus-production
  labels:
    name: zplus-production
    environment: production
---
apiVersion: v1
kind: Namespace
metadata:
  name: zplus-monitoring
  labels:
    name: zplus-monitoring
    environment: production
```

### 2. Helm Charts

#### Application Chart
```yaml
# helm/zplus-api/values.production.yaml
replicaCount: 3

image:
  repository: 123456789.dkr.ecr.us-west-2.amazonaws.com/zplus-api
  tag: "v1.0.0"
  pullPolicy: Always

resources:
  limits:
    cpu: 1000m
    memory: 1Gi
  requests:
    cpu: 500m
    memory: 512Mi

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 20
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80

service:
  type: ClusterIP
  port: 8080

ingress:
  enabled: true
  className: traefik
  annotations:
    traefik.ingress.kubernetes.io/router.tls: "true"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
  hosts:
    - host: api.zplus.io
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: api-zplus-io-tls
      hosts:
        - api.zplus.io

env:
  - name: DATABASE_URL
    valueFrom:
      secretKeyRef:
        name: database-secret
        key: url
  - name: REDIS_URL
    valueFrom:
      secretKeyRef:
        name: redis-secret
        key: url
```

#### Database Chart
```yaml
# helm/postgresql/values.production.yaml
auth:
  postgresPassword: ${POSTGRES_PASSWORD}
  username: zplus
  password: ${ZPLUS_PASSWORD}
  database: zplus

primary:
  persistence:
    enabled: true
    size: 100Gi
    storageClass: gp3
  
  resources:
    limits:
      memory: 4Gi
      cpu: 2000m
    requests:
      memory: 2Gi
      cpu: 1000m

backup:
  enabled: true
  schedule: "0 2 * * *"
  retention: 7d
```

### 3. Deployment Commands

```bash
# Install/upgrade applications
helm upgrade --install zplus-api ./helm/zplus-api \
  -f ./helm/zplus-api/values.production.yaml \
  -n zplus-production

helm upgrade --install zplus-frontend ./helm/zplus-frontend \
  -f ./helm/zplus-frontend/values.production.yaml \
  -n zplus-production

# Install databases
helm upgrade --install postgresql bitnami/postgresql \
  -f ./helm/postgresql/values.production.yaml \
  -n zplus-production

helm upgrade --install mongodb bitnami/mongodb \
  -f ./helm/mongodb/values.production.yaml \
  -n zplus-production

helm upgrade --install redis bitnami/redis \
  -f ./helm/redis/values.production.yaml \
  -n zplus-production
```

## CI/CD Pipeline

### 1. GitHub Actions Workflow

```yaml
# .github/workflows/deploy-production.yml
name: Deploy to Production

on:
  push:
    branches: [main]
    tags: ['v*']

env:
  AWS_REGION: us-west-2
  EKS_CLUSTER_NAME: zplus-production

jobs:
  build-and-deploy:
    runs-on: ubuntu-latest
    
    steps:
    - name: Checkout code
      uses: actions/checkout@v4
      
    - name: Configure AWS credentials
      uses: aws-actions/configure-aws-credentials@v4
      with:
        aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
        aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
        aws-region: ${{ env.AWS_REGION }}
        
    - name: Login to Amazon ECR
      id: login-ecr
      uses: aws-actions/amazon-ecr-login@v2
      
    - name: Build and push backend image
      env:
        ECR_REGISTRY: ${{ steps.login-ecr.outputs.registry }}
        ECR_REPOSITORY: zplus-api
        IMAGE_TAG: ${{ github.sha }}
      run: |
        docker build -t $ECR_REGISTRY/$ECR_REPOSITORY:$IMAGE_TAG ./backend
        docker push $ECR_REGISTRY/$ECR_REPOSITORY:$IMAGE_TAG
        
    - name: Build and push frontend image  
      env:
        ECR_REGISTRY: ${{ steps.login-ecr.outputs.registry }}
        ECR_REPOSITORY: zplus-frontend
        IMAGE_TAG: ${{ github.sha }}
      run: |
        docker build -t $ECR_REGISTRY/$ECR_REPOSITORY:$IMAGE_TAG ./frontend
        docker push $ECR_REGISTRY/$ECR_REPOSITORY:$IMAGE_TAG
        
    - name: Deploy to EKS
      run: |
        aws eks get-token --cluster-name $EKS_CLUSTER_NAME | kubectl apply -f -
        kubectl set image deployment/zplus-api zplus-api=$ECR_REGISTRY/zplus-api:${{ github.sha }} -n zplus-production
        kubectl set image deployment/zplus-frontend zplus-frontend=$ECR_REGISTRY/zplus-frontend:${{ github.sha }} -n zplus-production
        kubectl rollout status deployment/zplus-api -n zplus-production
        kubectl rollout status deployment/zplus-frontend -n zplus-production
```

### 2. ArgoCD Application

```yaml
# argocd/application.yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: zplus-production
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/ilmsadmin/Zplus_SaaS_Base
    targetRevision: main
    path: helm/zplus-api
    helm:
      valueFiles:
        - values.production.yaml
  destination:
    server: https://kubernetes.default.svc
    namespace: zplus-production
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - CreateNamespace=true
```

## Monitoring & Observability

### 1. Prometheus Setup

```yaml
# monitoring/prometheus/values.yaml
prometheus:
  prometheusSpec:
    retention: 30d
    storageSpec:
      volumeClaimTemplate:
        spec:
          storageClassName: gp3
          accessModes: ["ReadWriteOnce"]
          resources:
            requests:
              storage: 100Gi

grafana:
  persistence:
    enabled: true
    size: 10Gi
    storageClassName: gp3
  
  ingress:
    enabled: true
    hosts:
      - grafana.zplus.io
    tls:
      - secretName: grafana-tls
        hosts:
          - grafana.zplus.io

alertmanager:
  alertmanagerSpec:
    storage:
      volumeClaimTemplate:
        spec:
          storageClassName: gp3
          accessModes: ["ReadWriteOnce"]
          resources:
            requests:
              storage: 10Gi
```

### 2. Alerting Rules

```yaml
# monitoring/alerts/api-alerts.yaml
groups:
  - name: api-alerts
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value }} for {{ $labels.instance }}"
          
      - alert: HighMemoryUsage
        expr: (container_memory_usage_bytes / container_spec_memory_limit_bytes) > 0.9
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High memory usage"
          description: "Memory usage is {{ $value }} for {{ $labels.pod }}"
```

### 3. Grafana Dashboards

```bash
# Import dashboards
kubectl create configmap api-dashboard \
  --from-file=./monitoring/dashboards/api-dashboard.json \
  -n zplus-monitoring

kubectl create configmap kubernetes-dashboard \
  --from-file=./monitoring/dashboards/kubernetes-dashboard.json \
  -n zplus-monitoring
```

## Security

### 1. Network Policies

```yaml
# security/network-policies.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: zplus-api-netpol
  namespace: zplus-production
spec:
  podSelector:
    matchLabels:
      app: zplus-api
  policyTypes:
    - Ingress
    - Egress
  ingress:
    - from:
        - namespaceSelector:
            matchLabels:
              name: traefik
      ports:
        - protocol: TCP
          port: 8080
  egress:
    - to:
        - namespaceSelector:
            matchLabels:
              name: zplus-production
      ports:
        - protocol: TCP
          port: 5432  # PostgreSQL
        - protocol: TCP
          port: 27017 # MongoDB
        - protocol: TCP
          port: 6379  # Redis
```

### 2. Pod Security Standards

```yaml
# security/pod-security.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: zplus-production
  labels:
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/audit: restricted
    pod-security.kubernetes.io/warn: restricted
```

### 3. RBAC

```yaml
# security/rbac.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: zplus-production
  name: zplus-role
rules:
  - apiGroups: [""]
    resources: ["pods", "services", "endpoints"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["apps"]
    resources: ["deployments"]
    verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: zplus-rolebinding
  namespace: zplus-production
subjects:
  - kind: ServiceAccount
    name: zplus-serviceaccount
    namespace: zplus-production
roleRef:
  kind: Role
  name: zplus-role
  apiGroup: rbac.authorization.k8s.io
```

## Backup & Recovery

### 1. Database Backup

```yaml
# backup/postgres-backup.yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: postgres-backup
  namespace: zplus-production
spec:
  schedule: "0 2 * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
            - name: postgres-backup
              image: postgres:16
              command:
                - /bin/bash
                - -c
                - |
                  pg_dump $DATABASE_URL | gzip > /backup/postgres-$(date +%Y%m%d-%H%M%S).sql.gz
                  aws s3 cp /backup/ s3://zplus-backups/postgres/ --recursive
              env:
                - name: DATABASE_URL
                  valueFrom:
                    secretKeyRef:
                      name: database-secret
                      key: url
              volumeMounts:
                - name: backup-volume
                  mountPath: /backup
          volumes:
            - name: backup-volume
              emptyDir: {}
          restartPolicy: OnFailure
```

### 2. Velero Backup

```bash
# Install Velero
velero install \
  --provider aws \
  --plugins velero/velero-plugin-for-aws:v1.8.0 \
  --bucket zplus-velero-backups \
  --backup-location-config region=us-west-2 \
  --snapshot-location-config region=us-west-2

# Create backup schedule
velero schedule create daily-backup \
  --schedule="0 3 * * *" \
  --include-namespaces zplus-production \
  --ttl 720h
```

### 3. Disaster Recovery Plan

#### RTO/RPO Targets
- **RTO (Recovery Time Objective)**: 4 hours
- **RPO (Recovery Point Objective)**: 1 hour

#### Recovery Procedures
```bash
# 1. Restore from Velero backup
velero restore create --from-backup daily-backup-20240101

# 2. Restore database from S3
aws s3 cp s3://zplus-backups/postgres/latest.sql.gz /tmp/
gunzip /tmp/latest.sql.gz
psql $DATABASE_URL < /tmp/latest.sql

# 3. Verify services
kubectl get pods -n zplus-production
kubectl get services -n zplus-production
```

## Troubleshooting

### Common Issues

#### 1. Pod Startup Issues
```bash
# Check pod status
kubectl get pods -n zplus-production

# Check pod logs
kubectl logs -f deployment/zplus-api -n zplus-production

# Describe pod for events
kubectl describe pod <pod-name> -n zplus-production
```

#### 2. Database Connection Issues
```bash
# Test database connectivity
kubectl run postgres-client --rm -it --image postgres:16 -- psql $DATABASE_URL

# Check network policies
kubectl get networkpolicies -n zplus-production
```

#### 3. High Memory/CPU Usage
```bash
# Check resource usage
kubectl top pods -n zplus-production
kubectl top nodes

# Check HPA status
kubectl get hpa -n zplus-production
```

### Health Check Endpoints

```bash
# API health check
curl https://api.zplus.io/health

# Database health
kubectl exec -it deployment/postgresql -n zplus-production -- pg_isready

# Redis health
kubectl exec -it deployment/redis -n zplus-production -- redis-cli ping
```

### Log Analysis

```bash
# Aggregate logs with stern
stern zplus-api -n zplus-production

# Query logs in Loki
curl -G -s "http://loki:3100/loki/api/v1/query" \
  --data-urlencode 'query={namespace="zplus-production"}'
```

## Rollback Procedures

### Application Rollback
```bash
# Rollback deployment
kubectl rollout undo deployment/zplus-api -n zplus-production

# Rollback to specific revision
kubectl rollout undo deployment/zplus-api --to-revision=2 -n zplus-production

# Check rollout status
kubectl rollout status deployment/zplus-api -n zplus-production
```

### Database Rollback
```bash
# Restore from backup
pg_restore -d $DATABASE_URL /backup/postgres-backup.sql

# Run database migration rollback
kubectl exec -it deployment/zplus-api -n zplus-production -- \
  ./migrate -path migrations -database $DATABASE_URL down 1
```

## Performance Tuning

### Database Optimization
```sql
-- PostgreSQL tuning
ALTER SYSTEM SET shared_buffers = '256MB';
ALTER SYSTEM SET effective_cache_size = '1GB';
ALTER SYSTEM SET maintenance_work_mem = '64MB';
ALTER SYSTEM SET checkpoint_completion_target = 0.7;
ALTER SYSTEM SET wal_buffers = '16MB';
ALTER SYSTEM SET default_statistics_target = 100;
```

### Kubernetes Resource Tuning
```yaml
# HPA with custom metrics
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: zplus-api-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: zplus-api
  minReplicas: 3
  maxReplicas: 20
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: 80
```

## Maintenance Windows

### Scheduled Maintenance
```bash
# Drain nodes for maintenance
kubectl drain <node-name> --ignore-daemonsets --delete-emptydir-data

# Update node group
eksctl upgrade nodegroup --name worker-nodes --cluster zplus-production

# Uncordon nodes
kubectl uncordon <node-name>
```

### Zero-Downtime Deployments
```yaml
# Rolling update strategy
spec:
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
      maxSurge: 1
  
  # Readiness probe
  readinessProbe:
    httpGet:
      path: /health
      port: 8080
    initialDelaySeconds: 30
    periodSeconds: 10
```

---

**Last Updated**: August 7, 2025  
**Next Review**: August 21, 2025
