# 🎬 Roteiro do Vídeo — Fase 2 Tech Challenge FIAP

## Introdução (~1 min)

> "Bom dia, galera! Tudo bem? Me chamo **Matheus Nogueira** e hoje estou gravando esse vídeo referente à **Fase 2 do desafio Tech Challenge** da pós-graduação de **Arquitetura de Software na FIAP**.
>
> Nessa fase, o objetivo era **containerizar nossa aplicação**, fazer a **orquestração com Kubernetes**, provisionar a **infraestrutura como código** usando Terraform, e montar um **pipeline de CI/CD** completo.
>
> Então hoje vou mostrar pra vocês como tá tudo funcionando na prática — vamos **deployar a aplicação no AWS EKS** e fazer um **teste de carga ao vivo** para mostrar o **autoscaling** do Kubernetes em ação."

---

## Parte 1 — Visão Geral da Arquitetura (~2 min)

> "Antes de ir pro terminal, deixa eu explicar rapidamente a arquitetura:"

### Pontos para mencionar:
- **API:** Go (Golang) com Chi Router, PostgreSQL, JWT Auth
- **Docker:** Multi-stage build, imagem otimizada para produção
- **Kubernetes:** Deployment, Service (LoadBalancer), HPA (Horizontal Pod Autoscaler)
- **Infraestrutura (Terraform):** VPC, EKS Cluster, RDS PostgreSQL, ECR (registry de imagens)
- **CI/CD (GitHub Actions):** Build → Lint → Testes → Push para ECR → Deploy no EKS

> "Toda a infraestrutura é provisionada com **Terraform**, zero configuração manual. O pipeline do GitHub Actions faz o build, roda os testes, faz o push da imagem Docker pro ECR da AWS, e faz o deploy automático no cluster EKS."

---

## Parte 2 — Mostrando a Infra Provisionada (~2 min)

### Comandos para mostrar:

```bash
# Verificar o cluster EKS
kubectl get nodes
```
> "Aqui vemos que o cluster EKS está ativo, com **1 node** rodando Kubernetes **v1.29**."

```bash
# Verificar os pods da aplicação
kubectl get pods -n autorepair
```
> "Nosso deployment criou **2 réplicas** da API, as duas com status **Running**."

```bash
# Verificar o service (LoadBalancer)
kubectl get svc -n autorepair
```
> "Temos um **LoadBalancer** que expõe a API para a internet com uma URL pública da AWS."

```bash
# Verificar o HPA (autoscaling)
kubectl get hpa -n autorepair
```
> "E aqui temos o **HPA (Horizontal Pod Autoscaler)** configurado para escalar entre **2 e 10 réplicas**, baseado no uso de **CPU e memória**."

```bash
# Health check da API pelo LoadBalancer
curl http://<URL_DO_LOADBALANCER>/health
```
> "E fazendo um health check, a API retorna `status: ok` — tudo funcionando!"

---

## Parte 3 — Pipeline CI/CD (~1 min)

> "Agora vou mostrar rapidamente o pipeline no GitHub Actions."

### Mostrar no navegador:
- Abrir o repositório no GitHub → aba **Actions**
- Mostrar o último workflow executado com os 3 jobs:
  1. ✅ **build-and-test** — lint (golangci-lint), testes unitários com cobertura
  2. ✅ **docker-build-push** — build da imagem Docker e push para o ECR
  3. ✅ **deploy** — kubectl apply dos manifestos e deploy no EKS

> "Toda vez que fazemos um push na branch `main`, o pipeline roda automaticamente: faz o lint, os testes, builda a imagem Docker, faz o push pro ECR, e deploya no cluster. **Zero intervenção manual.**"

---

## Parte 4 — Demonstração do Autoscaling (~3 min)

> "Agora a parte mais legal: vamos simular um **teste de carga** e ver o **autoscaling** acontecendo em tempo real. Vou abrir **3 terminais**:"

### Terminal 1 — Monitoramento do HPA
```bash
kubectl get hpa -n autorepair -w
```
> "Nesse primeiro terminal, estou monitorando o **HPA em tempo real**. Ele mostra o uso de CPU e memória atual versus o target, e quantas réplicas estão rodando. Agora temos **2 réplicas** com CPU em **0%**."

### Terminal 2 — Monitoramento dos Pods
```bash
kubectl get pods -n autorepair -w
```
> "No segundo terminal, estou monitorando os **pods**. Quando o autoscaling acontecer, vamos ver os novos pods aparecendo aqui — passando de **Pending** para **ContainerCreating** e depois **Running**."

### Terminal 3 — Gerar Carga
```bash
kubectl run load-generator --image=busybox -n autorepair --restart=Never -- /bin/sh -c "while true; do wget -q -O- http://autorepair-service.autorepair.svc.cluster.local:80/health > /dev/null 2>&1; done"
```
> "E no terceiro terminal, vou **iniciar o teste de carga**. Esse comando cria um pod dentro do cluster que fica fazendo requisições contínuas ao endpoint `/health` da nossa API. A vantagem de rodar de dentro do cluster é que a latência é mínima e a carga é intensa."

### Narrar o que acontece:

> "Reparem no terminal 1: o uso de CPU está subindo... **3%, 9%, 22%, 33%** — já ultrapassou o target de 10%.
>
> E agora olhem o terminal 2: o HPA detectou que precisa de mais réplicas e está criando novos pods. 
> Vejam: `Pending → ContainerCreating → Running`. Saímos de **2 pods para 4 pods** automaticamente!
>
> Isso é o **Horizontal Pod Autoscaler** do Kubernetes em ação — ele monitora as métricas e **escala horizontalmente** quando a demanda aumenta."

### Parar a carga e mostrar o scale-down:
```bash
kubectl delete pod load-generator -n autorepair
```
> "Agora vou parar o gerador de carga. E em alguns segundos, reparem que o uso de CPU vai cair, e o HPA vai **remover os pods extras**, voltando para as **2 réplicas originais**.
>
> E pronto! O scale-down aconteceu automaticamente. Isso mostra que a aplicação é **resiliente e elástica** — escala quando precisa e economiza recursos quando a demanda cai."

---

## Encerramento (~1 min)

> "Então resumindo o que entregamos na Fase 2:
>
> 1. **Docker** — imagem otimizada com multi-stage build
> 2. **Kubernetes** — deployment, service com LoadBalancer, HPA para autoscaling
> 3. **Terraform** — toda a infra na AWS provisionada como código: VPC, EKS, RDS, ECR
> 4. **CI/CD** — pipeline completo: lint, testes, build, push e deploy automatizado
> 5. **Autoscaling** — demonstrado ao vivo com teste de carga
>
> Todos os arquivos estão no repositório do GitHub, incluindo a documentação detalhada de como replicar esse ambiente.
>
> Muito obrigado por assistir! Qualquer dúvida pode mandar nos comentários. Valeu! 👋"

---

## ⏱️ Tempo estimado total: ~10 minutos

## 📋 Checklist antes de gravar:

- [ ] Cluster EKS rodando (`kubectl get nodes` → Ready)
- [ ] 2 pods Running (`kubectl get pods -n autorepair`)
- [ ] HPA ativo (`kubectl get hpa -n autorepair`)
- [ ] LoadBalancer com URL (`kubectl get svc -n autorepair`)
- [ ] Health check OK (`curl http://<LB_URL>/health`)
- [ ] Pipeline no GitHub com os 3 jobs verdes
- [ ] 3 terminais preparados
- [ ] Sessão do AWS Academy ativa (verifica tempo restante!)
