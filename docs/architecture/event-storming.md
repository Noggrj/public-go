# Event Storming

This document represents the Event Storming for the Green Crop system, focusing on the core Service Order (Ordem de Serviço) flow.

## Legenda
- **Command (Blue)**: User actions / API requests.
- **Event (Orange)**: What happened in the system.
- **Aggregate (Yellow)**: Data clusters where commands are executed.
- **Policy (Purple)**: Rules triggered by events (often async or immediate side-effects).
- **Read Model (Green)**: Data views for the user.

```mermaid
eventStorming
    %% Order Creation Flow
    Client "Cliente" -> Command "Solicitar Serviço"
    Command "Solicitar Serviço" -> Aggregate "Ordem de Serviço"
    Aggregate "Ordem de Serviço" -> Event "Ordem de Serviço Recebida"
    Event "Ordem de Serviço Recebida" -> Policy "Notificar Gerente"
    
    %% Diagnosis Flow
    Mechanic "Mecânico" -> Command "Iniciar Diagnóstico"
    Command "Iniciar Diagnóstico" -> Aggregate "Ordem de Serviço"
    Aggregate "Ordem de Serviço" -> Event "Em Diagnóstico"
    
    Mechanic "Mecânico" -> Command "Registrar Peças e Serviços"
    Command "Registrar Peças e Serviços" -> Aggregate "Ordem de Serviço"
    Aggregate "Ordem de Serviço" -> Event "Orçamento Calculado"
    
    %% Approval Flow
    Manager "Gerente" -> Command "Enviar Orçamento"
    Command "Enviar Orçamento" -> Aggregate "Ordem de Serviço"
    Aggregate "Ordem de Serviço" -> Event "Aguardando Aprovação"
    
    Client "Cliente" -> Command "Aprovar Orçamento"
    Command "Aprovar Orçamento" -> Aggregate "Ordem de Serviço"
    Aggregate "Ordem de Serviço" -> Event "Orçamento Aprovado"
    
    Event "Orçamento Aprovado" -> Policy "Reservar Estoque"
    Policy "Reservar Estoque" -> Command "Baixar Peças"
    Command "Baixar Peças" -> Aggregate "Estoque"
    Aggregate "Estoque" -> Event "Estoque Atualizado"

    %% Execution Flow
    Mechanic "Mecânico" -> Command "Iniciar Execução"
    Command "Iniciar Execução" -> Aggregate "Ordem de Serviço"
    Aggregate "Ordem de Serviço" -> Event "Em Execução"
    
    Mechanic "Mecânico" -> Command "Finalizar Serviço"
    Command "Finalizar Serviço" -> Aggregate "Ordem de Serviço"
    Aggregate "Ordem de Serviço" -> Event "Serviço Finalizado"
    
    Event "Serviço Finalizado" -> Policy "Notificar Cliente"
    
    %% Delivery Flow
    Client "Cliente" -> Command "Retirar Veículo"
    Command "Retirar Veículo" -> Aggregate "Ordem de Serviço"
    Aggregate "Ordem de Serviço" -> Event "Veículo Entregue"
```
