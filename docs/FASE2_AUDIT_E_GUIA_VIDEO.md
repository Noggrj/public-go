# 🔍 Auditoria Fase 2 — Tech Challenge + Guia para Gravação do Vídeo

---

## 1. Checklist de Requisitos — Status Atual

### 1.1 Evolução da Aplicação

| Requisito | Status | Evidência |
|:---|:---:|:---|
| **Clean Code** (nomes claros, coesão) | ✅ | Nomes descritivos em todos os handlers/services/domain |
| **Clean Architecture** (separação de camadas) | ✅ | `internal/service/{domain, application, delivery/http, infrastructure}` |
| **Testes unitários** | ✅ | `tests/unit/` — 23 arquivos (identity, inventory, platform, service, sharedkernel) |
| **Testes de integração** | ✅ | `tests/integration/` — 6 arquivos (service, identity, inventory) |

### 1.2 APIs Obrigatórias

| API | Status | Endpoint | Handler |
|:---|:---:|:---|:---|
| Abertura de OS | ✅ | `POST /admin/orders` | `OrderHandler.Create` |
| Consulta de status da OS | ✅ | `GET /orders/{id}/track` | `OrderHandler.TrackOrder` (público) |
| Aprovação/Rejeição de orçamento | ✅ | `POST /orders/{id}/budget-response` | `OrderHandler.ApproveBudget` (webhook) |
| Listagem de OS ativa (prioridade) | ✅ | `GET /admin/orders` | `OrderHandler.ListActive` |
| Atualização de status via e-mail | ✅ | `PATCH /admin/orders/{id}/status` + `notification/` | `OrderHandler.UpdateStatus` + `console_email_service.go` |

### 1.3 Status da OS (6 estados)

| Status | Constante no código |
|:---|:---|
| Recebida | `OrderStatusReceived` |
| Diagnóstico | `OrderStatusInDiagnosis` |
| Aguardando Aprovação | `OrderStatusAwaitingApproval` |
| Em Execução | `OrderStatusInExecution` |
| Finalizada | `OrderStatusCompleted` |
| Entregue | `OrderStatusDelivered` |

Ordenação da listagem: **In Execution > Awaiting Approval > In Diagnosis > Received** (exclui Completed e Delivered). ✅

### 1.4 Infraestrutura

| Requisito | Status | Localização |
|:---|:---:|:---|
| **Dockerfile** (multi-stage) | ✅ | `Dockerfile` (builder → development → production) |
| **docker-compose** (dev local) | ✅ | `docker-compose.yml` (app + PostgreSQL + SonarQube) |
| **K8s Deployments** | ✅ | `k8s/deployment.yaml` (2 replicas, rolling update, probes) |
| **K8s Services** | ✅ | `k8s/service.yaml` + `k8s/db-service.yaml` |
| **K8s ConfigMaps** | ✅ | `k8s/configmap.yaml` |
| **K8s Secrets** | ✅ | `k8s/secret.yaml` (DB_PASSWORD, JWT_SECRET) |
| **K8s HPA** | ✅ | `k8s/hpa.yaml` (CPU 70%, Memory 80%, 2-10 pods) |
| **Terraform — VPC** | ✅ | `infra/main.tf` (3 AZs, public/private/database subnets, NAT) |
| **Terraform — EKS** | ✅ | `infra/main.tf` (managed node group, IRSA) |
| **Terraform — ECR** | ✅ | `infra/main.tf` (lifecycle policy, scan on push) |
| **Terraform — RDS** | ✅ | `infra/main.tf` (PostgreSQL 16.1, encrypted, backup) |
| **Terraform — docs** | ✅ | `infra/README.md` |

### 1.5 CI/CD

| Etapa do Pipeline | Status | Arquivo |
|:---|:---:|:---|
| Build da aplicação | ✅ | `.github/workflows/ci.yml` — Job `build-and-test` |
| Execução dos testes | ✅ | Lint (golangci-lint) + test + coverage |
| Build da imagem Docker | ✅ | Job `docker-build-push` (target production) |
| Push para ECR | ✅ | `aws-actions/amazon-ecr-login` + `docker/build-push-action` |
| Deploy no cluster K8s | ✅ | Job `deploy` (kubectl set image + rollout status) |
| Deploy do banco de dados | ✅ | Migrations via pod temporário no EKS |
| Aplicação dos manifestos YAML | ✅ | K8s via `kubectl` no pipeline |
| Release | ✅ | Job `release` (tag + GitHub Release) |

### 1.6 Entregáveis (README.md)

| Item | Status | Observação |
|:---|:---:|:---|
| Descrição da solução | ✅ | Seção principal do README |
| Desenho da arquitetura | ✅ | 2 diagramas Mermaid (infra + módulos) |
| Instruções de execução local | ✅ | Quick Start com `make up` |
| Instruções de deploy K8s | ✅ | Instruções `kubectl apply -f k8s/` |
| Instruções Terraform | ✅ | Seção com `terraform init/plan/apply` |
| Link Postman/Swagger | ✅ | `docs/postman_collection.json` + Swagger UI |
| Link para vídeo | ⚠️ | **Pendente** — precisa gravar e adicionar o link |

---

## 2. Guia Completo para Deploy e Gravação do Vídeo (até 15 min)

### 📋 Pré-requisitos

- Docker Desktop rodando
- AWS CLI configurado (`aws configure`)
- `kubectl` instalado
- `terraform` instalado
- Postman ou equivalente
- Software de gravação de tela (OBS Studio, Camtasia, etc.)

---

### 🎬 Roteiro do Vídeo (Sugestão de Tempo)

#### Parte 1 — Execução Local + APIs (≈ 4 min)

```
1. Mostrar a estrutura do projeto no editor (Clean Architecture)
   - internal/service/{domain, application, delivery, infrastructure}
   - Brevemente explicar as camadas

2. Subir ambiente local
   make up                    # Docker compose (app + DB + SonarQube)
   make migrate-docker        # Rodar migrations
   make seed-docker           # Popular dados de teste

3. Mostrar Swagger UI funcionando
   http://localhost:8080/swagger/index.html

4. Fazer login e obter JWT
   POST /auth/login  →  {"email":"admin@autorepair.com","password":"admin123"}
```

#### Parte 2 — Consumo das APIs (≈ 4 min)

Mostrar o fluxo completo de uma OS no Postman/Swagger:

```
# 1. Criar OS
POST /admin/orders
Body: { "client_id": "<id>", "vehicle_id": "<id>", "items": [...] }
→ Anotar o "id" retornado

# 2. Consultar status (público)
GET /orders/{id}/track
→ Status: "Received"

# 3. Iniciar diagnóstico
POST /admin/orders/{id}/diagnosis:start
→ Status: "In diagnosis"

# 4. Enviar orçamento (notifica cliente por e-mail/console)
POST /admin/orders/{id}/budget:send
→ Status: "Awaiting approval"

# 5. Aprovação externa do orçamento (webhook público)
POST /orders/{id}/budget-response  →  {"approved": true}
→ Status: "In execution"

# 6. Listar ordens ativas (ordenação por prioridade)
GET /admin/orders
→ Mostrar que "In Execution" aparece primeiro, Completed/Delivered excluídas

# 7. Finalizar e entregar
POST /admin/orders/{id}/finish    →  Status: "Completed"
POST /admin/orders/{id}/deliver   →  Status: "Delivered"

# 8. Verificar que a OS sumiu da listagem ativa
GET /admin/orders
```

#### Parte 3 — CI/CD (≈ 2 min)

```
1. Mostrar o arquivo .github/workflows/ci.yml
   - Explicar os 4 jobs: build-and-test → docker-build-push → deploy → release

2. Opções para demonstrar:
   a) Mostrar uma execução anterior no GitHub Actions (aba "Actions" do repo)
   b) OU fazer um push para main e mostrar o pipeline iniciando
   c) OU mostrar o log de uma run bem-sucedida com os 4 jobs verdes

3. Destacar:
   - Lint (golangci-lint)
   - Testes com coverage
   - Build Docker (multi-stage, target production)
   - Push ECR + Deploy EKS
```

#### Parte 4 — Deploy Local em Kubernetes com Kind (≈ 3 min)

> **Passo a passo testado e funcionando** — comandos exatos usados com Docker Desktop + Kind.

**Etapa 1 — Criar cluster Kind no Docker Desktop**

```
Docker Desktop → Settings → Kubernetes → Create Kubernetes Cluster → Kind
  - Nodes: 1
  - Version: 1.31.1
  - Clicar "Create" e aguardar ficar pronto
```

**Etapa 2 — Verificar o cluster**

```bash
kubectl get nodes
kubectl cluster-info
```

**Etapa 3 — Buildar a imagem de produção**

```bash
docker build --target production -t autorepair:latest .
```

**Etapa 4 — Carregar imagens no cluster Kind**

O Kind não compartilha imagens com o Docker host. É preciso copiar manualmente:

```bash
# Salvar imagem da API como .tar
docker save autorepair:latest -o tmp/autorepair.tar

# Copiar para o node worker do Kind e importar
docker cp tmp/autorepair.tar desktop-worker:/autorepair.tar
docker exec desktop-worker ctr -n k8s.io images import /autorepair.tar

# Fazer o mesmo com o PostgreSQL
docker pull postgres:16-alpine
docker save postgres:16-alpine -o tmp/postgres16.tar
docker cp tmp/postgres16.tar desktop-worker:/postgres16.tar
docker exec desktop-worker ctr -n k8s.io images import /postgres16.tar
```

> ⚠️ **Importante**: Os manifestos K8s já estão com `imagePullPolicy: Never` para usar imagens locais.

**Etapa 5 — Aplicar os manifestos Kubernetes**

```bash
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/
```

Resultado esperado:
```
namespace/autorepair created
configmap/autorepair-config created
secret/autorepair-secret created
persistentvolumeclaim/postgres-pvc created
deployment.apps/postgres created
service/postgres-service created
deployment.apps/autorepair-api created
service/autorepair-service created
horizontalpodautoscaler.autoscaling/autorepair-hpa created
```

**Etapa 6 — Rodar as migrations no banco K8s**

```bash
# Port-forward para o PostgreSQL do K8s
kubectl port-forward svc/postgres-service 5434:5432 -n autorepair

# Em outro terminal, rodar migrations via Docker
docker run --rm --network host -v ${PWD}/migrations:/migrations migrate/migrate -path=/migrations/ -database "postgres://postgres:postgres@host.docker.internal:5434/autorepair?sslmode=disable" up
```

**Etapa 7 — Verificar que tudo está funcionando**

```bash
# Checar pods (todos devem estar Running 1/1)
kubectl get pods -n autorepair

# Checar serviços
kubectl get svc -n autorepair

# Checar HPA
kubectl get hpa -n autorepair

# Port-forward da API
kubectl port-forward svc/autorepair-service 9090:8080 -n autorepair

# Testar health (em outro terminal)
curl http://localhost:9090/health
# → {"status":"ok"}
```

**Etapa 8 — Testar login e APIs via Postman**

Use `http://localhost:9090` como base URL no Postman:

```
POST http://localhost:9090/auth/login
Body: {"email":"admin@autorepair.com","password":"admin123"}
```

---

#### Parte 4B — Deploy na AWS com Terraform + EKS (Cloud)

> **Passo a passo para provisionamento real na nuvem AWS.**

**Pré-requisitos AWS:**
- AWS CLI configurado (`aws configure`)
- Conta AWS com permissões para VPC, EKS, RDS, ECR, IAM
- Terraform instalado

**Etapa 1 — Provisionar a infraestrutura com Terraform**

```bash
cd infra

# Copiar e preencher as variáveis
cp terraform.tfvars.example terraform.tfvars
# Editar terraform.tfvars com suas credenciais:
#   - db_password     = "sua-senha-segura"
#   - jwt_secret      = "seu-jwt-secret"
#   - db_username     = "postgres"
#   - db_name         = "autorepair"
#   - cluster_name    = "autorepair-cluster"

# Inicializar, planejar e aplicar
terraform init
terraform plan      # Revisar os recursos que serão criados
terraform apply     # Confirmar com "yes"
```

Recursos provisionados pelo Terraform:

| Recurso | Descrição |
|:---|:---|
| **VPC** | 3 AZs, subnets públicas/privadas/database, NAT Gateway |
| **EKS** | Cluster Kubernetes gerenciado com node group auto-scaling |
| **ECR** | Registro de containers com lifecycle policy |
| **RDS** | PostgreSQL 16, encrypted, backup automático 7 dias |
| **K8s Namespace** | `autorepair` criado automaticamente |
| **K8s Secret/ConfigMap** | Credenciais e config injetados via Terraform |

**Etapa 2 — Configurar kubectl para o EKS**

```bash
aws eks update-kubeconfig --region us-east-1 --name autorepair-cluster
kubectl get nodes  # Verificar que os nodes estão Ready
```

**Etapa 3 — Push da imagem para ECR**

```bash
# Login no ECR
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin <ACCOUNT_ID>.dkr.ecr.us-east-1.amazonaws.com

# Build e tag
docker build --target production -t autorepair:latest .
docker tag autorepair:latest <ACCOUNT_ID>.dkr.ecr.us-east-1.amazonaws.com/autorepair:latest

# Push
docker push <ACCOUNT_ID>.dkr.ecr.us-east-1.amazonaws.com/autorepair:latest
```

> Substitua `<ACCOUNT_ID>` pelo ID da sua conta AWS. Obtenha com: `aws sts get-caller-identity`

**Etapa 4 — Ajustar manifestos K8s para produção**

No `k8s/deployment.yaml`, ajustar a imagem para o ECR:
```yaml
image: <ACCOUNT_ID>.dkr.ecr.us-east-1.amazonaws.com/autorepair:latest
imagePullPolicy: Always
```

> Os ConfigMaps e Secrets já são criados pelo Terraform com os valores do RDS.

**Etapa 5 — Aplicar manifestos e migrations**

```bash
# Aplicar manifestos (namespace já foi criado pelo Terraform)
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
kubectl apply -f k8s/hpa.yaml
kubectl apply -f k8s/db-deployment.yaml  # Opcional: só para dev. Em prod, usar RDS
kubectl apply -f k8s/db-service.yaml

# Rodar migrations via pod temporário
kubectl run migrate-job --rm -i --restart=Never \
  --namespace=autorepair \
  --image=<ACCOUNT_ID>.dkr.ecr.us-east-1.amazonaws.com/autorepair:latest \
  --env="DB_URL=<RDS_CONNECTION_STRING>" \
  -- /bin/sh -c "migrate -path /root/migrations -database \$DB_URL up"
```

**Etapa 6 — Verificar e acessar**

```bash
kubectl get pods -n autorepair
kubectl get svc -n autorepair
kubectl get hpa -n autorepair

# Obter a URL do Load Balancer (se configurado)
kubectl get svc autorepair-service -n autorepair -o jsonpath='{.status.loadBalancer.ingress[0].hostname}'
```

**Etapa 7 — Destruir infraestrutura (quando terminar)**

```bash
cd infra
terraform destroy   # Confirmar com "yes"
```

> ⚠️ **IMPORTANTE**: Destrua a infra AWS após gravar o vídeo para evitar custos.

---

#### Parte 5 — Escalabilidade Automática (≈ 2 min)

```bash
# Mostrar HPA configurado
kubectl get hpa -n autorepair
kubectl describe hpa autorepair-hpa -n autorepair
# → min: 2, max: 10, CPU target: 70%, Memory target: 80%

# Instalar metrics-server (necessário para Kind/local)
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
# Para Kind, pode ser necessário adicionar --kubelet-insecure-tls ao metrics-server

# Simular carga (em outro terminal, com port-forward ativo)
# Opção 1: PowerShell
1..500 | ForEach-Object -Parallel { Invoke-WebRequest -Uri http://localhost:9090/health -UseBasicParsing } -ThrottleLimit 50

# Opção 2: usando hey (instalar: go install github.com/rakyll/hey@latest)
hey -z 60s -c 50 http://localhost:9090/health

# Observar o escalonamento em tempo real
kubectl get hpa -n autorepair -w
# → REPLICAS vai subir de 2 para 3, 4, etc.

kubectl get pods -n autorepair -w
# → Novos pods aparecendo automaticamente
```

---

### 🎯 Checklist Final Antes de Entregar

- [ ] Vídeo gravado (≤ 15 min) demonstrando todos os pontos acima
- [ ] Vídeo publicado no YouTube/Vimeo (público ou não listado)
- [ ] `README.md` atualizado com link do vídeo
- [ ] Repositório Git limpo (sem `terraform.tfvars` com secrets, sem `.env` real)
- [ ] Postman Collection atualizada em `docs/postman_collection.json`
- [ ] Swagger acessível em `/swagger/index.html`
- [ ] Todos os testes passando (`make test-unit` + `make test-integration`)
- [ ] Infraestrutura AWS destruída após gravar o vídeo (`terraform destroy`)

---

### ⚡ Comandos Rápidos de Referência

| Ação | Comando |
|:---|:---|
| **Ambiente Local (Docker Compose)** | |
| Subir ambiente local | `make up` |
| Rodar migrations | `make migrate-docker` |
| Seed de dados | `make seed-docker` |
| Parar tudo | `make down` |
| **Testes** | |
| Testes unitários | `make test-unit` |
| Testes de integração | `make test-integration` |
| Coverage completo | `make test-cover` |
| Lint | `make lint` |
| **Docker** | |
| Build produção | `docker build --target production -t autorepair:latest .` |
| **Kubernetes Local (Kind)** | |
| Aplicar manifestos | `kubectl apply -f k8s/` |
| Ver pods | `kubectl get pods -n autorepair` |
| Ver HPA | `kubectl get hpa -n autorepair` |
| Port-forward API | `kubectl port-forward svc/autorepair-service 9090:8080 -n autorepair` |
| Port-forward DB | `kubectl port-forward svc/postgres-service 5434:5432 -n autorepair` |
| Reiniciar API | `kubectl rollout restart deployment/autorepair-api -n autorepair` |
| Ver logs API | `kubectl logs deployment/autorepair-api -n autorepair --tail=30` |
| **AWS (Terraform)** | |
| Provisionar infra | `cd infra; terraform init; terraform plan; terraform apply` |
| Conectar ao EKS | `aws eks update-kubeconfig --region us-east-1 --name autorepair-cluster` |
| Destruir infra | `cd infra; terraform destroy` |
