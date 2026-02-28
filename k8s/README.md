# K8s Manifests — Ambientes

Este diretório contém manifestos separados por ambiente:

```
k8s/
├── base/              # Manifestos compartilhados (iguais nos dois ambientes)
│   ├── namespace.yaml
│   └── hpa.yaml
│
├── local/             # Kind (Docker Desktop) — deploy local
│   ├── deployment.yaml      # image: autorepair:latest, imagePullPolicy: Never
│   ├── service.yaml         # ClusterIP (acesso via port-forward)
│   ├── configmap.yaml       # DB_HOST = postgres-service (pod local)
│   ├── secret.yaml          # Credenciais base64 fixas (dev)
│   ├── db-deployment.yaml   # PostgreSQL como pod no cluster
│   └── db-service.yaml      # Service do pod PostgreSQL
│
└── production/        # AWS EKS — deploy na nuvem
    ├── deployment.yaml      # image: ECR URL, imagePullPolicy: Always
    └── service.yaml         # LoadBalancer (URL pública via ALB)
    # ⚠️ configmap e secret são criados pelo Terraform automaticamente
    # ⚠️ banco usa RDS (não precisa db-deployment/db-service)
```

## Como usar

### Local (Kind)
```bash
kubectl apply -f k8s/base/
kubectl apply -f k8s/local/
```

### Produção (AWS EKS)
```bash
# Terraform já cria: namespace, configmap, secret
cd infra && terraform apply

# Depois aplicar manifestos de produção
kubectl apply -f k8s/base/
kubectl apply -f k8s/production/
```
