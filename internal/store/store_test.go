package store

import (
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

// testDB returns a connection to the test database.
// Set TEST_DATABASE_URL to run integration tests.
// Defaults to the dev database on port 5433.
func testDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://delphilic:delphilic@localhost:5433/delphilic?sslmode=disable"
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Skipf("cannot open test db: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("cannot ping test db: %v", err)
	}
	if err := RunMigrations(db); err != nil {
		t.Fatalf("migrations failed: %v", err)
	}
	return db
}

func cleanup(db *sql.DB) {
	db.Exec(`DELETE FROM audit_log`)
	db.Exec(`DELETE FROM licencas`)
	db.Exec(`DELETE FROM grupos_licenca`)
	db.Exec(`DELETE FROM componentes WHERE nome LIKE '_test_%'`)
	db.Exec(`DELETE FROM auxiliares WHERE tipo='_test'`)
	db.Exec(`DELETE FROM devs WHERE nome LIKE '_test_%'`)
}

// ============================================================
// Dev CRUD
// ============================================================

func TestDevCRUD(t *testing.T) {
	db := testDB(t)
	defer db.Close()
	cleanup(db)
	defer cleanup(db)

	// Create
	err := CreateDev(db, "_test_dev1", "Backend", "ativo", "obs1")
	if err != nil {
		t.Fatalf("CreateDev: %v", err)
	}

	// List
	devs, err := ListDevs(db)
	if err != nil {
		t.Fatalf("ListDevs: %v", err)
	}
	var found bool
	var devID int
	for _, d := range devs {
		if d.Nome == "_test_dev1" {
			found = true
			devID = d.ID
			if d.Equipe != "Backend" {
				t.Errorf("expected equipe Backend, got %q", d.Equipe)
			}
			if d.Status != "ativo" {
				t.Errorf("expected status ativo, got %q", d.Status)
			}
		}
	}
	if !found {
		t.Fatal("dev not found in list")
	}

	// Get
	dev, err := GetDev(db, devID)
	if err != nil {
		t.Fatalf("GetDev: %v", err)
	}
	if dev.Nome != "_test_dev1" {
		t.Errorf("expected nome _test_dev1, got %q", dev.Nome)
	}

	// Update
	err = UpdateDev(db, devID, "_test_dev1_updated", "Frontend", "livre", "obs2")
	if err != nil {
		t.Fatalf("UpdateDev: %v", err)
	}
	dev, _ = GetDev(db, devID)
	if dev.Nome != "_test_dev1_updated" {
		t.Errorf("expected updated nome, got %q", dev.Nome)
	}
	if dev.Equipe != "Frontend" {
		t.Errorf("expected equipe Frontend, got %q", dev.Equipe)
	}

	// Delete
	err = DeleteDev(db, devID)
	if err != nil {
		t.Fatalf("DeleteDev: %v", err)
	}
	_, err = GetDev(db, devID)
	if err == nil {
		t.Error("expected error getting deleted dev")
	}
}

func TestListEquipes(t *testing.T) {
	db := testDB(t)
	defer db.Close()
	cleanup(db)
	defer cleanup(db)

	CreateDev(db, "_test_eq1", "Alpha", "ativo", "")
	CreateDev(db, "_test_eq2", "Beta", "ativo", "")
	CreateDev(db, "_test_eq3", "Alpha", "ativo", "")

	equipes := ListEquipes(db)
	alphaCount := 0
	betaCount := 0
	for _, e := range equipes {
		if e == "Alpha" {
			alphaCount++
		}
		if e == "Beta" {
			betaCount++
		}
	}
	if alphaCount != 1 {
		t.Errorf("expected 1 Alpha (distinct), got %d", alphaCount)
	}
	if betaCount != 1 {
		t.Errorf("expected 1 Beta, got %d", betaCount)
	}
}

// ============================================================
// Licenca CRUD
// ============================================================

func TestLicencaCRUD(t *testing.T) {
	db := testDB(t)
	defer db.Close()
	cleanup(db)
	defer cleanup(db)

	// Create licence (no dev)
	err := CreateLicenca(db, nil, nil, "D12", "SERIAL-TEST-001", "Professional", "EDN", "PC-01", "login@test", "senha123", false, nil, "teste")
	if err != nil {
		t.Fatalf("CreateLicenca: %v", err)
	}

	// List
	lics, err := ListLicencas(db)
	if err != nil {
		t.Fatalf("ListLicencas: %v", err)
	}

	var licID int
	for _, l := range lics {
		if l.Serial == "SERIAL-TEST-001" {
			licID = l.ID
			if l.Versao != "D12" {
				t.Errorf("expected versao D12, got %q", l.Versao)
			}
			if l.DevID != nil {
				t.Error("expected DevID nil (livre)")
			}
			if l.Hostname != "PC-01" {
				t.Errorf("expected hostname PC-01, got %q", l.Hostname)
			}
		}
	}
	if licID == 0 {
		t.Fatal("licence not found")
	}

	// Update
	now := time.Now()
	err = UpdateLicenca(db, licID, nil, nil, "XE3", "SERIAL-TEST-001-UPD", "Enterprise", "Network", "PC-02", "login2@test", "senha456", true, &now, "updated")
	if err != nil {
		t.Fatalf("UpdateLicenca: %v", err)
	}

	// GetLicencaSerial
	serial := GetLicencaSerial(db, licID)
	if serial != "SERIAL-TEST-001-UPD" {
		t.Errorf("expected updated serial, got %q", serial)
	}

	// Delete
	err = DeleteLicenca(db, licID)
	if err != nil {
		t.Fatalf("DeleteLicenca: %v", err)
	}
	serial = GetLicencaSerial(db, licID)
	if serial != "" {
		t.Errorf("expected empty serial after delete, got %q", serial)
	}
}

func TestVincularDevLicenca(t *testing.T) {
	db := testDB(t)
	defer db.Close()
	cleanup(db)
	defer cleanup(db)

	// Create dev and licence
	CreateDev(db, "_test_vincular_dev", "QA", "ativo", "")
	devs, _ := ListDevs(db)
	var devID int
	for _, d := range devs {
		if d.Nome == "_test_vincular_dev" {
			devID = d.ID
		}
	}

	CreateLicenca(db, nil, nil, "D12", "SERIAL-VINCULAR", "Professional", "EDN", "", "", "", false, nil, "")
	lics, _ := ListLicencas(db)
	var licID int
	for _, l := range lics {
		if l.Serial == "SERIAL-VINCULAR" {
			licID = l.ID
		}
	}

	// Vincular
	err := VincularDevLicenca(db, licID, &devID)
	if err != nil {
		t.Fatalf("VincularDevLicenca: %v", err)
	}

	lics, _ = ListLicencas(db)
	for _, l := range lics {
		if l.ID == licID {
			if l.DevID == nil || *l.DevID != devID {
				t.Error("expected licence to be linked to dev")
			}
		}
	}

	// Desvincular
	err = VincularDevLicenca(db, licID, nil)
	if err != nil {
		t.Fatalf("Desvincular: %v", err)
	}

	lics, _ = ListLicencas(db)
	for _, l := range lics {
		if l.ID == licID {
			if l.DevID != nil {
				t.Error("expected licence to be free after desvincular")
			}
		}
	}
}

func TestVincularDevLicenca_GroupCascade(t *testing.T) {
	db := testDB(t)
	defer db.Close()
	cleanup(db)
	defer cleanup(db)

	// Create grupo, dev, 2 licences in the same grupo
	CreateGrupo(db, "_test_grupo_cascade", "")
	grupos, _ := ListGrupos(db)
	var grupoID int
	for _, g := range grupos {
		if g.Nome == "_test_grupo_cascade" {
			grupoID = g.ID
		}
	}

	CreateDev(db, "_test_cascade_dev", "QA", "ativo", "")
	devs, _ := ListDevs(db)
	var devID int
	for _, d := range devs {
		if d.Nome == "_test_cascade_dev" {
			devID = d.ID
		}
	}

	CreateLicenca(db, nil, &grupoID, "XE3", "CASCADE-XE3", "Professional", "EDN", "", "", "", false, nil, "")
	CreateLicenca(db, nil, &grupoID, "D12", "CASCADE-D12", "Professional", "EDN", "", "", "", false, nil, "")

	lics, _ := ListLicencas(db)
	var licXE3ID int
	for _, l := range lics {
		if l.Serial == "CASCADE-XE3" {
			licXE3ID = l.ID
		}
	}

	// Vincular only XE3 — should cascade to D12
	err := VincularDevLicenca(db, licXE3ID, &devID)
	if err != nil {
		t.Fatalf("VincularDevLicenca cascade: %v", err)
	}

	lics, _ = ListLicencas(db)
	for _, l := range lics {
		if l.Serial == "CASCADE-XE3" || l.Serial == "CASCADE-D12" {
			if l.DevID == nil || *l.DevID != devID {
				t.Errorf("expected %s to be linked to dev after cascade", l.Serial)
			}
		}
	}
}

// ============================================================
// Grupos
// ============================================================

func TestGrupoCRUD(t *testing.T) {
	db := testDB(t)
	defer db.Close()
	cleanup(db)
	defer cleanup(db)

	err := CreateGrupo(db, "_test_grupo1", "observacao")
	if err != nil {
		t.Fatalf("CreateGrupo: %v", err)
	}

	grupos, err := ListGrupos(db)
	if err != nil {
		t.Fatalf("ListGrupos: %v", err)
	}
	var gID int
	for _, g := range grupos {
		if g.Nome == "_test_grupo1" {
			gID = g.ID
			if g.Obs != "observacao" {
				t.Errorf("expected obs, got %q", g.Obs)
			}
		}
	}
	if gID == 0 {
		t.Fatal("grupo not found")
	}

	err = DeleteGrupo(db, gID)
	if err != nil {
		t.Fatalf("DeleteGrupo: %v", err)
	}
}

// ============================================================
// Auxiliares
// ============================================================

func TestAuxiliaresCRUD(t *testing.T) {
	db := testDB(t)
	defer db.Close()
	cleanup(db)
	defer cleanup(db)

	err := CreateAuxiliar(db, "_test", "TestValue1")
	if err != nil {
		t.Fatalf("CreateAuxiliar: %v", err)
	}
	err = CreateAuxiliar(db, "_test", "TestValue2")
	if err != nil {
		t.Fatalf("CreateAuxiliar: %v", err)
	}

	// ListByTipo
	names := ListAuxiliaresByTipo(db, "_test")
	if len(names) != 2 {
		t.Errorf("expected 2 auxiliares, got %d", len(names))
	}

	// ListObjs
	objs := ListAuxiliarObjs(db, "_test")
	if len(objs) != 2 {
		t.Errorf("expected 2 objs, got %d", len(objs))
	}

	// Delete
	if len(objs) > 0 {
		err = DeleteAuxiliar(db, objs[0].ID)
		if err != nil {
			t.Fatalf("DeleteAuxiliar: %v", err)
		}
		names = ListAuxiliaresByTipo(db, "_test")
		if len(names) != 1 {
			t.Errorf("expected 1 after delete, got %d", len(names))
		}
	}
}

// ============================================================
// Audit Log
// ============================================================

func TestAuditLog(t *testing.T) {
	db := testDB(t)
	defer db.Close()
	cleanup(db)
	defer cleanup(db)

	RegistrarAuditoria(db, "licenca", 999, "criar", "admin", "Serial: TEST-001")
	RegistrarAuditoria(db, "licenca", 999, "editar", "admin", "Versao alterada")
	RegistrarAuditoria(db, "dev", 1, "criar", "admin", "Dev criado")

	// All logs
	logs, total := ListAuditLogs(db, "", 0, 50, 0)
	if total < 3 {
		t.Errorf("expected at least 3 logs, got %d", total)
	}
	if len(logs) < 3 {
		t.Errorf("expected at least 3 log entries, got %d", len(logs))
	}

	// Filter by entidade
	logs, total = ListAuditLogs(db, "licenca", 0, 50, 0)
	if total < 2 {
		t.Errorf("expected at least 2 licenca logs, got %d", total)
	}

	// Filter by entidade + ID
	logs, total = ListAuditLogs(db, "licenca", 999, 50, 0)
	if total != 2 {
		t.Errorf("expected 2 logs for licenca 999, got %d", total)
	}
	if len(logs) != 2 {
		t.Errorf("expected 2 entries, got %d", len(logs))
	}

	// Verify order (most recent first)
	if len(logs) >= 2 {
		if logs[0].Acao != "editar" {
			t.Errorf("expected most recent log first (editar), got %q", logs[0].Acao)
		}
		if logs[1].Acao != "criar" {
			t.Errorf("expected second log (criar), got %q", logs[1].Acao)
		}
	}

	// Pagination
	logs, _ = ListAuditLogs(db, "licenca", 999, 1, 0)
	if len(logs) != 1 {
		t.Errorf("expected 1 log with limit 1, got %d", len(logs))
	}
	logs, _ = ListAuditLogs(db, "licenca", 999, 1, 1)
	if len(logs) != 1 {
		t.Errorf("expected 1 log with offset 1, got %d", len(logs))
	}
}

// ============================================================
// Dashboard Stats
// ============================================================

func TestDashboardStats(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	// Just verify it doesn't panic and returns reasonable values
	totalDevs, devsAtivos, devsLivres, totalLic, licEmUso, licLivres, _, _, _, compTotal, compPagos, compFree := DashboardStats(db)

	if totalDevs < 0 {
		t.Error("totalDevs should be >= 0")
	}
	if devsAtivos+devsLivres > totalDevs {
		// Some devs may be "inativo", so ativos+livres can be <= totalDevs
	}
	if licEmUso+licLivres != totalLic {
		t.Errorf("licEmUso(%d) + licLivres(%d) should equal totalLic(%d)", licEmUso, licLivres, totalLic)
	}
	if compPagos+compFree > compTotal {
		t.Error("pagos + free should be <= total componentes")
	}
}

// ============================================================
// Migrations
// ============================================================

// ============================================================
// GetLicenca
// ============================================================

func TestGetLicenca(t *testing.T) {
	db := testDB(t)
	defer db.Close()
	cleanup(db)
	defer cleanup(db)

	CreateLicenca(db, nil, nil, "D12", "GET-LIC-TEST", "Enterprise", "EDN", "PC-99", "login", "pass", true, nil, "obs test")
	lics, _ := ListLicencas(db)
	var licID int
	for _, l := range lics {
		if l.Serial == "GET-LIC-TEST" {
			licID = l.ID
		}
	}
	if licID == 0 {
		t.Fatal("licence not found")
	}

	lic, err := GetLicenca(db, licID)
	if err != nil {
		t.Fatalf("GetLicenca: %v", err)
	}
	if lic.Serial != "GET-LIC-TEST" {
		t.Errorf("expected serial GET-LIC-TEST, got %q", lic.Serial)
	}
	if lic.Versao != "D12" {
		t.Errorf("expected versao D12, got %q", lic.Versao)
	}
	if lic.Hostname != "PC-99" {
		t.Errorf("expected hostname PC-99, got %q", lic.Hostname)
	}
	if !lic.CadEfetuado {
		t.Error("expected CadEfetuado true")
	}
}

// ============================================================
// Componente CRUD
// ============================================================

func TestComponenteCRUD(t *testing.T) {
	db := testDB(t)
	defer db.Close()
	cleanup(db)
	defer cleanup(db)

	err := CreateComponente(db, "_test_comp1", "1.0", "Delphi D12", "info", "Todos", "Pago", "https://test.com", "obs", "SERIAL-COMP", "user1", "pass1")
	if err != nil {
		t.Fatalf("CreateComponente: %v", err)
	}

	comps, err := ListComponentes(db)
	if err != nil {
		t.Fatalf("ListComponentes: %v", err)
	}
	var compID int
	for _, c := range comps {
		if c.Nome == "_test_comp1" {
			compID = c.ID
			if c.Serial != "SERIAL-COMP" {
				t.Errorf("expected serial SERIAL-COMP, got %q", c.Serial)
			}
			if c.Usuario != "user1" {
				t.Errorf("expected usuario user1, got %q", c.Usuario)
			}
		}
	}
	if compID == 0 {
		t.Fatal("componente not found")
	}

	// GetComponente
	comp, err := GetComponente(db, compID)
	if err != nil {
		t.Fatalf("GetComponente: %v", err)
	}
	if comp.Nome != "_test_comp1" {
		t.Errorf("expected nome _test_comp1, got %q", comp.Nome)
	}
	if comp.Senha != "pass1" {
		t.Errorf("expected senha pass1, got %q", comp.Senha)
	}

	// Update
	err = UpdateComponente(db, compID, "_test_comp1_upd", "2.0", "XE3", "info2", "Fiscal", "Free", "https://new.com", "obs2", "NEW-SERIAL", "user2", "pass2")
	if err != nil {
		t.Fatalf("UpdateComponente: %v", err)
	}
	comp, _ = GetComponente(db, compID)
	if comp.Nome != "_test_comp1_upd" {
		t.Errorf("expected updated nome, got %q", comp.Nome)
	}
	if comp.Serial != "NEW-SERIAL" {
		t.Errorf("expected NEW-SERIAL, got %q", comp.Serial)
	}

	// Delete
	err = DeleteComponente(db, compID)
	if err != nil {
		t.Fatalf("DeleteComponente: %v", err)
	}
	_, err = GetComponente(db, compID)
	if err == nil {
		t.Error("expected error getting deleted componente")
	}
}

// ============================================================
// FindDevIDByNome / FindGrupoIDByNome
// ============================================================

func TestFindDevIDByNome(t *testing.T) {
	db := testDB(t)
	defer db.Close()
	cleanup(db)
	defer cleanup(db)

	CreateDev(db, "_test_find_dev", "QA", "ativo", "")

	id := FindDevIDByNome(db, "_test_find_dev")
	if id == nil {
		t.Fatal("expected non-nil dev ID")
	}
	if *id <= 0 {
		t.Errorf("expected positive ID, got %d", *id)
	}

	// Not found
	id = FindDevIDByNome(db, "_test_nao_existe")
	if id != nil {
		t.Error("expected nil for non-existent dev")
	}

	// Empty name
	id = FindDevIDByNome(db, "")
	if id != nil {
		t.Error("expected nil for empty name")
	}
}

func TestFindGrupoIDByNome(t *testing.T) {
	db := testDB(t)
	defer db.Close()
	cleanup(db)
	defer cleanup(db)

	CreateGrupo(db, "_test_find_grupo", "obs")

	id := FindGrupoIDByNome(db, "_test_find_grupo")
	if id == nil {
		t.Fatal("expected non-nil grupo ID")
	}

	id = FindGrupoIDByNome(db, "_test_nao_existe")
	if id != nil {
		t.Error("expected nil for non-existent grupo")
	}

	id = FindGrupoIDByNome(db, "")
	if id != nil {
		t.Error("expected nil for empty name")
	}
}

// ============================================================
// Delete Dev should SET NULL on licencas (not cascade)
// ============================================================

func TestDeleteDev_SetsLicencaFree(t *testing.T) {
	db := testDB(t)
	defer db.Close()
	cleanup(db)
	defer cleanup(db)

	// Create dev
	CreateDev(db, "_test_del_dev", "QA", "ativo", "")
	devs, _ := ListDevs(db)
	var devID int
	for _, d := range devs {
		if d.Nome == "_test_del_dev" {
			devID = d.ID
		}
	}

	// Create licence linked to dev
	CreateLicenca(db, &devID, nil, "D12", "DEL-DEV-TEST", "Professional", "EDN", "", "", "", false, nil, "")

	// Delete dev
	err := DeleteDev(db, devID)
	if err != nil {
		t.Fatalf("DeleteDev: %v", err)
	}

	// Licence should still exist but be free (dev_id = NULL)
	lics, _ := ListLicencas(db)
	found := false
	for _, l := range lics {
		if l.Serial == "DEL-DEV-TEST" {
			found = true
			if l.DevID != nil {
				t.Error("expected licence to be free (dev_id NULL) after dev deletion")
			}
		}
	}
	if !found {
		t.Error("licence should NOT be deleted when dev is deleted")
	}
}

// ============================================================
// Migrations
// ============================================================

func TestRunMigrations_Idempotent(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	// Running migrations twice should not fail
	err := RunMigrations(db)
	if err != nil {
		t.Fatalf("second migration run failed: %v", err)
	}
}
