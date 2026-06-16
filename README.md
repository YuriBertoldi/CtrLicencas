# CtrlLicença

Sistema web para controle de licenças Delphi/RAD Studio, componentes de terceiros e desenvolvedores.

## Stack

| Camada     | Tecnologia                              |
|------------|-----------------------------------------|
| Backend    | Go 1.24, `net/http`, `html/template`    |
| Banco      | PostgreSQL 16 (Docker)                  |
| Frontend   | Pico CSS v2, CSS customizado, JS vanilla|
| Infra      | Docker Compose                          |

## Funcionalidades

- **Dashboard** — visão consolidada com cards de stats (devs, licenças, componentes)
- **Desenvolvedores** — cadastro com equipe, status (ativo/livre/inativo) e observações
- **Licenças** — seriais com vínculo a devs, grupos, controle EDN/Network, dados EDN (login/senha)
- **Grupos de Licença** — agrupamento de seriais de versões diferentes (ex: XE3 + D12) da mesma licença física; vincular um dev a qualquer serial do grupo vincula todos automaticamente
- **Componentes** — catálogo de componentes/bibliotecas com licenciamento (pago/free), serial, usuário, senha e site
- **Cadastros Auxiliares** — tabelas de apoio (equipes, versões, tipos de licença, tipo de controle) usadas nos formulários
- **Auditoria** — histórico de todas as alterações com usuário, data e detalhes (valores antigos → novos), exportável em CSV
- **Importar/Exportar** — importação e exportação de desenvolvedores, licenças e componentes via CSV, com download de modelo (CSV só com cabeçalho)
- **Autenticação** — login com sessão via cookie, middleware `Protected` e `AdminOnly`
- **Gestão de Usuários** — criação, ativação/desativação, toggle admin, reset de senha (admin only)
- **Responsivo** — layout com sidebar colapsável, tabelas com scroll horizontal, touch targets para mobile

## Setup

### Pré-requisitos

- Docker e Docker Compose

### Subir o ambiente

```bash
docker compose up -d
```

A aplicação fica disponível em `http://localhost:8081`.

Credenciais padrão: `admin@ctrllicenca.local` / `admin123`

### Dados de demonstração

```bash
docker exec -i <container-db> psql -U ctrllicenca -d ctrllicenca < seed_demo.sql
```

### Variáveis de ambiente

| Variável       | Default                                                                    | Descrição                |
|----------------|----------------------------------------------------------------------------|--------------------------|
| `PORT`         | `8081`                                                                     | Porta da aplicação       |
| `DATABASE_URL` | `postgres://ctrllicenca:ctrllicenca@db:5432/ctrllicenca?sslmode=disable`   | Connection string do PG  |
| `ADMIN_EMAIL`  | `admin@ctrllicenca.local`                                                  | Email do admin inicial   |
| `ADMIN_SENHA`  | `admin123`                                                                 | Senha do admin inicial   |

### Desenvolvimento

Templates e CSS são montados via volume Docker — alterações refletem com reload do browser (CSS) ou restart do container (templates). Alterações em código Go requerem rebuild:

```bash
# Rebuild Go
docker compose up -d --build app

# Restart (recarrega templates sem rebuild)
docker compose restart app
```

## Estrutura do Projeto

```
ctrllicenca/
  main.go                          # Entrypoint, rotas
  Dockerfile                       # Build multi-stage (golang:1.24-alpine → alpine:3.19)
  docker-compose.yml               # Dev: app + PostgreSQL
  docker-compose.prod.yml          # Prod: app only (DB compartilhado)
  seed_demo.sql                    # Dados fictícios para demonstração
  go.mod / go.sum
  internal/
    auth/auth.go                   # Sessões, cookies, middlewares Protected/AdminOnly
    auth/auth_test.go
    handler/handler.go             # Handlers HTTP, template rendering, CSV export/import
    handler/handler_test.go
    models/models.go               # Structs (entidades + page data)
    models/models_test.go
    store/store.go                 # Migrations, CRUD, queries, auditoria
    store/store_test.go
  templates/
    base.html                      # Layout base (sidebar, nav, scripts, tema dark/light)
    login.html                     # Tela de login (layout próprio)
    dashboard.html                 # Dashboard com stats
    desenvolvedores.html           # CRUD devs (dialog unificado)
    licencas.html                  # CRUD licenças + grupos (dialog unificado)
    componentes.html               # CRUD componentes (dialog unificado)
    auxiliares.html                # Cadastros auxiliares (cards com listas)
    auditoria.html                 # Histórico de alterações + export CSV
    importexport.html              # Importar/exportar CSV + download modelo
    usuarios.html                  # Gestão de usuários (admin only)
    network.html                   # Licenças network
  static/
    app.css                        # Design system completo (dark glassmorphism)
```

## Banco de Dados

Migrations são gerenciadas automaticamente via tabela `schema_migrations`. Ao iniciar, a aplicação aplica todas as migrations pendentes.

| Versão | Descrição                                        |
|--------|--------------------------------------------------|
| 1      | Schema inicial: devs, licencas, licencas_network, componentes, usuarios, sessions |
| 2      | dev_id nullable (licença livre), FK ON DELETE SET NULL |
| 3      | Campo canal (EDN/Network)                        |
| 4      | Tabela grupos_licenca + FK grupo_id              |
| 5      | Tabela auxiliares + campo hostname               |
| 6      | Tabela audit_log para auditoria                  |
| 7      | Campos serial, usuario, senha em componentes     |

## Permissões

| Ação | Admin | Usuário |
|------|-------|---------|
| Visualizar dados (licenças, devs, componentes) | Sim | Sim |
| Criar/editar/excluir licenças | Sim | Não |
| Vincular/desvincular dev em licença | Sim | Sim |
| Criar/editar/excluir devs e componentes | Sim | Sim |
| Importar CSV | Sim | Não |
| Exportar CSV / Baixar modelo | Sim | Sim |
| Gerenciar usuários | Sim | Não |

## Testes

```bash
go test ./...
```

Testes de integração (store) requerem o banco PostgreSQL rodando na porta 5433. Caso contrário, são automaticamente ignorados (`t.Skipf`).

## Deploy em Novo Servidor

### Pré-requisitos no servidor

- Docker e Docker Compose instalados
- Porta `8081` liberada no firewall / security list

### Arquivos necessários

Leve apenas estes arquivos para o servidor:

```
ctrllicenca/
  docker-compose.prod.yml    # Compose de produção (renomear para docker-compose.yml)
  Dockerfile                 # Build multi-stage da aplicação
  go.mod
  go.sum
  main.go
  internal/                  # Todo o diretório (auth, handler, models, store)
  templates/                 # Todo o diretório (11 arquivos .html)
  static/                    # Todo o diretório (app.css)
```

> **Não é necessário** levar: `docker-compose.yml` (esse é o de dev local), `seed_demo.sql`, testes (`*_test.go`), `CLAUDE.md`, `SPECS.md`.

### Passo a passo

**1. Criar o `docker-compose.yml` de produção**

Se o servidor **já tem** um PostgreSQL rodando, use o `docker-compose.prod.yml` como base — ele conecta a um banco externo via rede Docker:

```yaml
services:
  app:
    build: .
    restart: unless-stopped
    environment:
      PORT: 8081
      DATABASE_URL: postgres://USUARIO:SENHA@NOME_CONTAINER_POSTGRES:5432/NOME_BANCO?sslmode=disable
      ADMIN_EMAIL: admin@suaempresa.com
      ADMIN_SENHA: suaSenhaSegura
    ports:
      - "8081:8081"
    networks:
      - rede_do_postgres

networks:
  rede_do_postgres:
    external: true
```

Se o servidor **não tem** PostgreSQL, use um compose completo (app + banco):

```yaml
services:
  db:
    image: postgres:16-alpine
    restart: unless-stopped
    environment:
      POSTGRES_USER: ctrllicenca
      POSTGRES_PASSWORD: ctrllicenca
      POSTGRES_DB: ctrllicenca
    volumes:
      - pgdata:/var/lib/postgresql/data

  app:
    build: .
    restart: unless-stopped
    depends_on:
      - db
    environment:
      PORT: 8081
      DATABASE_URL: postgres://ctrllicenca:ctrllicenca@db:5432/ctrllicenca?sslmode=disable
      ADMIN_EMAIL: admin@suaempresa.com
      ADMIN_SENHA: suaSenhaSegura
    ports:
      - "8081:8081"

volumes:
  pgdata:
```

**2. Subir a aplicação**

```bash
docker compose up -d --build
```

**3. Verificar se subiu corretamente**

```bash
docker compose logs app --tail 10
```

Deve mostrar:
```
CtrlLicença iniciado em http://localhost:8081
```

**4. Acessar**

Abra `http://IP_DO_SERVIDOR:8081` no navegador.

Login inicial: o email e senha definidos em `ADMIN_EMAIL` / `ADMIN_SENHA`.

### Banco de dados — criação automática

**Sim, toda a estrutura é criada automaticamente.** Ao iniciar, a aplicação:

1. Aguarda o PostgreSQL ficar disponível (até 10 tentativas, 2s entre cada)
2. Cria a tabela `schema_migrations` se não existir
3. Aplica todas as migrations pendentes em sequência (atualmente 7 versões)
4. Cria o usuário admin se não existir

Não é necessário rodar nenhum SQL manualmente. Basta que o banco PostgreSQL exista e esteja acessível — as tabelas, índices e dados iniciais são criados pela própria aplicação.

### Dados de demonstração (opcional)

Para popular com dados fictícios, copie o `seed_demo.sql` e execute:

```bash
docker exec -i NOME_CONTAINER_POSTGRES psql -U USUARIO -d BANCO < seed_demo.sql
```


## Design

- Tema dark com glassmorphism (com suporte a light mode)
- Paleta: base `#0f1419`, accent `#3498db`
- Dialogs modais unificados (mesma tela para criar e editar)
- Tabelas com filtro por texto e select, colunas ordenáveis com ↕
- Badges coloridos por tipo/status/versão
- Ações em tabelas com ícones 32×32px em grid 2×2 ou 3 colunas
- Layout responsivo (sidebar colapsável em mobile, touch targets 38px)
