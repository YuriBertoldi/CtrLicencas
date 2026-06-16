-- ============================================================
-- CtrlLicença — Schema de referência
-- As migrations reais rodam via RunMigrations() ao subir.
-- ============================================================

-- Desenvolvedores
CREATE TABLE devs (
    id        SERIAL PRIMARY KEY,
    nome      VARCHAR(120) NOT NULL,
    equipe    VARCHAR(60)  NOT NULL DEFAULT '',
    status    VARCHAR(20)  NOT NULL DEFAULT 'ativo', -- ativo | livre | inativo
    obs       TEXT         NOT NULL DEFAULT '',
    criado_em TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Licenças por desenvolvedor (XE3, D12, Interbase, HTML5Builder, etc.)
CREATE TABLE licencas (
    id             SERIAL PRIMARY KEY,
    dev_id         INTEGER      NOT NULL REFERENCES devs(id) ON DELETE CASCADE,
    versao         VARCHAR(30)  NOT NULL DEFAULT '', -- XE3 | D12 | Interbase | HTML5Builder
    serial         VARCHAR(120) NOT NULL DEFAULT '',
    tipo           VARCHAR(30)  NOT NULL DEFAULT '', -- Professional | Enterprise | Network
    edn_login      VARCHAR(120) NOT NULL DEFAULT '',
    edn_senha      VARCHAR(120) NOT NULL DEFAULT '',
    cad_efetuado   BOOLEAN      NOT NULL DEFAULT false,
    data_cad       DATE,
    obs            TEXT         NOT NULL DEFAULT '',
    criado_em      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Licenças de servidor de rede (Network License Server interno)
CREATE TABLE licencas_network (
    id          SERIAL PRIMARY KEY,
    sku         VARCHAR(60)  NOT NULL DEFAULT '',
    descricao   VARCHAR(200) NOT NULL DEFAULT '',
    tipo        VARCHAR(30)  NOT NULL DEFAULT '', -- Professional | Enterprise
    total_seats INTEGER      NOT NULL DEFAULT 0,
    login_name  VARCHAR(120) NOT NULL DEFAULT '',
    senha       VARCHAR(120) NOT NULL DEFAULT '',
    obs         TEXT         NOT NULL DEFAULT '',
    criado_em   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Componentes (TMS, FastReport, ACBR, etc.)
CREATE TABLE componentes (
    id           SERIAL PRIMARY KEY,
    nome         VARCHAR(120) NOT NULL,
    versao       VARCHAR(60)  NOT NULL DEFAULT '',
    ferramenta   VARCHAR(60)  NOT NULL DEFAULT '',
    informacoes  TEXT         NOT NULL DEFAULT '',
    uso          VARCHAR(200) NOT NULL DEFAULT '',
    licenciamento VARCHAR(30) NOT NULL DEFAULT '', -- Pago | Free | Pago Por uso
    tipo_licenca VARCHAR(50)  NOT NULL DEFAULT '', -- Aquisição | Nativo | Grátis com embarcadero
    site         VARCHAR(200) NOT NULL DEFAULT '',
    status       VARCHAR(30)  NOT NULL DEFAULT '', -- Documentado | (vazio)
    obs          TEXT         NOT NULL DEFAULT '',
    diretorio    VARCHAR(200) NOT NULL DEFAULT '',
    criado_em    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Usuários do sistema
CREATE TABLE usuarios (
    id         SERIAL PRIMARY KEY,
    nome       VARCHAR(100) NOT NULL,
    email      VARCHAR(150) UNIQUE NOT NULL,
    senha_hash TEXT         NOT NULL,
    admin      BOOLEAN      NOT NULL DEFAULT false,
    ativo      BOOLEAN      NOT NULL DEFAULT true,
    criado_em  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Sessões
CREATE TABLE sessions (
    token     TEXT PRIMARY KEY,
    user_id   INTEGER NOT NULL REFERENCES usuarios(id) ON DELETE CASCADE,
    expira_em TIMESTAMPTZ NOT NULL,
    criado_em TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Controle de migrations
CREATE TABLE schema_migrations (
    version    INTEGER PRIMARY KEY,
    name       VARCHAR(200) NOT NULL,
    applied_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_licencas_dev       ON licencas(dev_id);
CREATE INDEX idx_licencas_versao    ON licencas(versao);
CREATE INDEX idx_componentes_ferr   ON componentes(ferramenta);
CREATE INDEX idx_componentes_lic    ON componentes(licenciamento);
