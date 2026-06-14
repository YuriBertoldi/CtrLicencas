package models

import (
	"testing"
	"time"
)

func TestLicenca_NilDevID(t *testing.T) {
	l := Licenca{ID: 1, Serial: "ABC"}
	if l.DevID != nil {
		t.Error("expected DevID to be nil")
	}
	if l.GrupoID != nil {
		t.Error("expected GrupoID to be nil")
	}
}

func TestLicenca_WithDevID(t *testing.T) {
	id := 5
	l := Licenca{ID: 1, DevID: &id, DevNome: "Joao"}
	if l.DevID == nil || *l.DevID != 5 {
		t.Error("expected DevID to be 5")
	}
}

func TestAuditLog_Fields(t *testing.T) {
	a := AuditLog{
		ID:         1,
		Entidade:   "licenca",
		EntidadeID: 42,
		Acao:       "criar",
		Usuario:    "admin",
		Detalhes:   "Serial criado",
		CriadoEm:  time.Now(),
	}
	if a.Entidade != "licenca" {
		t.Errorf("expected entidade 'licenca', got %q", a.Entidade)
	}
	if a.Acao != "criar" {
		t.Errorf("expected acao 'criar', got %q", a.Acao)
	}
}

func TestBasePage_Fields(t *testing.T) {
	u := &Usuario{ID: 1, Nome: "Admin", Admin: true}
	bp := BasePage{CurrentUser: u, Active: "dashboard", Title: "Dashboard"}
	if bp.CurrentUser.Nome != "Admin" {
		t.Error("expected user name Admin")
	}
	if bp.Active != "dashboard" {
		t.Error("expected active dashboard")
	}
}

func TestDashboardData_NewFields(t *testing.T) {
	d := DashboardData{
		TotalLicencas:  10,
		LicencasEmUso:  7,
		LicencasLivres: 3,
	}
	if d.TotalLicencas != d.LicencasEmUso+d.LicencasLivres {
		t.Error("expected total = em uso + livres")
	}
}

func TestAuditoriaPage_Pagination(t *testing.T) {
	p := AuditoriaPage{
		Total:     120,
		Pagina:    2,
		PorPagina: 50,
	}
	totalPages := (p.Total + p.PorPagina - 1) / p.PorPagina
	if totalPages != 3 {
		t.Errorf("expected 3 pages, got %d", totalPages)
	}
}

func TestGrupoLicenca_Fields(t *testing.T) {
	g := GrupoLicenca{ID: 1, Nome: "Lic Fisica 01", Obs: "teste"}
	if g.Nome != "Lic Fisica 01" {
		t.Errorf("expected nome 'Lic Fisica 01', got %q", g.Nome)
	}
}

func TestAuxiliar_Types(t *testing.T) {
	tipos := []string{"equipe", "versao", "tipo_lic", "canal"}
	for _, tipo := range tipos {
		a := Auxiliar{Tipo: tipo, Nome: "test"}
		if a.Tipo != tipo {
			t.Errorf("expected tipo %q, got %q", tipo, a.Tipo)
		}
	}
}

func TestUsuario_AdminFlag(t *testing.T) {
	u := Usuario{ID: 1, Admin: true, Ativo: true}
	if !u.Admin {
		t.Error("expected admin true")
	}
	if !u.Ativo {
		t.Error("expected ativo true")
	}
}
