package models

import "time"

// --- Entidades ---

type Dev struct {
	ID       int
	Nome     string
	Equipe   string
	Status   string // ativo | livre | inativo
	Obs      string
	CriadoEm time.Time
}

type Licenca struct {
	ID          int
	DevID       *int   // nil = livre
	DevNome     string // vazio = livre
	GrupoID     *int
	GrupoNome   string
	Versao      string // XE3 | D12 | Interbase | HTML5Builder
	Serial      string
	Tipo        string // Professional | Enterprise
	Canal       string // EDN | Network
	Hostname    string
	EdnLogin    string
	EdnSenha    string
	CadEfetuado bool
	DataCad     *time.Time
	Obs         string
	CriadoEm   time.Time
}

type GrupoLicenca struct {
	ID       int
	Nome     string
	Obs      string
	CriadoEm time.Time
}

type Auxiliar struct {
	ID       int
	Tipo     string // equipe | versao | tipo_lic | canal
	Nome     string
	CriadoEm time.Time
}

type LicencaNetwork struct {
	ID         int
	SKU        string
	Descricao  string
	Tipo       string // Professional | Enterprise
	TotalSeats int
	LoginName  string
	Senha      string
	Obs        string
	CriadoEm  time.Time
}

type Componente struct {
	ID           int
	Nome         string
	Versao       string
	Ferramenta   string
	Informacoes  string
	Uso          string
	Licenciamento string
	TipoLicenca  string
	Site         string
	Status       string
	Obs          string
	Diretorio    string
	Serial       string
	Usuario      string
	Senha        string
	CriadoEm    time.Time
}

type Usuario struct {
	ID       int
	Nome     string
	Email    string
	Admin    bool
	Ativo    bool
	CriadoEm time.Time
}

type AuditLog struct {
	ID         int
	Entidade   string // licenca, dev, componente, grupo
	EntidadeID int
	Acao       string // criar, editar, excluir, vincular, desvincular
	Usuario    string
	Detalhes   string
	CriadoEm   time.Time
}

// --- Page Data ---

type BasePage struct {
	CurrentUser *Usuario
	Active      string
	Title       string
}

type DashboardData struct {
	BasePage
	TotalDevs        int
	DevsAtivos       int
	DevsLivres       int
	TotalLicencas    int
	LicencasEmUso    int
	LicencasLivres   int
	TotalLicencasXE3 int
	TotalLicencasD12 int
	TotalLicencasIB  int
	TotalComponentes int
	ComponentesPagos int
	ComponentesFree  int
	RecentesLicencas []Licenca
}

type DevsPage struct {
	BasePage
	Devs   []Dev
	Equipes []string
	Erro   string
	Sucesso string
}

type LicencasPage struct {
	BasePage
	Licencas []Licenca
	Devs     []Dev
	Grupos   []GrupoLicenca
	Versoes  []string
	Tipos    []string
	Canais   []string
	DevID    int
	Erro     string
	Sucesso  string
}

type AuxiliaresPage struct {
	BasePage
	Equipes  []Auxiliar
	Versoes  []Auxiliar
	TiposLic []Auxiliar
	Canais   []Auxiliar
	Erro     string
}

type NetworkPage struct {
	BasePage
	Licencas []LicencaNetwork
	Erro     string
	Sucesso  string
}

type ComponentesPage struct {
	BasePage
	Componentes   []Componente
	Ferramentas   []string
	Licenciamentos []string
	Erro          string
	Sucesso       string
}

type UsuariosPage struct {
	BasePage
	Usuarios []Usuario
	Erro     string
	Sucesso  string
}

type AuditoriaPage struct {
	BasePage
	Logs       []AuditLog
	Entidade   string
	EntidadeID int
	Total      int
	Pagina     int
	PorPagina  int
}

type LoginPage struct {
	Title string
	Erro  string
}
