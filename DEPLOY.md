# 🚀 Deploy — Comandos Rápidos

---

## Deploy Local (Kind — Docker Desktop)

```bash
# 1. Criar cluster Kind
# Docker Desktop → Settings → Kubernetes → Create Cluster → Kind (1 node, v1.31.1)

# 2. Verificar cluster
kubectl get nodes

# 3. Buildar imagem de produção
docker build --target production -t autorepair:latest .

# 4. Carregar imagens no Kind
docker save autorepair:latest -o tmp/autorepair.tar
docker cp tmp/autorepair.tar desktop-worker:/autorepair.tar
docker exec desktop-worker ctr -n k8s.io images import /autorepair.tar

docker pull postgres:16-alpine
docker save postgres:16-alpine -o tmp/postgres16.tar
docker cp tmp/postgres16.tar desktop-worker:/postgres16.tar
docker exec desktop-worker ctr -n k8s.io images import /postgres16.tar

# 5. Aplicar manifestos
kubectl apply -f k8s/base/
kubectl apply -f k8s/local/

# 6. Rodar migrations
kubectl port-forward svc/postgres-service 5434:5432 -n autorepair
# (em outro terminal)
docker run --rm --network host -v ${PWD}/migrations:/migrations migrate/migrate -path=/migrations/ -database "postgres://postgres:postgres@host.docker.internal:5434/autorepair?sslmode=disable" up

# 7. Verificar tudo
kubectl get pods -n autorepair
kubectl get svc -n autorepair
kubectl get hpa -n autorepair

# 8. Acessar a API
kubectl port-forward svc/autorepair-service 9090:8080 -n autorepair
# API disponível em http://localhost:9090
```

---

## Deploy AWS (EKS + Terraform)

```bash
# 1. Provisionar infraestrutura
cd infra
cp terraform.tfvars.example terraform.tfvars
# Editar terraform.tfvars com credenciais reais
terraform init
terraform plan
terraform apply

# 2. Conectar ao cluster EKS
aws eks update-kubeconfig --region us-east-1 --name autorepair-cluster
kubectl get nodes

# 3. Push da imagem para ECR
aws sts get-caller-identity --query Account --output text   # obter Account ID
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin <ACCOUNT_ID>.dkr.ecr.us-east-1.amazonaws.com
docker build --target production -t autorepair:latest .
docker tag autorepair:latest <ACCOUNT_ID>.dkr.ecr.us-east-1.amazonaws.com/autorepair:latest
docker push <ACCOUNT_ID>.dkr.ecr.us-east-1.amazonaws.com/autorepair:latest

# 4. Aplicar manifestos de produção
kubectl apply -f k8s/base/
kubectl apply -f k8s/production/

# 5. Rodar migrations
kubectl run migrate-job --rm -i --restart=Never \
  --namespace=autorepair \
  --image=<ACCOUNT_ID>.dkr.ecr.us-east-1.amazonaws.com/autorepair:latest \
  --env="DB_URL=postgres://postgres:<DB_PASS>@<RDS_ENDPOINT>:5432/autorepair?sslmode=require" \
  -- /bin/sh -c "migrate -path /root/migrations -database \$DB_URL up"

# 6. Verificar
kubectl get pods -n autorepair
kubectl get svc -n autorepair
kubectl get hpa -n autorepair

# 7. Obter URL pública
kubectl get svc autorepair-service -n autorepair -o jsonpath='{.status.loadBalancer.ingress[0].hostname}'

# 8. Destruir infra após gravar vídeo
cd infra
terraform destroy
```

---

## Testar Autoscaling

```bash
# Observar em 3 terminais simultâneos:

# Terminal 1 — HPA
kubectl get hpa -n autorepair -w

# Terminal 2 — Pods
kubectl get pods -n autorepair -w

# Terminal 3 — CPU/RAM
kubectl top pods -n autorepair

# Gerar carga (PowerShell)
while ($true) { 1..50 | ForEach-Object -Parallel { Invoke-WebRequest -Uri http://localhost:9090/health -UseBasicParsing | Out-Null } -ThrottleLimit 50 }

# Gerar carga (hey)
hey -z 60s -c 100 http://localhost:9090/health
```
