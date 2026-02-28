# Guia de Contribuição

Obrigado por contribuir com o projeto AutoRepair! Siga as diretrizes abaixo para garantir a qualidade e consistência do código.

## 🧪 Testes e Cobertura de Código

Nós valorizamos testes robustos. O projeto utiliza o framework nativo `testing` do Go, complementado pela biblioteca [Testify](https://github.com/stretchr/testify) para assertions e mocks.

### Pré-requisitos
- Go 1.23+
- Docker (opcional, para testes de integração com banco de dados real)

### Executando Testes

Para rodar todos os testes (unitários e de integração) e gerar o relatório de cobertura, utilize o script facilitador na raiz do projeto:

**Windows (PowerShell):**
```powershell
.\test.ps1
```

Este script irá:
1. Executar todos os testes (`./tests/...`).
2. Gerar um arquivo de perfil de cobertura (`coverage.out`).
3. Filtrar arquivos não relacionados (cmd, docs, seeds).
4. Exibir um resumo no terminal.
5. Gerar um relatório HTML detalhado (`coverage.html`) e abri-lo automaticamente (se possível).

**Linux/Mac (Make):**
```bash
make cover
# ou para ver o HTML:
make cover-html
```

### Requisitos de Cobertura
- **Mínimo Global:** Buscamos manter a cobertura acima de **80%**.
- **Código Crítico:** Serviços e Handlers principais devem ter cobertura próxima de 100% para caminhos felizes e de erro.
- **Novas Funcionalidades:** Todo novo código deve vir acompanhado de testes unitários.

### Estrutura de Testes
- **Unitários:** Localizados em `tests/unit/`. Devem testar a lógica de negócio isoladamente, usando mocks para dependências.
- **Integração:** Localizados em `tests/integration/`. Testam a interação com banco de dados e outros componentes reais. Use a tag `-tags=integration` (o script `test.ps1` já faz isso).

## 📊 Qualidade de Código (SonarQube)

O projeto está integrado com SonarQube para análise estática.
Para rodar a análise localmente (necessário Docker):

```powershell
# Inicie o servidor SonarQube
docker compose up -d sonarqube

# Execute o scanner (certifique-se de ter gerado o coverage.out antes)
docker run --rm `
    -e SONAR_HOST_URL="http://sonarqube:9000" `
    -e SONAR_SCANNER_OPTS="-Dsonar.projectKey=go-garage-dev" `
    -e SONAR_TOKEN="seu_token_aqui" `
    -v "${PWD}:/usr/src" `
    --network go_autorepair-net `
    sonarsource/sonar-scanner-cli
```

## 🚀 CI/CD

O projeto utiliza GitHub Actions para Integração Contínua. O workflow está definido em `.github/workflows/ci.yml` e executa:
1. Linting (`golangci-lint`)
2. Testes Unitários e de Integração
3. Verificação de Cobertura
4. Build da aplicação

## 📝 Padrões de Código

- Use `go fmt` antes de commitar.
- Siga as convenções de nomes do Go (CamelCase).
- Documente funções exportadas.

---
Dúvidas? Abra uma issue ou contate a equipe de desenvolvimento.
