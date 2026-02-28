$ErrorActionPreference = "Stop"

Write-Host "========================================================" -ForegroundColor Cyan
Write-Host "  RODANDO TESTES COM COVERAGE (Unitários + Integração)  " -ForegroundColor Cyan
Write-Host "========================================================" -ForegroundColor Cyan

# 1. Rodar testes
Write-Host "`n[1/4] Executando 'go test'..." -ForegroundColor Yellow
# Remove arquivo antigo para garantir que não estamos lendo lixo
if (Test-Path coverage.out) { Remove-Item coverage.out }

go test -v -tags=integration -coverpkg=./internal/... -coverprofile=coverage.out ./tests/...

if (-not (Test-Path coverage.out)) {
    Write-Error "coverage.out não foi gerado!"
    exit 1
}

# 2. Filtrar
Write-Host "`n[2/4] Filtrando arquivos (cmd, docs, seeds)..." -ForegroundColor Yellow

$excludePatterns = @(
    "github.com/noggrj/autorepair/cmd/",
    "github.com/noggrj/autorepair/docs/",
    "github.com/noggrj/autorepair/seeds/"
)

# Ler todas as linhas
$lines = Get-Content coverage.out

# A primeira linha deve ser o mode
$header = $lines[0]
if ($header -notmatch "^mode:") {
    Write-Warning "Arquivo coverage.out inválido ou sem header 'mode:'. Usando 'mode: set' como padrão."
    $header = "mode: set"
    # Se a primeira linha não for mode, processamos ela como dados também
    $dataLines = $lines
} else {
    $dataLines = $lines | Select-Object -Skip 1
}

# Filtrar linhas de dados
$filteredData = $dataLines | Where-Object {
    $line = $_
    $keep = $true
    foreach ($pattern in $excludePatterns) {
        if ($line -match $pattern) {
            $keep = $false
            break
        }
    }
    $keep
}

# Combinar header e dados filtrados
$finalContent = @($header) + $filteredData

# Escrever arquivo filtrado (ASCII para garantir compatibilidade)
$finalContent | Out-File -FilePath coverage_filtered.out -Encoding ASCII -Force

# 3. Report Texto
Write-Host "`n[3/4] Resumo do Coverage:" -ForegroundColor Yellow
go tool cover -func=coverage_filtered.out

# 4. Report HTML
Write-Host "`n[4/4] Gerando relatório HTML..." -ForegroundColor Yellow
go tool cover -html=coverage_filtered.out -o coverage.html

Write-Host "`nSUCESSO!" -ForegroundColor Green
Write-Host "Relatório salvo em: $(Get-Location)\coverage.html"
Write-Host "Para abrir o relatório: Invoke-Item coverage.html"
