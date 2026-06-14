# delphiLic

Sistema web para controle de licenças Delphi, componentes e desenvolvedores.

## Stack

| Camada     | Tecnologia                              |
|------------|-----------------------------------------|
| Backend    | Go 1.24, `net/http`, `html/template`    |
| Banco      | PostgreSQL 16 (Docker)                  |
| Frontend   | HTMX 1.9, Pico CSS v2, CSS customizado |
| Infra      | Docker Compose (app + db)               |

## Funcionalidades

- **Desenvolvedores** — cadastro com equipe, status (ativo/livre/inativo) e observações
- **Licenças Delphi** — seriais XE3, D12, Interbase, HTML5Builder com vínculo a devs
- **Grupos de Licença** — agrupamento de seriais de versões diferentes pertencentes à mesma licença física; vincular um dev a qualquer serial do grupo vincula todos automaticamente
- **Componentes** — catálogo de componentes Delphi com licenciamento (pago/free), serial, usuário e senha
- **Cadastros Auxiliares** — tabelas de apoio (equipes, versões, tipos de versão, tipo de controle) usadas nos formulários principais
- **Auditoria** — histórico de todas as alterações com usuário, data e detalhes (valores antigos → novos), exportável em CSV
- **Importar/Exportar** — importação e exportação de desenvolvedores, licenças e componentes via CSV
- **Dashboard** — visão consolidada de devs, licenças (total, em uso, livres) e componentes
- **Autenticação** — login com sessão via cookie, middleware `Protected` e `AdminOnly`
- **Gestão de Usuários** — criação, ativação/desativação, toggle admin, reset de senha (admin only)

## Setup

### Pré-requisitos

- Docker e Docker Compose

### Subir o ambiente

```bash
docker compose up -d
```

A aplicação fica disponível em `http://localhost:8081`.

Credenciais padrão: `admin@delphilic.local` / `admin123`

### Variáveis de ambiente

| Variável       | Default                                                              | Descrição                |
|----------------|----------------------------------------------------------------------|--------------------------|
| `PORT`         | `8081`                                                               | Porta da aplicação       |
| `DATABASE_URL` | `postgres://delphilic:delphilic@db:5432/delphilic?sslmode=disable`   | Connection string do PG  |
| `ADMIN_EMAIL`  | `admin@delphilic.local`                                              | Email do admin inicial   |
| `ADMIN_SENHA`  | `admin123`                                                           | Senha do admin inicial   |

### Desenvolvimento

Templates e CSS são montados via volume Docker — alterações nestes arquivos refletem com reload do browser. Alterações em código Go requerem rebuild:

```bash
docker compose up -d --build app
```

## Estrutura do Projeto

```
delphiLic/
  main.go                          # Entrypoint, rotas
  Dockerfile                       # Build multi-stage
  docker-compose.yml               # App + PostgreSQL
  go.mod / go.sum
  internal/
    auth/auth.go                   # Sessões, cookies, middlewares Protected/AdminOnly
    handler/handler.go             # Handlers HTTP, template rendering, CSV export/import
    models/models.go               # Structs (entidades + page data)
    store/store.go                 # Migrations, CRUD, queries, auditoria
  templates/
    base.html                      # Layout base (sidebar, scripts)
    dashboard.html                 # Dashboard com stats
    desenvolvedores.html           # CRUD devs (dialog unificado)
    licencas.html                  # CRUD licenças + grupos (dialog unificado)
    componentes.html               # CRUD componentes (dialog unificado)
    auxiliares.html                # Cadastros auxiliares
    auditoria.html                 # Histórico de alterações + export CSV
    importexport.html              # Importar/exportar dados via CSV
    usuarios.html                  # Gestão de usuários (admin)
    network.html                   # Licenças network
    login.html                     # Tela de login
  static/
    app.css                        # Design system (dark glassmorphism)
```

## Banco de Dados

Migrations são gerenciadas automaticamente via tabela `schema_migrations`. Ao iniciar, a aplicação aplica todas as migrations pendentes.

| Versão | Descrição                                        |
|--------|--------------------------------------------------|
| 1      | Tabelas: devs, licencas, licencas_network, componentes, usuarios, sessions |
| 2      | dev_id nullable (licença livre), FK ON DELETE SET NULL |
| 3      | Campo canal (EDN/Network)                        |
| 4      | Tabela grupos_licenca + FK grupo_id              |
| 5      | Tabela auxiliares + campo hostname               |
| 6      | Tabela audit_log para auditoria                  |
| 7      | Campos serial, usuario, senha em componentes     |

## Testes

```bash
go test ./...
```

Testes de integração (store) requerem o banco PostgreSQL rodando na porta 5433. Caso contrário, são automaticamente ignorados (`t.Skipf`).

## Design

- Tema dark com glassmorphism
- Paleta: base `#0f1419`, accent `#3498db`
- Dialogs modais unificados (mesma tela para criar e editar)
- Tabelas com filtro por texto e select, colunas ordenáveis
- Badges coloridos por tipo/status
- Layout responsivo (sidebar colapsável em mobile)
