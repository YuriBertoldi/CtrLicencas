# delphiLic — Instruções para Claude Code

## Visão Geral

Sistema web Go para controle de licenças Delphi. Monolito server-side com html/template + HTMX.

## Comandos

```bash
# Build e rodar
docker compose up -d --build app

# Apenas restart (sem rebuild Go)
docker compose restart app

# Testes
go test ./...

# Compilar localmente (verificar erros)
go build ./...
```

## Arquitetura

- **Monolito MVC**: `handler` (controllers) → `store` (queries) → `models` (structs)
- **Templates**: `html/template` com layout `base.html` + pages
- **CSS**: Pico CSS v2 classless + `static/app.css` (design system customizado)
- **Auth**: Cookie `dlicsession`, middlewares `Protected` e `AdminOnly`
- **Migrations**: Sequenciais em `store.go`, tabela `schema_migrations` (atualmente 7 versões)

## Convenções

- **Idioma do código**: variáveis e funções em inglês/português misto (ex: `HandleCreateLicenca`, `DevNome`)
- **Idioma da UI**: Português brasileiro com acentos corretos (licença, versão, ação)
- **Nullable fields**: `*int` nos models, `sql.NullInt64` no scan
- **Templates**: Registrados em `InitTemplates()`, array `pages` deve incluir novos templates
- **Rotas**: Definidas em `main.go` com pattern `METHOD /path`
- **CSS**: Pico CSS aplica `width: 100%` em inputs — usar `!important` para overrides
- **Docker volumes**: `./templates` e `./static` são montados — CSS/HTML changes sem rebuild
- **Dialog unificado**: Cada módulo (devs, licenças, componentes) usa um único dialog HTML para criar e editar. O JavaScript preenche os campos via `data-*` attributes e altera o `action` do form.
- **Auditoria com diff**: Handlers de update buscam o registro antigo antes de salvar, comparam campo a campo via `diffFields()`, e registram apenas os campos que mudaram no formato `Campo: [antigo] → [novo]`.

## Regras de Negócio

- Licença com `dev_id = NULL` está livre
- Excluir dev **não** exclui licenças — FK usa `ON DELETE SET NULL` (licenças ficam livres)
- Grupos de licença: vincular dev a um serial vincula todos do mesmo grupo
- Auxiliares (equipe, versao, tipo_lic, canal) alimentam selects dos formulários
- Auditoria registra criar/editar/excluir/vincular/desvincular com valores antigos e novos

## Módulos

- **Desenvolvedores** — CRUD devs com equipe, status, observações
- **Licenças Delphi** — seriais com vínculo a devs, grupos, controle EDN/Network
- **Componentes** — catálogo com serial, usuário, senha, licenciamento
- **Auditoria** — histórico de alterações com export CSV
- **Importar/Exportar** — CSV para devs, licenças e componentes
- **Auxiliares** — tabelas de apoio para selects
- **Dashboard** — visão consolidada com cards de stats

## Cuidados

- Ao adicionar nova página: incluir nome em `pages` no `InitTemplates()`
- Ao adicionar nova migration: incrementar versão, adicionar no array `migrations`
- Ao adicionar nova rota: registrar em `main.go` com middleware adequado
- CSS: testar com Pico CSS — botões e inputs dentro de forms/dialogs podem precisar de `!important`
- Não remover LicencasNetwork do store — ainda usado pela página `/network`
- Dialogs: usar o padrão de dialog unificado com JS (não duplicar dialogs para create/edit)
