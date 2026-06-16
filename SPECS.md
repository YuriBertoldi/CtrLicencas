# CtrlLicença — Especificações Técnicas

## 1. Entidades e Relacionamentos

### 1.1 Desenvolvedores (`devs`)
| Campo | Tipo | Obrigatório | Descrição |
|-------|------|-------------|-----------|
| id | SERIAL PK | auto | |
| nome | VARCHAR(120) | Sim | Nome completo |
| equipe | VARCHAR(60) | Não | Alimentado por auxiliares tipo=equipe |
| status | VARCHAR(20) | Sim | `ativo` \| `livre` \| `inativo` |
| obs | TEXT | Não | Observações |
| criado_em | TIMESTAMPTZ | auto | |

**Regras:**
- Ao excluir dev, licenças vinculadas ficam livres (`ON DELETE SET NULL`)
- Status é exibido como badge colorido (green/blue/gray)

### 1.2 Licenças (`licencas`)
| Campo | Tipo | Obrigatório | Descrição |
|-------|------|-------------|-----------|
| id | SERIAL PK | auto | |
| dev_id | INTEGER FK | Não | NULL = licença livre |
| grupo_id | INTEGER FK | Não | Grupo de licença física |
| versao | VARCHAR(30) | Sim | XE3, D12, Interbase, HTML5Builder |
| serial | VARCHAR(120) | Não | Chave serial |
| tipo | VARCHAR(30) | Não | Professional, Enterprise |
| canal | VARCHAR(10) | Sim | EDN, Network (default: EDN) |
| hostname | VARCHAR(120) | Não | Nome da máquina |
| edn_login | VARCHAR(120) | Não | Login EDN |
| edn_senha | VARCHAR(120) | Não | Senha EDN (armazenada em texto) |
| cad_efetuado | BOOLEAN | Sim | Cadastro EDN efetuado |
| data_cad | DATE | Não | Data do cadastro |
| obs | TEXT | Não | |
| criado_em | TIMESTAMPTZ | auto | |

**Regras:**
- `dev_id = NULL` → licença está livre
- Vincular dev a licença de um grupo vincula todas do grupo automaticamente
- Somente admin pode criar/editar/excluir (rota `AdminOnly`)
- Qualquer usuário pode vincular/desvincular dev (rota `Protected`)

### 1.3 Grupos de Licença (`grupos_licenca`)
| Campo | Tipo | Obrigatório | Descrição |
|-------|------|-------------|-----------|
| id | SERIAL PK | auto | |
| nome | VARCHAR(120) | Sim | Ex: "Licença Física 01" |
| obs | TEXT | Não | |
| criado_em | TIMESTAMPTZ | auto | |

**Regras:**
- Ao excluir grupo, licenças ficam sem grupo (`ON DELETE SET NULL`)
- Vincular dev a um serial do grupo → todas as licenças do grupo recebem o mesmo dev

### 1.4 Componentes (`componentes`)
| Campo | Tipo | Obrigatório | Descrição |
|-------|------|-------------|-----------|
| id | SERIAL PK | auto | |
| nome | VARCHAR(120) | Sim | Nome do componente |
| versao | VARCHAR(30) | Não | Versão do componente |
| ferramenta | VARCHAR(30) | Não | D12, XE3, etc |
| informacoes | TEXT | Não | Descrição detalhada |
| uso | VARCHAR(120) | Não | Onde é utilizado |
| licenciamento | VARCHAR(30) | Não | Pago, Free |
| site | VARCHAR(255) | Não | URL do site |
| serial | VARCHAR(255) | Não | Chave serial |
| usuario | VARCHAR(120) | Não | Login da conta |
| senha | VARCHAR(120) | Não | Senha da conta |
| obs | TEXT | Não | |
| criado_em | TIMESTAMPTZ | auto | |

### 1.5 Auxiliares (`auxiliares`)
| Campo | Tipo | Obrigatório | Descrição |
|-------|------|-------------|-----------|
| id | SERIAL PK | auto | |
| tipo | VARCHAR(30) | Sim | `equipe` \| `versao` \| `tipo_lic` \| `canal` |
| nome | VARCHAR(120) | Sim | Valor do auxiliar |
| criado_em | TIMESTAMPTZ | auto | |

**Tipos:**
- `equipe` → select de equipe em devs
- `versao` → select de versão em licenças (XE3, D12, Interbase, HTML5Builder)
- `tipo_lic` → select de tipo em licenças (Professional, Enterprise, Network)
- `canal` → select de controle em licenças (EDN, Network)

### 1.6 Licenças Network (`licencas_network`)
| Campo | Tipo | Obrigatório | Descrição |
|-------|------|-------------|-----------|
| id | SERIAL PK | auto | |
| sku | VARCHAR(60) | Sim | Código do produto |
| descricao | VARCHAR(255) | Não | |
| tipo | VARCHAR(30) | Não | Professional, Enterprise |
| total_seats | INTEGER | Sim | Número de assentos |
| login_name | VARCHAR(120) | Não | |
| senha | VARCHAR(120) | Não | |
| obs | TEXT | Não | |
| criado_em | TIMESTAMPTZ | auto | |

### 1.7 Usuários (`usuarios`)
| Campo | Tipo | Obrigatório | Descrição |
|-------|------|-------------|-----------|
| id | SERIAL PK | auto | |
| nome | VARCHAR(100) | Sim | |
| email | VARCHAR(100) UNIQUE | Sim | Usado no login |
| senha_hash | VARCHAR(255) | Sim | bcrypt hash |
| admin | BOOLEAN | Sim | default false |
| ativo | BOOLEAN | Sim | default true |
| criado_em | TIMESTAMPTZ | auto | |

### 1.8 Sessões (`sessions`)
| Campo | Tipo | Descrição |
|-------|------|-----------|
| token | VARCHAR(64) PK | Token da sessão (cookie) |
| user_id | INTEGER FK | Usuário autenticado |
| criado_em | TIMESTAMPTZ | |

### 1.9 Auditoria (`audit_log`)
| Campo | Tipo | Descrição |
|-------|------|-----------|
| id | SERIAL PK | |
| entidade | VARCHAR(30) | licenca, dev, componente, grupo |
| entidade_id | INTEGER | ID do registro afetado |
| acao | VARCHAR(30) | criar, editar, excluir, vincular, desvincular |
| usuario | VARCHAR(100) | Nome do usuário que realizou a ação |
| detalhes | TEXT | Diff no formato `Campo: [antigo] → [novo]` |
| criado_em | TIMESTAMPTZ | |

## 2. Rotas da API

### Autenticação
| Método | Rota | Middleware | Handler |
|--------|------|-----------|---------|
| GET | `/login` | — | HandleLogin |
| POST | `/login` | — | HandleLogin |
| POST | `/logout` | — | HandleLogout |

### Dashboard
| Método | Rota | Middleware | Handler |
|--------|------|-----------|---------|
| GET | `/` | Protected | HandleDashboard |

### Desenvolvedores
| Método | Rota | Middleware | Handler |
|--------|------|-----------|---------|
| GET | `/desenvolvedores` | Protected | HandleDevs |
| POST | `/desenvolvedores` | Protected | HandleCreateDev |
| POST | `/desenvolvedores/{id}/update` | Protected | HandleUpdateDev |
| POST | `/desenvolvedores/{id}/delete` | Protected | HandleDeleteDev |

### Licenças
| Método | Rota | Middleware | Handler |
|--------|------|-----------|---------|
| GET | `/licencas` | Protected | HandleLicencas |
| POST | `/licencas` | **AdminOnly** | HandleCreateLicenca |
| POST | `/licencas/{id}/update` | **AdminOnly** | HandleUpdateLicenca |
| POST | `/licencas/{id}/vincular` | Protected | HandleVincularDevLicenca |
| POST | `/licencas/{id}/delete` | **AdminOnly** | HandleDeleteLicenca |

### Grupos de Licença
| Método | Rota | Middleware | Handler |
|--------|------|-----------|---------|
| POST | `/grupos` | **AdminOnly** | HandleCreateGrupo |
| POST | `/grupos/{id}/delete` | **AdminOnly** | HandleDeleteGrupo |

### Componentes
| Método | Rota | Middleware | Handler |
|--------|------|-----------|---------|
| GET | `/componentes` | Protected | HandleComponentes |
| POST | `/componentes` | Protected | HandleCreateComponente |
| POST | `/componentes/{id}/update` | Protected | HandleUpdateComponente |
| POST | `/componentes/{id}/delete` | Protected | HandleDeleteComponente |

### Auxiliares
| Método | Rota | Middleware | Handler |
|--------|------|-----------|---------|
| GET | `/auxiliares` | Protected | HandleAuxiliares |
| POST | `/auxiliares` | Protected | HandleCreateAuxiliar |
| POST | `/auxiliares/{id}/delete` | Protected | HandleDeleteAuxiliar |

### Auditoria
| Método | Rota | Middleware | Handler |
|--------|------|-----------|---------|
| GET | `/auditoria` | Protected | HandleAuditoria |
| GET | `/auditoria/export` | Protected | HandleExportAuditCSV |

### Importar / Exportar
| Método | Rota | Middleware | Handler | Observação |
|--------|------|-----------|---------|------------|
| GET | `/importexport` | Protected | HandleImportExport | Página |
| GET | `/importexport/export?tipo=X` | Protected | HandleExportCSV | Dados completos |
| GET | `/importexport/export?tipo=X&modelo=true` | Protected | HandleExportCSV | Só cabeçalho |
| POST | `/importexport/import` | **AdminOnly** | HandleImportCSV | Upload CSV |

**Tipos válidos:** `desenvolvedores`, `licencas`, `componentes`

### Usuários (Admin Only)
| Método | Rota | Middleware | Handler |
|--------|------|-----------|---------|
| GET | `/usuarios` | AdminOnly | HandleUsuarios |
| POST | `/usuarios` | AdminOnly | HandleCreateUsuario |
| POST | `/usuarios/{id}/toggle-ativo` | AdminOnly | HandleToggleUsuarioAtivo |
| POST | `/usuarios/{id}/toggle-admin` | AdminOnly | HandleToggleUsuarioAdmin |
| POST | `/usuarios/{id}/reset-senha` | AdminOnly | HandleResetSenha |
| POST | `/minha-senha` | Protected | HandleMinhaSenha |

## 3. Formato CSV (Import/Export)

### Desenvolvedores
```
Nome;Equipe;Status;Obs
```

### Licenças
```
Dev;Grupo;Versão;Serial;Tipo;Controle;Hostname;EDN Login;EDN Senha;Cad. Efetuado;Data Cad.;Obs
```
- **Dev** e **Grupo**: vinculados pelo nome. Se não encontrados, a licença é criada sem vínculo.
- **Cad. Efetuado**: `Sim` ou `Não`
- **Data Cad.**: formato `dd/mm/yyyy`

### Componentes
```
Nome;Versão;Ferramenta;Uso;Licenciamento;Serial;Usuário;Senha;Site;Informações;Obs
```

**Regras gerais:**
- Separador: ponto-e-vírgula (`;`)
- Encoding: UTF-8 com BOM (compatível com Excel)
- Primeira linha é cabeçalho (ignorada na importação)

## 4. Design System (CSS)

### Paleta
| Token | Valor | Uso |
|-------|-------|-----|
| `--bg-base` | `#0f1419` | Fundo principal |
| `--bg-elev-1` | `rgba(26,32,44,0.85)` | Cards, tabelas |
| `--bg-elev-2` | `rgba(15,20,25,0.6)` | Inputs, botões secondary |
| `--color-accent` | `#3498db` | Primário (azul) |
| `--color-green` | `#27ae60` | Sucesso, ativo |
| `--color-red` | `#e74c3c` | Erro, excluir |
| `--color-purple` | `#9b59b6` | Admin, D12 |
| `--color-orange` | `#f39c12` | Alerta, Interbase |

### Componentes CSS
| Classe | Uso |
|--------|-----|
| `.btn-primary` | Botão principal (38px height, gradient azul) |
| `.btn-secondary` | Botão secundário (38px height, borda) |
| `.btn-sm` | Botão compacto (30px height, com `!important`) |
| `.btn-icon` | Botão ícone 32×32px em tabelas |
| `.actions-grid` | Grid 2 colunas de 32px para ações |
| `.actions-grid--3` | Grid 3 colunas de 32px |
| `.col-actions` | Célula de ações (76px min-width) |
| `.badge badge--{cor}` | Badge colorido (green, blue, purple, orange, gray, red, teal, indigo) |
| `.table-wrapper` | Wrapper com overflow-x auto |
| `.toolbar` | Barra de filtros (search + selects) |
| `.stat-card` | Card de estatística no dashboard |
| `.data-view` / `.data-row` | Visualização de dados em dialog |
| `.form-grid-2` | Grid 2 colunas para formulários |
| `.ie-card` | Card de import/export |

### Breakpoints Responsivos
| Largura | Comportamento |
|---------|--------------|
| ≤ 960px | Sidebar colapsa, toggle hamburger |
| ≤ 768px | Page header empilha, dialog compacto |
| ≤ 600px | Touch targets 38px, toolbar empilha, dialog fullscreen |
| ≤ 480px | Padding mínimo, form grid 1 coluna |

## 5. Fluxos Importantes

### 5.1 Vincular Dev a Licença com Grupo
1. Usuário clica no ícone 🔗 da licença
2. Dialog mostra select com devs (incluindo "Livre / Desvincular")
3. POST `/licencas/{id}/vincular` com `dev_id`
4. Handler verifica se licença tem `grupo_id`
5. Se tem grupo: busca todas as licenças do grupo e vincula o mesmo dev a todas
6. Se não tem: vincula apenas a licença individual
7. Auditoria registra vincular/desvincular com detalhes

### 5.2 Importação CSV
1. Admin acessa `/importexport`, pode baixar modelo (CSV só com cabeçalho)
2. Seleciona arquivo CSV e clica "Importar"
3. POST `/importexport/import` com multipart form
4. Handler lê CSV, pula cabeçalho
5. Para licenças: busca dev e grupo pelo nome para vincular
6. Insere registros no banco
7. Redireciona com mensagem de sucesso/erro

### 5.3 Auditoria
1. Todo create/update/delete chama `store.InsertAuditLog()`
2. Para updates: handler busca registro antigo, compara com novo via `diffFields()`
3. Apenas campos alterados são registrados no formato `Campo: [antigo] → [novo]`
4. Auditoria é visualizada em `/auditoria` com paginação e filtros por entidade
5. Export CSV disponível em `/auditoria/export`
