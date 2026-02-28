🚀 **Mais uma etapa concluída com sucesso! Fase 2 do Tech Challenge da Pós-Graduação em Arquitetura de Software na FIAP entregue!** 💻☁️

Nas últimas semanas, mergulhei de cabeça no desafio de evoluir a arquitetura do nosso sistema de AutoRepair (Oficina Mecânica), saindo do desenvolvimento puro para um ambiente real de orquestração e nuvem! 

A ideia aqui não era apenas "fazer o código rodar", mas sim criar uma fundação sólida, escalável e automatizada. Queria compartilhar com a comunidade um pouco dessa stack e do que foi construído:

🛠️ **A Stack e Arquitetura:**
- **Linguagem:** Nosso backend foi construído em **Go (Golang)**, utilizando os princípios da Clean Architecture. Alta performance, concorrência nativa e baixo consumo de recursos!
- **Banco de Dados:** **PostgreSQL** para garantir robustez e relacionamento seguro dos dados.

☁️ **Infraestrutura e Orquestração (O grande salto):**
- **Docker:** Criei os containers focando no menor tamanho possível através de Multi-Stage Builds (nossa API compilada rodando nas imagens Alpine!).
- **Kubernetes (K8s):** Todo o tráfego e ciclo de vida da aplicação agora é orquestrado pelo K8s. Implementamos Deployments, Services (LoadBalancers), ConfigMaps, Secrets e, o mais legal: **HPA (Horizontal Pod Autoscaler)**, garantindo que o sistema escale automaticamente (criando novos Pods) quando a CPU ou memória atingem limites predefinidos sob carga!

🏗️ **Testes Locais vs. Nuvem Real:**
Para validar tudo, criei duas esteiras de provisionamento:
1. **Ambiente Local:** Utilize o Kubernetes integrado do **Docker Desktop** (ideal para debugar, simular o HPA com testes de carga e validar manifestos rapidamente sem custos).
2. **Ambiente de Produção na AWS:** Utilize EKS (Elastic Kubernetes Service) e RDS na nuvem da Amazon. 

⚙️ **Infraestrutura como Código (IaC) e CI/CD:**
Para não fazer nada clicando em tela (ClickOps), usamos o **Terraform** para provisionar 100% da infraestrutura na AWS (VPC, Cluster EKS, banco RDS e repositório ECR).
E para fechar com chave de ouro: um pipeline completo no **GitHub Actions** que faz o lint, roda os testes de cobertura, builda a imagem Docker, envia pro repositório ECR e faz o deploy transparente lá no cluster EKS! 🔄

A jornada de pegar um monolito em Go e transformá-lo numa solução containerizada, altamente disponível e com CI/CD na nuvem traz um aprendizado absurdo sobre DevSecOps e Arquitetura Cloud-Native. Pude validar métricas e ver o Autoscaling brilhando sob estresse ao vivo! 🔥

Agradeço imensamente aos professores e mentores da FIAP por essa base técnica intensa. Que venha a Fase 3! 💪

Se alguém estiver estudando Golang, Kubernetes ou Terraform e quiser trocar uma ideia sobre os desafios enfrentados, bora bater um papo nos comentários! 👇

#SoftwareArchitecture #GoLang #Kubernetes #AWS #Terraform #DevOps #CICD #TechChallenge #FIAP #CloudNative #Backend
