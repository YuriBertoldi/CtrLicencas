package store

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"

	"ctrllicenca/internal/models"
)

func NewDB() *sql.DB {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://ctrllicenca:ctrllicenca@localhost:5433/ctrllicenca?sslmode=disable"
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("db open:", err)
	}
	for i := 0; i < 10; i++ {
		if err = db.Ping(); err == nil {
			break
		}
		log.Printf("aguardando banco... (%d/10)", i+1)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatal("db ping:", err)
	}
	return db
}

// ============================================================
// Migrations
// ============================================================

type migration struct {
	version int
	name    string
	sql     string
}

var migrations = []migration{
	{1, "criar_schema_inicial", sqlSchema},
	{2, "licencas_dev_nullable", sqlMigration2},
	{3, "licencas_add_canal", sqlMigration3},
	{4, "grupos_licenca", sqlMigration4},
	{5, "auxiliares_e_hostname", sqlMigration5},
	{6, "audit_log", sqlMigration6},
	{7, "componentes_serial_usuario_senha", sqlMigration7},
}

const sqlMigration2 = `
ALTER TABLE licencas ALTER COLUMN dev_id DROP NOT NULL;
ALTER TABLE licencas DROP CONSTRAINT IF EXISTS licencas_dev_id_fkey;
ALTER TABLE licencas ADD CONSTRAINT licencas_dev_id_fkey FOREIGN KEY (dev_id) REFERENCES devs(id) ON DELETE SET NULL;
`

const sqlMigration3 = `
ALTER TABLE licencas ADD COLUMN IF NOT EXISTS canal VARCHAR(10) NOT NULL DEFAULT 'EDN';
`

const sqlMigration4 = `
CREATE TABLE IF NOT EXISTS grupos_licenca (
    id        SERIAL PRIMARY KEY,
    nome      VARCHAR(120) NOT NULL,
    obs       TEXT         NOT NULL DEFAULT '',
    criado_em TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
ALTER TABLE licencas ADD COLUMN IF NOT EXISTS grupo_id INTEGER REFERENCES grupos_licenca(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_licencas_grupo ON licencas(grupo_id);
`

const sqlMigration5 = `
CREATE TABLE IF NOT EXISTS auxiliares (
    id        SERIAL PRIMARY KEY,
    tipo      VARCHAR(30)  NOT NULL,
    nome      VARCHAR(120) NOT NULL,
    criado_em TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_auxiliares_tipo ON auxiliares(tipo);
INSERT INTO auxiliares (tipo, nome)
SELECT tipo, nome FROM (VALUES
    ('versao',   'XE3'),
    ('versao',   'D12'),
    ('versao',   'Interbase'),
    ('versao',   'HTML5Builder'),
    ('tipo_lic', 'Professional'),
    ('tipo_lic', 'Enterprise'),
    ('tipo_lic', 'Network'),
    ('canal',    'EDN'),
    ('canal',    'Network')
) AS seed(tipo, nome)
WHERE NOT EXISTS (SELECT 1 FROM auxiliares LIMIT 1);
ALTER TABLE licencas ADD COLUMN IF NOT EXISTS hostname VARCHAR(120) NOT NULL DEFAULT '';
`

const sqlMigration6 = `
CREATE TABLE IF NOT EXISTS audit_log (
    id          SERIAL PRIMARY KEY,
    entidade    VARCHAR(50)  NOT NULL,
    entidade_id INTEGER      NOT NULL,
    acao        VARCHAR(50)  NOT NULL,
    usuario     VARCHAR(120) NOT NULL DEFAULT '',
    detalhes    TEXT         NOT NULL DEFAULT '',
    criado_em   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_audit_entidade ON audit_log(entidade, entidade_id);
CREATE INDEX IF NOT EXISTS idx_audit_criado ON audit_log(criado_em DESC);
`

const sqlMigration7 = `
ALTER TABLE componentes ADD COLUMN IF NOT EXISTS serial VARCHAR(120) NOT NULL DEFAULT '';
ALTER TABLE componentes ADD COLUMN IF NOT EXISTS usuario VARCHAR(120) NOT NULL DEFAULT '';
ALTER TABLE componentes ADD COLUMN IF NOT EXISTS senha VARCHAR(120) NOT NULL DEFAULT '';
`

const sqlSchema = `
CREATE TABLE IF NOT EXISTS devs (
    id        SERIAL PRIMARY KEY,
    nome      VARCHAR(120) NOT NULL,
    equipe    VARCHAR(60)  NOT NULL DEFAULT '',
    status    VARCHAR(20)  NOT NULL DEFAULT 'ativo',
    obs       TEXT         NOT NULL DEFAULT '',
    criado_em TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS licencas (
    id           SERIAL PRIMARY KEY,
    dev_id       INTEGER      NOT NULL REFERENCES devs(id) ON DELETE CASCADE,
    versao       VARCHAR(30)  NOT NULL DEFAULT '',
    serial       VARCHAR(120) NOT NULL DEFAULT '',
    tipo         VARCHAR(30)  NOT NULL DEFAULT '',
    edn_login    VARCHAR(120) NOT NULL DEFAULT '',
    edn_senha    VARCHAR(120) NOT NULL DEFAULT '',
    cad_efetuado BOOLEAN      NOT NULL DEFAULT false,
    data_cad     DATE,
    obs          TEXT         NOT NULL DEFAULT '',
    criado_em    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS licencas_network (
    id          SERIAL PRIMARY KEY,
    sku         VARCHAR(60)  NOT NULL DEFAULT '',
    descricao   VARCHAR(200) NOT NULL DEFAULT '',
    tipo        VARCHAR(30)  NOT NULL DEFAULT '',
    total_seats INTEGER      NOT NULL DEFAULT 0,
    login_name  VARCHAR(120) NOT NULL DEFAULT '',
    senha       VARCHAR(120) NOT NULL DEFAULT '',
    obs         TEXT         NOT NULL DEFAULT '',
    criado_em   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS componentes (
    id            SERIAL PRIMARY KEY,
    nome          VARCHAR(120) NOT NULL,
    versao        VARCHAR(60)  NOT NULL DEFAULT '',
    ferramenta    VARCHAR(60)  NOT NULL DEFAULT '',
    informacoes   TEXT         NOT NULL DEFAULT '',
    uso           VARCHAR(200) NOT NULL DEFAULT '',
    licenciamento VARCHAR(30)  NOT NULL DEFAULT '',
    tipo_licenca  VARCHAR(50)  NOT NULL DEFAULT '',
    site          VARCHAR(200) NOT NULL DEFAULT '',
    status        VARCHAR(30)  NOT NULL DEFAULT '',
    obs           TEXT         NOT NULL DEFAULT '',
    diretorio     VARCHAR(200) NOT NULL DEFAULT '',
    criado_em     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS usuarios (
    id         SERIAL PRIMARY KEY,
    nome       VARCHAR(100) NOT NULL,
    email      VARCHAR(150) UNIQUE NOT NULL,
    senha_hash TEXT         NOT NULL,
    admin      BOOLEAN      NOT NULL DEFAULT false,
    ativo      BOOLEAN      NOT NULL DEFAULT true,
    criado_em  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS sessions (
    token     TEXT PRIMARY KEY,
    user_id   INTEGER NOT NULL REFERENCES usuarios(id) ON DELETE CASCADE,
    expira_em TIMESTAMPTZ NOT NULL,
    criado_em TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_licencas_dev     ON licencas(dev_id);
CREATE INDEX IF NOT EXISTS idx_licencas_versao  ON licencas(versao);
CREATE INDEX IF NOT EXISTS idx_componentes_ferr ON componentes(ferramenta);
`

func RunMigrations(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version    INTEGER PRIMARY KEY,
		name       VARCHAR(200) NOT NULL,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`)
	if err != nil {
		return err
	}
	for _, m := range migrations {
		var exists bool
		db.QueryRow(`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version=$1)`, m.version).Scan(&exists)
		if exists {
			continue
		}
		if _, err := db.Exec(m.sql); err != nil {
			return err
		}
		db.Exec(`INSERT INTO schema_migrations (version, name) VALUES ($1, $2)`, m.version, m.name)
		log.Printf("migration %d (%s) aplicada", m.version, m.name)
	}
	return nil
}

func EnsureAdmin(db *sql.DB) error {
	email := os.Getenv("ADMIN_EMAIL")
	senha := os.Getenv("ADMIN_SENHA")
	if email == "" {
		email = "admin@ctrllicenca.local"
	}
	if senha == "" {
		senha = "admin123"
	}
	var exists bool
	db.QueryRow(`SELECT EXISTS(SELECT 1 FROM usuarios WHERE email=$1)`, email).Scan(&exists)
	if exists {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = db.Exec(`INSERT INTO usuarios (nome, email, senha_hash, admin) VALUES ($1, $2, $3, true)`,
		"Admin", email, string(hash))
	if err != nil {
		return err
	}
	log.Printf("admin criado: %s / %s", email, senha)
	return nil
}

// ============================================================
// Devs
// ============================================================

func ListDevs(db *sql.DB) ([]models.Dev, error) {
	rows, err := db.Query(`SELECT id, nome, equipe, status, obs, criado_em FROM devs ORDER BY nome`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.Dev
	for rows.Next() {
		var d models.Dev
		rows.Scan(&d.ID, &d.Nome, &d.Equipe, &d.Status, &d.Obs, &d.CriadoEm)
		list = append(list, d)
	}
	return list, nil
}

func GetDev(db *sql.DB, id int) (models.Dev, error) {
	var d models.Dev
	err := db.QueryRow(`SELECT id, nome, equipe, status, obs, criado_em FROM devs WHERE id=$1`, id).
		Scan(&d.ID, &d.Nome, &d.Equipe, &d.Status, &d.Obs, &d.CriadoEm)
	return d, err
}

func CreateDev(db *sql.DB, nome, equipe, status, obs string) error {
	_, err := db.Exec(`INSERT INTO devs (nome, equipe, status, obs) VALUES ($1,$2,$3,$4)`,
		nome, equipe, status, obs)
	return err
}

func UpdateDev(db *sql.DB, id int, nome, equipe, status, obs string) error {
	_, err := db.Exec(`UPDATE devs SET nome=$1, equipe=$2, status=$3, obs=$4 WHERE id=$5`,
		nome, equipe, status, obs, id)
	return err
}

func DeleteDev(db *sql.DB, id int) error {
	_, err := db.Exec(`DELETE FROM devs WHERE id=$1`, id)
	return err
}

func ListEquipes(db *sql.DB) []string {
	rows, _ := db.Query(`SELECT DISTINCT equipe FROM devs WHERE equipe <> '' ORDER BY equipe`)
	defer rows.Close()
	var list []string
	for rows.Next() {
		var s string
		rows.Scan(&s)
		list = append(list, s)
	}
	return list
}

// ============================================================
// Licencas por Dev
// ============================================================

func scanLicenca(rows *sql.Rows) (models.Licenca, error) {
	var l models.Licenca
	var devID sql.NullInt64
	var grupoID sql.NullInt64
	var grupoNome sql.NullString
	err := rows.Scan(&l.ID, &devID, &l.DevNome, &l.Versao, &l.Serial, &l.Tipo, &l.Canal,
		&l.Hostname, &l.EdnLogin, &l.EdnSenha, &l.CadEfetuado, &l.DataCad, &l.Obs, &l.CriadoEm,
		&grupoID, &grupoNome)
	if devID.Valid {
		id := int(devID.Int64)
		l.DevID = &id
	}
	if grupoID.Valid {
		id := int(grupoID.Int64)
		l.GrupoID = &id
	}
	l.GrupoNome = grupoNome.String
	return l, err
}

func ListLicencas(db *sql.DB) ([]models.Licenca, error) {
	rows, err := db.Query(`
		SELECT l.id, l.dev_id, COALESCE(d.nome, ''), l.versao, l.serial, l.tipo, l.canal,
		       l.hostname, l.edn_login, l.edn_senha, l.cad_efetuado, l.data_cad, l.obs, l.criado_em,
		       l.grupo_id, COALESCE(g.nome, '')
		FROM licencas l
		LEFT JOIN devs d ON d.id = l.dev_id
		LEFT JOIN grupos_licenca g ON g.id = l.grupo_id
		ORDER BY COALESCE(g.nome, 'ZZZZZ'), COALESCE(d.nome, 'ZZZZZ'), l.versao
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.Licenca
	for rows.Next() {
		l, _ := scanLicenca(rows)
		list = append(list, l)
	}
	return list, nil
}

func ListLicencasByDev(db *sql.DB, devID int) ([]models.Licenca, error) {
	rows, err := db.Query(`
		SELECT l.id, l.dev_id, COALESCE(d.nome, ''), l.versao, l.serial, l.tipo, l.canal,
		       l.hostname, l.edn_login, l.edn_senha, l.cad_efetuado, l.data_cad, l.obs, l.criado_em,
		       l.grupo_id, COALESCE(g.nome, '')
		FROM licencas l
		LEFT JOIN devs d ON d.id = l.dev_id
		LEFT JOIN grupos_licenca g ON g.id = l.grupo_id
		WHERE l.dev_id = $1
		ORDER BY l.versao
	`, devID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.Licenca
	for rows.Next() {
		l, _ := scanLicenca(rows)
		list = append(list, l)
	}
	return list, nil
}

func CreateLicenca(db *sql.DB, devID *int, grupoID *int, versao, serial, tipo, canal, hostname, ednLogin, ednSenha string, cadEfetuado bool, dataCad *time.Time, obs string) error {
	_, err := db.Exec(`
		INSERT INTO licencas (dev_id, grupo_id, versao, serial, tipo, canal, hostname, edn_login, edn_senha, cad_efetuado, data_cad, obs)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
	`, devID, grupoID, versao, serial, tipo, canal, hostname, ednLogin, ednSenha, cadEfetuado, dataCad, obs)
	return err
}

func UpdateLicenca(db *sql.DB, id int, devID *int, grupoID *int, versao, serial, tipo, canal, hostname, ednLogin, ednSenha string, cadEfetuado bool, dataCad *time.Time, obs string) error {
	_, err := db.Exec(`
		UPDATE licencas SET dev_id=$1, grupo_id=$2, versao=$3, serial=$4, tipo=$5, canal=$6,
		hostname=$7, edn_login=$8, edn_senha=$9, cad_efetuado=$10, data_cad=$11, obs=$12
		WHERE id=$13
	`, devID, grupoID, versao, serial, tipo, canal, hostname, ednLogin, ednSenha, cadEfetuado, dataCad, obs, id)
	return err
}

func VincularDevLicenca(db *sql.DB, id int, devID *int) error {
	// If the licence belongs to a group, cascade to all serials in the group
	var grupoID sql.NullInt64
	db.QueryRow(`SELECT grupo_id FROM licencas WHERE id=$1`, id).Scan(&grupoID)
	if grupoID.Valid {
		_, err := db.Exec(`UPDATE licencas SET dev_id=$1 WHERE grupo_id=$2`, devID, grupoID.Int64)
		return err
	}
	_, err := db.Exec(`UPDATE licencas SET dev_id=$1 WHERE id=$2`, devID, id)
	return err
}

func DeleteLicenca(db *sql.DB, id int) error {
	_, err := db.Exec(`DELETE FROM licencas WHERE id=$1`, id)
	return err
}

func CountLicencasByVersao(db *sql.DB, versao string) int {
	var n int
	db.QueryRow(`SELECT COUNT(*) FROM licencas WHERE versao=$1`, versao).Scan(&n)
	return n
}

// ============================================================
// Auxiliares
// ============================================================

func ListAuxiliaresByTipo(db *sql.DB, tipo string) []string {
	rows, _ := db.Query(`SELECT nome FROM auxiliares WHERE tipo=$1 ORDER BY nome`, tipo)
	defer rows.Close()
	var list []string
	for rows.Next() {
		var s string
		rows.Scan(&s)
		list = append(list, s)
	}
	return list
}

func ListAuxiliarObjs(db *sql.DB, tipo string) []models.Auxiliar {
	rows, _ := db.Query(`SELECT id, tipo, nome, criado_em FROM auxiliares WHERE tipo=$1 ORDER BY nome`, tipo)
	defer rows.Close()
	var list []models.Auxiliar
	for rows.Next() {
		var a models.Auxiliar
		rows.Scan(&a.ID, &a.Tipo, &a.Nome, &a.CriadoEm)
		list = append(list, a)
	}
	return list
}

func CreateAuxiliar(db *sql.DB, tipo, nome string) error {
	_, err := db.Exec(`INSERT INTO auxiliares (tipo, nome) VALUES ($1,$2)`, tipo, nome)
	return err
}

func DeleteAuxiliar(db *sql.DB, id int) error {
	_, err := db.Exec(`DELETE FROM auxiliares WHERE id=$1`, id)
	return err
}

// ============================================================
// Grupos de Licenca
// ============================================================

func ListGrupos(db *sql.DB) ([]models.GrupoLicenca, error) {
	rows, err := db.Query(`SELECT id, nome, obs, criado_em FROM grupos_licenca ORDER BY nome`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.GrupoLicenca
	for rows.Next() {
		var g models.GrupoLicenca
		rows.Scan(&g.ID, &g.Nome, &g.Obs, &g.CriadoEm)
		list = append(list, g)
	}
	return list, nil
}

func CreateGrupo(db *sql.DB, nome, obs string) error {
	_, err := db.Exec(`INSERT INTO grupos_licenca (nome, obs) VALUES ($1,$2)`, nome, obs)
	return err
}

func DeleteGrupo(db *sql.DB, id int) error {
	_, err := db.Exec(`DELETE FROM grupos_licenca WHERE id=$1`, id)
	return err
}

// ============================================================
// Licencas Network
// ============================================================

func ListLicencasNetwork(db *sql.DB) ([]models.LicencaNetwork, error) {
	rows, err := db.Query(`
		SELECT id, sku, descricao, tipo, total_seats, login_name, senha, obs, criado_em
		FROM licencas_network ORDER BY tipo, descricao
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.LicencaNetwork
	for rows.Next() {
		var l models.LicencaNetwork
		rows.Scan(&l.ID, &l.SKU, &l.Descricao, &l.Tipo, &l.TotalSeats, &l.LoginName, &l.Senha, &l.Obs, &l.CriadoEm)
		list = append(list, l)
	}
	return list, nil
}

func CreateLicencaNetwork(db *sql.DB, sku, descricao, tipo string, seats int, loginName, senha, obs string) error {
	_, err := db.Exec(`
		INSERT INTO licencas_network (sku, descricao, tipo, total_seats, login_name, senha, obs)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
	`, sku, descricao, tipo, seats, loginName, senha, obs)
	return err
}

func UpdateLicencaNetwork(db *sql.DB, id int, sku, descricao, tipo string, seats int, loginName, senha, obs string) error {
	_, err := db.Exec(`
		UPDATE licencas_network SET sku=$1, descricao=$2, tipo=$3, total_seats=$4,
		login_name=$5, senha=$6, obs=$7 WHERE id=$8
	`, sku, descricao, tipo, seats, loginName, senha, obs, id)
	return err
}

func DeleteLicencaNetwork(db *sql.DB, id int) error {
	_, err := db.Exec(`DELETE FROM licencas_network WHERE id=$1`, id)
	return err
}

// ============================================================
// Componentes
// ============================================================

func ListComponentes(db *sql.DB) ([]models.Componente, error) {
	rows, err := db.Query(`
		SELECT id, nome, versao, ferramenta, informacoes, uso, licenciamento,
		       tipo_licenca, site, status, obs, diretorio, serial, usuario, senha, criado_em
		FROM componentes ORDER BY nome
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.Componente
	for rows.Next() {
		var c models.Componente
		rows.Scan(&c.ID, &c.Nome, &c.Versao, &c.Ferramenta, &c.Informacoes, &c.Uso,
			&c.Licenciamento, &c.TipoLicenca, &c.Site, &c.Status, &c.Obs, &c.Diretorio,
			&c.Serial, &c.Usuario, &c.Senha, &c.CriadoEm)
		list = append(list, c)
	}
	return list, nil
}

func CreateComponente(db *sql.DB, nome, versao, ferramenta, informacoes, uso, licenciamento, site, obs, serial, usuario, senha string) error {
	_, err := db.Exec(`
		INSERT INTO componentes (nome, versao, ferramenta, informacoes, uso, licenciamento, site, obs, serial, usuario, senha)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	`, nome, versao, ferramenta, informacoes, uso, licenciamento, site, obs, serial, usuario, senha)
	return err
}

func UpdateComponente(db *sql.DB, id int, nome, versao, ferramenta, informacoes, uso, licenciamento, site, obs, serial, usuario, senha string) error {
	_, err := db.Exec(`
		UPDATE componentes SET nome=$1, versao=$2, ferramenta=$3, informacoes=$4, uso=$5,
		licenciamento=$6, site=$7, obs=$8, serial=$9, usuario=$10, senha=$11
		WHERE id=$12
	`, nome, versao, ferramenta, informacoes, uso, licenciamento, site, obs, serial, usuario, senha, id)
	return err
}

func DeleteComponente(db *sql.DB, id int) error {
	_, err := db.Exec(`DELETE FROM componentes WHERE id=$1`, id)
	return err
}

func ListFerramentas(db *sql.DB) []string {
	rows, _ := db.Query(`SELECT DISTINCT ferramenta FROM componentes WHERE ferramenta <> '' ORDER BY ferramenta`)
	defer rows.Close()
	var list []string
	for rows.Next() {
		var s string
		rows.Scan(&s)
		list = append(list, s)
	}
	return list
}

// ============================================================
// Usuarios
// ============================================================

func ListUsuarios(db *sql.DB) ([]models.Usuario, error) {
	rows, err := db.Query(`SELECT id, nome, email, admin, ativo, criado_em FROM usuarios ORDER BY nome`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.Usuario
	for rows.Next() {
		var u models.Usuario
		rows.Scan(&u.ID, &u.Nome, &u.Email, &u.Admin, &u.Ativo, &u.CriadoEm)
		list = append(list, u)
	}
	return list, nil
}

func CreateUsuario(db *sql.DB, nome, email, senha string, admin bool) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = db.Exec(`INSERT INTO usuarios (nome, email, senha_hash, admin) VALUES ($1,$2,$3,$4)`,
		nome, email, string(hash), admin)
	return err
}

func ToggleUsuarioAtivo(db *sql.DB, id int) error {
	_, err := db.Exec(`UPDATE usuarios SET ativo = NOT ativo WHERE id=$1`, id)
	return err
}

func ToggleUsuarioAdmin(db *sql.DB, id int) error {
	_, err := db.Exec(`UPDATE usuarios SET admin = NOT admin WHERE id=$1`, id)
	return err
}

func ResetSenha(db *sql.DB, id int, novaSenha string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(novaSenha), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = db.Exec(`UPDATE usuarios SET senha_hash=$1 WHERE id=$2`, string(hash), id)
	return err
}

func Authenticate(db *sql.DB, email, senha string) (*models.Usuario, error) {
	var u models.Usuario
	var hash string
	err := db.QueryRow(`SELECT id, nome, email, admin, ativo, senha_hash FROM usuarios WHERE email=$1 AND ativo=true`, email).
		Scan(&u.ID, &u.Nome, &u.Email, &u.Admin, &u.Ativo, &hash)
	if err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(senha)); err != nil {
		return nil, err
	}
	return &u, nil
}

// ============================================================
// Dashboard stats
// ============================================================

func DashboardStats(db *sql.DB) (totalDevs, devsAtivos, devsLivres, totalLic, licEmUso, licLivres, xe3, d12, ib, compTotal, compPagos, compFree int) {
	db.QueryRow(`SELECT COUNT(*) FROM devs`).Scan(&totalDevs)
	db.QueryRow(`SELECT COUNT(*) FROM devs WHERE status='ativo'`).Scan(&devsAtivos)
	db.QueryRow(`SELECT COUNT(*) FROM devs WHERE status='livre'`).Scan(&devsLivres)
	db.QueryRow(`SELECT COUNT(*) FROM licencas`).Scan(&totalLic)
	db.QueryRow(`SELECT COUNT(*) FROM licencas WHERE dev_id IS NOT NULL`).Scan(&licEmUso)
	db.QueryRow(`SELECT COUNT(*) FROM licencas WHERE dev_id IS NULL`).Scan(&licLivres)
	db.QueryRow(`SELECT COUNT(*) FROM licencas WHERE versao='XE3'`).Scan(&xe3)
	db.QueryRow(`SELECT COUNT(*) FROM licencas WHERE versao='D12'`).Scan(&d12)
	db.QueryRow(`SELECT COUNT(*) FROM licencas WHERE versao='Interbase'`).Scan(&ib)
	db.QueryRow(`SELECT COUNT(*) FROM componentes`).Scan(&compTotal)
	db.QueryRow(`SELECT COUNT(*) FROM componentes WHERE licenciamento='Pago'`).Scan(&compPagos)
	db.QueryRow(`SELECT COUNT(*) FROM componentes WHERE licenciamento='Free'`).Scan(&compFree)
	return
}

// ============================================================
// Audit Log
// ============================================================

func RegistrarAuditoria(db *sql.DB, entidade string, entidadeID int, acao, usuario, detalhes string) {
	db.Exec(`INSERT INTO audit_log (entidade, entidade_id, acao, usuario, detalhes) VALUES ($1,$2,$3,$4,$5)`,
		entidade, entidadeID, acao, usuario, detalhes)
}

func ListAuditLogs(db *sql.DB, entidade string, entidadeID, limit, offset int) ([]models.AuditLog, int) {
	var total int
	var rows *sql.Rows
	var err error

	if entidade != "" && entidadeID > 0 {
		db.QueryRow(`SELECT COUNT(*) FROM audit_log WHERE entidade=$1 AND entidade_id=$2`, entidade, entidadeID).Scan(&total)
		rows, err = db.Query(`
			SELECT id, entidade, entidade_id, acao, usuario, detalhes, criado_em
			FROM audit_log WHERE entidade=$1 AND entidade_id=$2
			ORDER BY criado_em DESC LIMIT $3 OFFSET $4`,
			entidade, entidadeID, limit, offset)
	} else if entidade != "" {
		db.QueryRow(`SELECT COUNT(*) FROM audit_log WHERE entidade=$1`, entidade).Scan(&total)
		rows, err = db.Query(`
			SELECT id, entidade, entidade_id, acao, usuario, detalhes, criado_em
			FROM audit_log WHERE entidade=$1
			ORDER BY criado_em DESC LIMIT $2 OFFSET $3`,
			entidade, limit, offset)
	} else {
		db.QueryRow(`SELECT COUNT(*) FROM audit_log`).Scan(&total)
		rows, err = db.Query(`
			SELECT id, entidade, entidade_id, acao, usuario, detalhes, criado_em
			FROM audit_log ORDER BY criado_em DESC LIMIT $1 OFFSET $2`,
			limit, offset)
	}
	if err != nil {
		return nil, 0
	}
	defer rows.Close()
	var logs []models.AuditLog
	for rows.Next() {
		var l models.AuditLog
		rows.Scan(&l.ID, &l.Entidade, &l.EntidadeID, &l.Acao, &l.Usuario, &l.Detalhes, &l.CriadoEm)
		logs = append(logs, l)
	}
	return logs, total
}

func GetLicencaSerial(db *sql.DB, id int) string {
	var s string
	db.QueryRow(`SELECT serial FROM licencas WHERE id=$1`, id).Scan(&s)
	return s
}

func GetLicenca(db *sql.DB, id int) (models.Licenca, error) {
	var l models.Licenca
	var devID sql.NullInt64
	var grupoID sql.NullInt64
	var grupoNome sql.NullString
	err := db.QueryRow(`
		SELECT l.id, l.dev_id, COALESCE(d.nome, ''), l.versao, l.serial, l.tipo, l.canal,
		       l.hostname, l.edn_login, l.edn_senha, l.cad_efetuado, l.data_cad, l.obs, l.criado_em,
		       l.grupo_id, COALESCE(g.nome, '')
		FROM licencas l
		LEFT JOIN devs d ON d.id = l.dev_id
		LEFT JOIN grupos_licenca g ON g.id = l.grupo_id
		WHERE l.id = $1
	`, id).Scan(&l.ID, &devID, &l.DevNome, &l.Versao, &l.Serial, &l.Tipo, &l.Canal,
		&l.Hostname, &l.EdnLogin, &l.EdnSenha, &l.CadEfetuado, &l.DataCad, &l.Obs, &l.CriadoEm,
		&grupoID, &grupoNome)
	if devID.Valid {
		id := int(devID.Int64)
		l.DevID = &id
	}
	if grupoID.Valid {
		id := int(grupoID.Int64)
		l.GrupoID = &id
	}
	l.GrupoNome = grupoNome.String
	return l, err
}

func GetComponente(db *sql.DB, id int) (models.Componente, error) {
	var c models.Componente
	err := db.QueryRow(`
		SELECT id, nome, versao, ferramenta, informacoes, uso, licenciamento,
		       tipo_licenca, site, status, obs, diretorio, serial, usuario, senha, criado_em
		FROM componentes WHERE id=$1
	`, id).Scan(&c.ID, &c.Nome, &c.Versao, &c.Ferramenta, &c.Informacoes, &c.Uso,
		&c.Licenciamento, &c.TipoLicenca, &c.Site, &c.Status, &c.Obs, &c.Diretorio,
		&c.Serial, &c.Usuario, &c.Senha, &c.CriadoEm)
	return c, err
}

func FindDevIDByNome(db *sql.DB, nome string) *int {
	if nome == "" {
		return nil
	}
	var id int
	err := db.QueryRow(`SELECT id FROM devs WHERE nome=$1 LIMIT 1`, nome).Scan(&id)
	if err != nil {
		return nil
	}
	return &id
}

func FindGrupoIDByNome(db *sql.DB, nome string) *int {
	if nome == "" {
		return nil
	}
	var id int
	err := db.QueryRow(`SELECT id FROM grupos_licenca WHERE nome=$1 LIMIT 1`, nome).Scan(&id)
	if err != nil {
		return nil
	}
	return &id
}
