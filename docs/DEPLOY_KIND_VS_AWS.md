# 🚀 Deploy Local (Kind) vs Deploy AWS (EKS) — Guia Completo

---

## Visão Geral — O que muda e o que não muda

```
                    ┌──────────────────────────────────┐
                    │   O QUE É IGUAL NOS DOIS         │
                    │                                  │
                    │  • Manifestos K8s (/k8s/)        │
                    │  • Dockerfile (multi-stage)      │
                    │  • API Go (mesma imagem)         │
                    │  • Migrations (mesmos scripts)   │
                    │  • kubectl apply / get / logs    │
                    │  • HPA (mesma config)            │
                    └──────────────────────────────────┘

     KIND (Local)                           AWS (EKS)
┌────────────────────┐            ┌──────────────────────┐
│ Cluster: Docker    │            │ Cluster: AWS EKS     │
│ Desktop            │            │ (gerenciado)         │
│                    │            │                      │
│ Imagens: .tar →    │            │ Imagens: ECR         │
│ docker cp + ctr    │            │ (registry na nuvem)  │
│                    │            │                      │
│ DB: PostgreSQL     │            │ DB: RDS PostgreSQL   │
│ como pod           │            │ (gerenciado, backup)  │
│                    │            │                      │
│ Acesso: port-      │            │ Acesso: Load         │
│ forward            │            │ Balancer (ALB)       │
│ localhost:9090     │            │ URL pública          │
│                    │            │                      │
│ Secrets: YAML      │            │ Secrets: Terraform   │
│ com base64         │            │ → K8s Secrets        │
│                    │            │                      │
│ Custo: R$ 0        │            │ Custo: ~$0.30/hora   │
│                    │            │ (EKS + EC2 + RDS)    │
│                    │            │                      │
│ Infra: 1 clique    │            │ Infra: terraform     │
│ Docker Desktop     │            │ apply (~15 min)      │
└────────────────────┘            └──────────────────────┘
```

---

## 1. Comparativo Detalhado

| Aspecto | Kind (Local) | AWS (EKS) |
|:---|:---|:---|
| **Onde roda** | Seu PC (Docker Desktop) | Nuvem AWS |
| **Criar cluster** | Docker Desktop → Kind | `terraform apply` (15 min) |
| **Imagem Docker** | `docker cp` + `ctr import` (`.tar`) | `docker push` para ECR |
| **Banco de dados** | Pod PostgreSQL no cluster | RDS PostgreSQL gerenciado |
| **Credenciais** | `k8s/secret.yaml` (base64 manual) | Terraform injeta automaticamente |
| **Acessar API** | `kubectl port-forward` (localhost) | Load Balancer com URL pública |
| **Metrics** | Instalar metrics-server manualmente | Já vem com CloudWatch |
| **Custo** | Grátis | ~$0.30/hora (~$7/dia)  |
| **Quando usar** | Dev, testes, gravação de vídeo | Produção real |

---

## 2. Deploy na AWS — Passo a Passo Completo

### 2.1 Pré-requisitos

```bash
# Verificar ferramentas instaladas
aws --version         # AWS CLI
terraform --version   # Terraform
kubectl version       # kubectl
docker --version      # Docker

# Configurar credenciais AWS
aws configure
# → AWS Access Key ID: sua-chave
# → AWS Secret Access Key: sua-secret
# → Default region: us-east-1
# → Default output: json
```

### 2.2 Provisionar infraestrutura com Terraform

```bash
cd infra

# 1. Preencher variáveis
cp terraform.tfvars.example terraform.tfvars
```

Editar `terraform.tfvars`:
```hcl
# Cluster
cluster_name    = "autorepair-cluster"
cluster_version = "1.28"
environment     = "production"

# Rede
vpc_cidr = "10.0.0.0/16"

# Nodes (máquinas do cluster)
node_instance_types = ["t3.medium"]
node_min_size       = 1
node_max_size       = 3
node_desired_size   = 2

# Banco de dados (RDS)
db_instance_class = "db.t3.micro"
db_name           = "autorepair"
db_username       = "postgres"
db_password       = "SuaSenhaSegura123!"   # ⚠️ NUNCA commitar

# JWT
jwt_secret = "meu-jwt-secret-super-seguro"  # ⚠️ NUNCA commitar
```

> ⚠️ O `terraform.tfvars` já está no `.gitignore`. Ele NUNCA deve ir pro repositório.

```bash
# 2. Provisionar (demora ~15 minutos)
terraform init
terraform plan     # Revise tudo antes
terraform apply    # Digite "yes"
```

**O que o Terraform cria na AWS:**

| Recurso | O que faz | Como as credenciais chegam lá |
|:---|:---|:---|
| **VPC** | Rede isolada (3 AZs, subnets) | — |
| **EKS** | Cluster Kubernetes gerenciado | — |
| **ECR** | Registry de imagens Docker | — |
| **RDS** | PostgreSQL 16 (encrypted, backup 7 dias) | `db_password` do `terraform.tfvars` |
| **K8s Namespace** | `autorepair` | — |
| **K8s Secret** | `autorepair-secret` | `db_password` + `jwt_secret` → injetados pelo Terraform |
| **K8s ConfigMap** | `autorepair-config` | `DB_HOST` = endpoint do RDS (auto) |

> 🔑 **Credenciais**: o Terraform pega `db_password` e `jwt_secret` do `terraform.tfvars` e injeta diretamente como K8s Secret. O `DB_HOST` do ConfigMap aponta automaticamente para o endpoint do RDS criado. Você NÃO precisa configurar nada manualmente.

### 2.3 Conectar ao cluster EKS

```bash
# Configurar kubectl para apontar ao EKS
aws eks update-kubeconfig --region us-east-1 --name autorepair-cluster

# Verificar conexão
kubectl get nodes
# NAME                                          STATUS   ROLES    AGE
# ip-10-0-10-xxx.ec2.internal                   Ready    <none>   5m
# ip-10-0-11-xxx.ec2.internal                   Ready    <none>   5m
```

### 2.4 Push da imagem para ECR

```bash
# Obter ID da conta AWS
aws sts get-caller-identity --query Account --output text
# → 123456789012 (seu Account ID)

# Login no ECR
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 123456789012.dkr.ecr.us-east-1.amazonaws.com

# Build da imagem de produção
docker build --target production -t autorepair:latest .

# Tag para ECR
docker tag autorepair:latest 123456789012.dkr.ecr.us-east-1.amazonaws.com/autorepair:latest

# Push
docker push 123456789012.dkr.ecr.us-east-1.amazonaws.com/autorepair:latest
```

### 2.5 Deploy no EKS

```bash
# Ajustar deployment.yaml para usar imagem do ECR:
#   image: 123456789012.dkr.ecr.us-east-1.amazonaws.com/autorepair:latest
#   imagePullPolicy: Always

# Aplicar manifestos (namespace/secret/configmap já foram criados pelo Terraform)
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
kubectl apply -f k8s/hpa.yaml

# Verificar pods subindo
kubectl get pods -n autorepair -w
# → autorepair-api-xxx   1/1   Running   0   30s
# → autorepair-api-yyy   1/1   Running   0   30s
```

### 2.6 Rodar migrations no RDS

```bash
# O Terraform já criou o RDS. Pegar a connection string:
cd infra
terraform output db_endpoint
# → autorepair-db.xxxxxx.us-east-1.rds.amazonaws.com

# Rodar migrations via pod temporário
kubectl run migrate-job --rm -i --restart=Never \
  --namespace=autorepair \
  --image=123456789012.dkr.ecr.us-east-1.amazonaws.com/autorepair:latest \
  --env="DB_URL=postgres://postgres:SuaSenhaSegura123!@autorepair-db.xxxxxx.us-east-1.rds.amazonaws.com:5432/autorepair?sslmode=require" \
  -- /bin/sh -c "migrate -path /root/migrations -database \$DB_URL up"
```

---

## 3. Testando Autoscaling + Monitoramento (para gravar o vídeo)

### 3.1 Verificar HPA inicial

```bash
kubectl get hpa -n autorepair
# NAME             REFERENCE                   TARGETS           MINPODS  MAXPODS  REPLICAS
# autorepair-hpa   Deployment/autorepair-api   5%/70%, 10%/80%   2        10       2
```

### 3.2 Gerar carga para provocar autoscaling

**Opção 1 — PowerShell (simples)**
```powershell
# Abre 500 requisições paralelas
1..500 | ForEach-Object -Parallel {
    Invoke-WebRequest -Uri http://localhost:9090/health -UseBasicParsing
} -ThrottleLimit 100
```

**Opção 2 — hey (recomendado para o vídeo)**
```bash
# Instalar
go install github.com/rakyll/hey@latest

# 60 segundos de carga, 100 conexões simultâneas
hey -z 60s -c 100 http://localhost:9090/health
```

**Opção 3 — Loop PowerShell contínuo**
```powershell
# Loop infinito com 50 threads (Ctrl+C para parar)
while ($true) {
    1..50 | ForEach-Object -Parallel {
        Invoke-WebRequest -Uri http://localhost:9090/health -UseBasicParsing | Out-Null
    } -ThrottleLimit 50
}
```

### 3.3 Onde observar o autoscaling?

Abra **3 terminais** durante a gravação do vídeo:

**Terminal 1 — HPA em tempo real (escala subindo)**
```bash
kubectl get hpa -n autorepair -w
# → REPLICAS vai mudar: 2 → 3 → 4 → 5 ...
```

**Terminal 2 — Pods aparecendo em tempo real**
```bash
kubectl get pods -n autorepair -w
# → Novos pods vão aparecer com status Pending → Running
```

**Terminal 3 — Consumo de CPU e RAM por pod**
```bash
kubectl top pods -n autorepair
# NAME                              CPU(cores)   MEMORY(bytes)
# autorepair-api-xxx                250m         45Mi
# autorepair-api-yyy                230m         42Mi
# autorepair-api-zzz                180m         38Mi   ← novo pod!
```

> ⚠️ `kubectl top` requer o **metrics-server**. No Kind local, instale com:
> ```bash
> kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
> ```

### 3.4 Na AWS — Monitoramento adicional via CloudWatch

Além do terminal, na AWS você pode mostrar no vídeo:

1. **AWS Console → EKS → Clusters → autorepair-cluster**
   - Mostra nodes, pods, eventos do cluster

2. **AWS Console → CloudWatch → Metrics → ContainerInsights**
   - Gráficos de CPU, RAM, número de pods
   - Ideal para screenshot/gravação do vídeo

3. **AWS Console → EC2 → Instances**
   - Mostra as máquinas (nodes) reais rodando

---

## 4. Resumo Visual — Fluxo de Credenciais

```
                    terraform.tfvars
                    ┌──────────────────┐
                    │ db_password      │
                    │ jwt_secret       │
                    │ db_username      │
                    │ db_name          │
                    └────────┬─────────┘
                             │
                    terraform apply
                             │
              ┌──────────────┼──────────────┐
              ▼              ▼              ▼
        ┌──────────┐  ┌──────────┐  ┌──────────┐
        │ RDS      │  │ K8s      │  │ K8s      │
        │ Postgres │  │ Secret   │  │ ConfigMap│
        │          │  │          │  │          │
        │ user:    │  │ DB_PASS  │  │ DB_HOST= │
        │ postgres │  │ JWT_SEC  │  │ (RDS     │
        │ pass:    │  │          │  │ endpoint)│
        │ ******** │  │          │  │          │
        └──────────┘  └──────────┘  └──────────┘
                             │              │
                             ▼              ▼
                      ┌────────────────────────┐
                      │   API Pod (container)   │
                      │                        │
                      │ DB_URL = postgres://    │
                      │   $(DB_USER):$(DB_PASS) │
                      │   @$(DB_HOST):5432/     │
                      │   $(DB_NAME)?sslmode=   │
                      │   $(DB_SSLMODE)         │
                      └────────────────────────┘
```

No **Kind local**, as credenciais vêm direto do `k8s/secret.yaml` (valores base64 fixos).
Na **AWS**, o Terraform injeta os valores reais automaticamente.

---

## 5. Destruir infraestrutura AWS (após o vídeo)

```bash
cd infra
terraform destroy   # Confirmar com "yes"
```

**Isso remove:** VPC, EKS, ECR, RDS, tudo. Sem custo residual.

> ⚠️ **Faça isso logo após gravar o vídeo.** A infra custa ~$0.30/hora rodando.
