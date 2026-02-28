# ☁️ Deploy Kubernetes Produção (AWS EKS)

Guia rápido para provisionar a infraestrutura e rodar a aplicação AutoRepair na AWS.

## Pré-requisitos
- Conta AWS (ou AWS Academy) com credenciais configuradas.
- `terraform`, `aws-cli` e `kubectl` instalados localmente.
- GitHub Actions configurado com os secrets necessários (para deploy automatizado).

## Passo a Passo

### 1. Provisionar a Infraestrutura (Terraform)
Isso criará a VPC, o Cluster EKS, o Banco de Dados RDS e o repositório ECR.
```bash
cd infra/
terraform init
terraform apply -auto-approve
```

### 2. Conectar o `kubectl` ao Cluster EKS
Após o Terraform terminar, atualize seu kubeconfig para acessar o cluster:
```bash
aws eks update-kubeconfig --region us-east-1 --name autorepair-cluster
```

### 3. Deploy da Aplicação (GitHub Actions)
O deploy em produção é gerenciado pelo pipeline de **CI/CD** (`.github/workflows/ci.yml`).
1. Faça um `git push` ou de um **Merge** na branch `main`.
2. O GitHub Actions fará tudo automaticamente:
   - Build da imagem Docker
   - Push da imagem para o AWS ECR criado no passo 1.
   - Aplica os manifestos K8s (`k8s/base/` e `k8s/production/`).
   - Roda as migrations do banco de dados no RDS.

> **💡 Deploy Manual (Apenas se necessário):**
> Caso deseje aplicar os manifestos manualmente pelo terminal:
> ```bash
> kubectl apply -f k8s/base/
> kubectl apply -f k8s/production/
> # Atualiza a imagem do deployment (se já fez o push pro ECR manualmente)
> kubectl set image deployment/autorepair-api autorepair-api=<SEU_CONTA_ID>.dkr.ecr.us-east-1.amazonaws.com/autorepair:<TAG> -n autorepair
> ```

### 4. Verificar o Ambiente
1. Veja os pods escalados:
   ```bash
   kubectl get pods -n autorepair -w
   ```
2. Veja o HPA e suas métricas:
   ```bash
   kubectl get hpa -n autorepair -w
   ```
3. Pegue a URL pública do **LoadBalancer** para acessar a API:
   ```bash
   kubectl get svc -n autorepair
   # Copie a URL gerada e acesse: http://<URL_DO_LOADBALANCER>/health
   ```

### 5. Destruindo o Ambiente
Para evitar cobranças indesejadas na AWS, lembre-se de destruir a infraestrutura após o uso:
```bash
cd infra/
terraform destroy -auto-approve
```
