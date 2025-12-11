# Massage-Bot CI/CD Pipeline

## Overview
This pipeline automates testing, building, and deployment of the Massage-Bot application.

## Pipeline Stages
1. **Test**: Runs Go unit tests and vetting
2. **Build**: Builds Docker image and pushes to GitLab Registry
3. **Deploy**: Deploys to Kubernetes (manual trigger)

## Requirements
- GitLab Container Registry access
- Kubernetes cluster (minikube for local development)
- Proper kubeconfig in CI/CD variables

## Local Development
\`\`\`bash
# Test locally
go test ./...

# Build and run locally
docker build -t massage-bot:local .
docker run -p 8080:8080 massage-bot:local

# Deploy to local minikube
kubectl apply -f k8s/
\`\`\`

## Deployment Notes
- Deployment stage is manual (requires local runner for minikube)
- Image: \`registry.gitlab.com/kfilin/massage-bot:latest\`
- Service: ClusterIP on port 8080
- Health endpoint: \`/health\`
