# 🚀 Deploy Kubernetes Local (Docker Desktop)

Guia rápido para rodar a aplicação AutoRepair localmente sem precisar da AWS.

## Pré-requisitos
- Docker Desktop instalado com o **Kubernetes ativado** nas configurações.
- `kubectl` instalado.

## Passo a Passo

1. **Build da Imagem Local**
   ```bash
   docker build -t autorepair:local .
   ```

2. **Aplicar os Manifestos K8s**
   ```bash
   # Cria namespace
   kubectl apply -f k8s/base/namespace.yaml

   # Inicia PostgreSQL e API
   kubectl apply -f k8s/local/

   # Aplica o HPA (Autoscaling)
   kubectl apply -f k8s/base/hpa.yaml
   ```

3. **Verificar os Pods**
   Aguarde os pods ficarem com o status `Running`:
   ```bash
   kubectl get pods -n autorepair -w
   ```

4. **Rodar Migrations do Banco**
   Quando o pod do `postgres` estiver `Running`, rode a migração:
   ```bash
   kubectl run migrate-job --rm -it --restart=Never -n autorepair --image=autorepair:local --env="DB_URL=postgres://postgres:postgres@postgres-service:5432/autorepair?sslmode=disable" -- /bin/sh -c "migrate -path /app/migrations -database \$DB_URL up"
   ```

5. **Acessar a Aplicação**
   Faça o redirecionamento da porta para sua máquina:
   ```bash
   kubectl port-forward svc/autorepair-service -n autorepair 8080:80
   ```
   Acesse: `http://localhost:8080/health` ou `http://localhost:8080/swagger/index.html`

## Teste de Autoscaling Local

Se quiser simular o HPA funcionando no seu computador:

1. Acompanhe os pods e o HPA em abas separadas:
   ```bash
   kubectl get pods -n autorepair -w
   kubectl get hpa -n autorepair -w
   ```
2. Crie pods de carga (rode em outra aba):
   ```bash
   kubectl run load-gen-1 --image=busybox -n autorepair --restart=Never -- /bin/sh -c "while true; do wget -q -O- http://autorepair-service.autorepair.svc.cluster.local:8080/health > /dev/null 2>&1; done"
   ```
3. Para parar o teste e limpar a carga:
   ```bash
   kubectl delete pod load-gen-1 -n autorepair
   ```

## Limpando o Ambiente Local
Para remover todos os recursos criados:
```bash
kubectl delete -f k8s/local/
kubectl delete namespace autorepair
```
