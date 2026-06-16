# CtrlLicença — Instruções para Claude Code

## Visão Geral

Sistema web Go para controle de licenças Delphi/RAD Studio, componentes de terceiros e desenvolvedores. Monolito server-side com html/template + JS vanilla (sem HTMX real, apenas forms tradicionais).

## Comandos

```bash
# Build e rodar local
docker compose up -d --build app

# Apenas restart (recarrega templates sem rebuild Go)
docker compose restart app

# Testes
go test ./...

# Compilar localmente (verificar erros)
go build ./...

# Deploy produção (ver README.md para passo a passo completo)
docker compose up -d --build
```

## Arquitetura

```
main.go                          # Entrypoint, rotas (net/http ServeMux)
internal/
  auth/auth.go                   # Sessões cookie, middlewares Protected/AdminOnly
  handler/handler.go             # Handlers HTTP, templates, CSV export/import, auditoria
  models/models.go               # Structs (entidades + page data)
  store/store.go                 # Migrations, CRUD, queries PostgreSQL
templates/                       # html/template (base.html + pages)
static/app.css                   # Design system (dark glassmorphism, Pico CSS overrides)
```

- **Monolito MVC**: `handler` (controllers) → `store` (queries) → `models` (structs)
- **Templates**: `html/template` com layout `base.html` + pages. Registrados em `InitTemplates()`.
- **CSS**: Pico CSS v2 classless + `static/app.css` (design system customizado com `!important` overrides)
- **Auth**: Cookie `ctrllic_session`, middlewares `Protected` (login obrigatório) e `AdminOnly` (admin obrigatório)
- **Migrations**: Sequenciais em `store.go`, tabela `schema_migrations` (atualmente 7 versões)
- **Auditoria**: `audit_log` com diff de campos — `diffFields()` compara antes/depois

## Convenções

- **Idioma do código**: variáveis e funções em inglês/português misto (ex: `HandleCreateLicenca`, `DevNome`)
- **Idioma da UI**: Português brasileiro com acentos corretos (licença, versão, ação)
- **Nullable fields**: `*int` nos models, `sql.NullInt64` no scan
- **Templates**: Registrados em `InitTemplates()` — ao adicionar nova página, incluir no array `pages`
- **Rotas**: Definidas em `main.go` com pattern `METHOD /path`
- **CSS**: Pico CSS aplica `width: 100%` em buttons/inputs — usar `!important` para overrides em `.btn-sm`, `.btn-icon`, etc.
- **Docker volumes**: `./templates` e `./static` são montados — CSS/HTML changes sem rebuild, mas precisa restart para Go recarregar templates
- **Dialog unificado**: Cada módulo (devs, licenças, componentes) usa um único `<dialog>` para criar e editar. JS preenche campos via `data-*` attributes e altera `action` do form.
- **Auditoria com diff**: Handlers de update buscam registro antigo antes de salvar, comparam via `diffFields()`, registram no formato `Campo: [antigo] → [novo]`.
- **Ações em tabelas**: Usar `btn-icon` (32×32px) com `actions-grid` (2 ou 3 colunas). Formulários dentro usam `display: contents`. Incluir legenda abaixo da tabela quando os ícones não são óbvios.
- **Export modelo**: Rota `/importexport/export?tipo=X&modelo=true` retorna CSV só com cabeçalho (sem dados).

## Permissões (Admin vs Usuário)

| Ação | Admin | Usuário |
|------|-------|---------|
| Ver licenças, devs, componentes | Sim | Sim |
| Criar/editar/excluir licenças | Sim | Não |
| Vincular/desvincular dev em licença | Sim | Sim |
| Criar/editar/excluir devs | Sim | Sim |
| Importar CSV | Sim | Não |
| Exportar CSV / Modelo | Sim | Sim |
| Gerenciar usuários | Sim | Não |

- **Server-side**: Rotas protegidas com `auth.AdminOnly(db, handler)` em `main.go`
- **Template-side**: Botões ocultos com `{{if .CurrentUser.Admin}}` ou `{{if $.CurrentUser.Admin}}`

## Regras de Negócio

- Licença com `dev_id = NULL` está **livre**
- Excluir dev **não** exclui licenças — FK usa `ON DELETE SET NULL` (licenças ficam livres)
- Grupos de licença: vincular dev a um serial vincula **todos** os seriais do mesmo grupo
- Auxiliares (equipe, versao, tipo_lic, canal) alimentam selects dos formulários
- Auditoria registra criar/editar/excluir/vincular/desvincular com valores antigos e novos
- Export modelo: CSV com apenas cabeçalho para servir de template de importação

## Módulos e Templates

| Módulo | Template | Rota GET | Descrição |
|--------|----------|----------|-----------|
| Dashboard | `dashboard.html` | `/` | Cards de stats consolidados |
| Desenvolvedores | `desenvolvedores.html` | `/desenvolvedores` | CRUD devs com equipe, status |
| Licenças | `licencas.html` | `/licencas` | CRUD licenças + grupos + vincular dev |
| Componentes | `componentes.html` | `/componentes` | Catálogo com serial, usuário, senha |
| Auxiliares | `auxiliares.html` | `/auxiliares` | Tabelas de apoio (equipe, versao, tipo, canal) |
| Auditoria | `auditoria.html` | `/auditoria` | Histórico de alterações + export CSV |
| Import/Export | `importexport.html` | `/importexport` | CSV import/export + download modelo |
| Usuários | `usuarios.html` | `/usuarios` | Gestão de usuários (admin only) |
| Network | `network.html` | — | Licenças network (sem rota registrada atualmente) |
| Login | `login.html` | `/login` | Tela de login (layout próprio, sem base.html) |

## Banco de Dados — Migrations

| Versão | Nome | Descrição |
|--------|------|-----------|
| 1 | criar_schema_inicial | devs, licencas, licencas_network, componentes, usuarios, sessions |
| 2 | licencas_dev_nullable | dev_id nullable, FK ON DELETE SET NULL |
| 3 | licencas_add_canal | Campo canal (EDN/Network) |
| 4 | grupos_licenca | Tabela grupos_licenca + FK grupo_id |
| 5 | auxiliares_e_hostname | Tabela auxiliares + campo hostname |
| 6 | audit_log | Tabela audit_log para auditoria |
| 7 | componentes_serial_usuario_senha | Campos serial, usuario, senha em componentes |

## Deploy

### Local
```bash
docker compose up -d --build app
# Acesso: http://localhost:8081
# Admin: admin@ctrllicenca.local / admin123
# DB: PostgreSQL na porta 5433
```

### Produção
- Ver README.md para passo a passo completo de deploy em novo servidor
- Porta padrão: 8081
- Banco e tabelas são criados automaticamente ao iniciar

## Cuidados

- Ao adicionar nova página: incluir nome em `pages` no `InitTemplates()` em `handler.go`
- Ao adicionar nova migration: incrementar versão, adicionar no array `migrations` em `store.go`
- Ao adicionar nova rota: registrar em `main.go` com middleware adequado (`Protected` ou `AdminOnly`)
- CSS: Pico CSS briga com botões/inputs — sempre usar `!important` em overrides customizados
- Templates são lidos no startup — alterar HTML requer `docker compose restart app`
- CSS/JS são servidos como estáticos — alterar CSS só precisa reload no browser
- Dialogs: usar padrão unificado com JS (não duplicar dialogs para create/edit)
- Não remover `LicencasNetwork` do store — template `network.html` existe mesmo sem rota ativa
- Deploy: ao buildar para servidor remoto, verificar a arquitetura (`uname -m`) e usar `--platform` adequado no buildx
