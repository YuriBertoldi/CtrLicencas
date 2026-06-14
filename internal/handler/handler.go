package handler

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"delphilic/internal/auth"
	"delphilic/internal/models"
	"delphilic/internal/store"
	"golang.org/x/crypto/bcrypt"
)

// ============================================================
// Templates
// ============================================================

var tmplPages = map[string]*template.Template{}

func buildFuncMap() template.FuncMap {
	return template.FuncMap{
		"date":    func(t time.Time) string { return t.Format("02/01/2006") },
		"datep":   func(t *time.Time) string { if t == nil { return "" }; return t.Format("02/01/2006") },
		"dateInput": func(t *time.Time) string { if t == nil { return "" }; return t.Format("2006-01-02") },
		"contains": func(slice []string, s string) bool {
			for _, v := range slice {
				if v == s {
					return true
				}
			}
			return false
		},
		"checkmark": func(b bool) string {
			if b {
				return "Sim"
			}
			return "Nao"
		},
		"statusLabel": func(s string) string {
			switch s {
			case "ativo":
				return "Ativo"
			case "livre":
				return "Livre"
			case "inativo":
				return "Inativo"
			}
			return s
		},
		"derefInt": func(p *int) int {
			if p == nil {
				return 0
			}
			return *p
		},
		"versaoBadge": func(v string) string {
			switch v {
			case "D12":
				return "purple"
			case "XE3":
				return "blue"
			case "Interbase":
				return "orange"
			case "HTML5Builder":
				return "teal"
			}
			return "gray"
		},
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"mul": func(a, b int) int { return a * b },
		"gt":  func(a, b int) bool { return a > b },
		"lt":  func(a, b int) bool { return a < b },
	}
}

func InitTemplates() {
	pages := []string{"dashboard", "desenvolvedores", "licencas", "componentes", "usuarios", "auxiliares", "auditoria", "importexport"}
	for _, page := range pages {
		t := template.New("").Funcs(buildFuncMap())
		template.Must(t.ParseFiles("templates/base.html", "templates/"+page+".html"))
		tmplPages[page] = t
	}
	lt := template.Must(template.New("login").Funcs(buildFuncMap()).ParseFiles("templates/login.html"))
	tmplPages["login"] = lt
}

func render(w http.ResponseWriter, page string, data any) {
	t, ok := tmplPages[page]
	if !ok {
		http.Error(w, "template nao encontrado: "+page, http.StatusInternalServerError)
		return
	}
	if err := t.ExecuteTemplate(w, "base", data); err != nil {
		log.Printf("template %s error: %v", page, err)
		http.Error(w, "Erro interno", http.StatusInternalServerError)
	}
}

func renderLogin(w http.ResponseWriter, data any) {
	t := tmplPages["login"]
	if err := t.ExecuteTemplate(w, "login", data); err != nil {
		log.Printf("login template error: %v", err)
	}
}

func currentUser(db *sql.DB, r *http.Request) *models.Usuario {
	return auth.GetUserFromSession(db, r)
}

func currentUserName(db *sql.DB, r *http.Request) string {
	u := currentUser(db, r)
	if u != nil {
		return u.Nome
	}
	return "sistema"
}

func diffFields(fields [][3]string) string {
	var changes []string
	for _, f := range fields {
		if f[1] != f[2] {
			changes = append(changes, fmt.Sprintf("%s: [%s] → [%s]", f[0], f[1], f[2]))
		}
	}
	if len(changes) == 0 {
		return "Nenhuma alteração"
	}
	return strings.Join(changes, " | ")
}

func basePage(db *sql.DB, r *http.Request, active, title string) models.BasePage {
	return models.BasePage{
		CurrentUser: currentUser(db, r),
		Active:      active,
		Title:       title,
	}
}

func intParam(r *http.Request, key string) int {
	v, _ := strconv.Atoi(r.FormValue(key))
	return v
}

func pathID(r *http.Request) int {
	parts := strings.Split(r.URL.Path, "/")
	for i, p := range parts {
		if p == "id" || (i > 0 && (parts[i-1] == "devs" || parts[i-1] == "licencas" || parts[i-1] == "network" || parts[i-1] == "componentes" || parts[i-1] == "usuarios")) {
			v, _ := strconv.Atoi(p)
			if v > 0 {
				return v
			}
		}
	}
	// fallback: last segment
	segs := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	for i := len(segs) - 1; i >= 0; i-- {
		if v, err := strconv.Atoi(segs[i]); err == nil && v > 0 {
			return v
		}
	}
	return 0
}

// ============================================================
// Login / Logout
// ============================================================

func HandleLogin(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			renderLogin(w, models.LoginPage{Title: "Login"})
			return
		}
		email := strings.TrimSpace(r.FormValue("email"))
		senha := r.FormValue("senha")
		u, err := store.Authenticate(db, email, senha)
		if err != nil {
			renderLogin(w, models.LoginPage{Title: "Login", Erro: "Email ou senha invalidos"})
			return
		}
		token, err := auth.CreateSession(db, u.ID)
		if err != nil {
			renderLogin(w, models.LoginPage{Title: "Login", Erro: "Erro ao criar sessao"})
			return
		}
		auth.SetSessionCookie(w, token)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func HandleLogout(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("dlicsession")
		if err == nil {
			auth.DeleteSession(db, c.Value)
		}
		auth.ClearSessionCookie(w)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

// ============================================================
// Dashboard
// ============================================================

func HandleDashboard(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		totalDevs, devsAtivos, devsLivres, totalLic, licEmUso, licLivres, xe3, d12, ib, compTotal, compPagos, compFree := store.DashboardStats(db)
		recentes, _ := store.ListLicencas(db)
		if len(recentes) > 10 {
			recentes = recentes[:10]
		}
		render(w, "dashboard", models.DashboardData{
			BasePage:         basePage(db, r, "dashboard", "Dashboard"),
			TotalDevs:        totalDevs,
			DevsAtivos:       devsAtivos,
			DevsLivres:       devsLivres,
			TotalLicencas:    totalLic,
			LicencasEmUso:    licEmUso,
			LicencasLivres:   licLivres,
			TotalLicencasXE3: xe3,
			TotalLicencasD12: d12,
			TotalLicencasIB:  ib,
			TotalComponentes: compTotal,
			ComponentesPagos: compPagos,
			ComponentesFree:  compFree,
			RecentesLicencas: recentes,
		})
	}
}

// ============================================================
// Desenvolvedores
// ============================================================

func HandleDevs(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		devs, _ := store.ListDevs(db)
		equipes := store.ListAuxiliaresByTipo(db, "equipe")
		if len(equipes) == 0 {
			equipes = store.ListEquipes(db)
		}
		render(w, "desenvolvedores", models.DevsPage{
			BasePage: basePage(db, r, "desenvolvedores", "Desenvolvedores"),
			Devs:     devs,
			Equipes:  equipes,
		})
	}
}

func HandleCreateDev(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nome := strings.TrimSpace(r.FormValue("nome"))
		equipe := strings.TrimSpace(r.FormValue("equipe"))
		status := r.FormValue("status")
		obs := strings.TrimSpace(r.FormValue("obs"))
		if nome == "" {
			devs, _ := store.ListDevs(db)
			render(w, "desenvolvedores", models.DevsPage{
				BasePage: basePage(db, r, "desenvolvedores", "Desenvolvedores"),
				Devs:     devs,
				Equipes:  store.ListEquipes(db),
				Erro:     "Nome e obrigatorio",
			})
			return
		}
		if err := store.CreateDev(db, nome, equipe, status, obs); err != nil {
			log.Printf("CreateDev: %v", err)
		} else {
			var newID int
			db.QueryRow(`SELECT id FROM devs WHERE nome=$1 ORDER BY id DESC LIMIT 1`, nome).Scan(&newID)
			store.RegistrarAuditoria(db, "dev", newID, "criar",
				currentUserName(db, r),
				fmt.Sprintf("Nome: %s | Equipe: %s | Status: %s", nome, equipe, status))
		}
		http.Redirect(w, r, "/desenvolvedores", http.StatusSeeOther)
	}
}

func HandleUpdateDev(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathID(r)
		old, _ := store.GetDev(db, id)
		nome := strings.TrimSpace(r.FormValue("nome"))
		equipe := strings.TrimSpace(r.FormValue("equipe"))
		status := r.FormValue("status")
		obs := strings.TrimSpace(r.FormValue("obs"))
		if err := store.UpdateDev(db, id, nome, equipe, status, obs); err != nil {
			log.Printf("UpdateDev: %v", err)
		} else {
			detalhes := diffFields([][3]string{
				{"Nome", old.Nome, nome},
				{"Equipe", old.Equipe, equipe},
				{"Status", old.Status, status},
				{"Obs", old.Obs, obs},
			})
			store.RegistrarAuditoria(db, "dev", id, "editar", currentUserName(db, r), detalhes)
		}
		http.Redirect(w, r, "/desenvolvedores", http.StatusSeeOther)
	}
}

func HandleDeleteDev(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathID(r)
		dev, _ := store.GetDev(db, id)
		store.DeleteDev(db, id)
		store.RegistrarAuditoria(db, "dev", id, "excluir",
			currentUserName(db, r),
			fmt.Sprintf("Dev excluído: %s", dev.Nome))
		http.Redirect(w, r, "/desenvolvedores", http.StatusSeeOther)
	}
}

// ============================================================
// Licencas
// ============================================================

func HandleLicencas(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		devIDFilter := 0
		if v := r.URL.Query().Get("dev_id"); v != "" {
			devIDFilter, _ = strconv.Atoi(v)
		}
		var licencas []models.Licenca
		if devIDFilter > 0 {
			licencas, _ = store.ListLicencasByDev(db, devIDFilter)
		} else {
			licencas, _ = store.ListLicencas(db)
		}
		devs, _ := store.ListDevs(db)
		grupos, _ := store.ListGrupos(db)
		versoes := store.ListAuxiliaresByTipo(db, "versao")
		if len(versoes) == 0 {
			versoes = []string{"XE3", "D12", "Interbase", "HTML5Builder"}
		}
		tipos := store.ListAuxiliaresByTipo(db, "tipo_lic")
		if len(tipos) == 0 {
			tipos = []string{"Professional", "Enterprise", "Network"}
		}
		canais := store.ListAuxiliaresByTipo(db, "canal")
		if len(canais) == 0 {
			canais = []string{"EDN", "Network"}
		}
		render(w, "licencas", models.LicencasPage{
			BasePage: basePage(db, r, "licencas", "Licencas Delphi"),
			Licencas: licencas,
			Devs:     devs,
			Grupos:   grupos,
			Versoes:  versoes,
			Tipos:    tipos,
			Canais:   canais,
			DevID:    devIDFilter,
		})
	}
}

func optionalDevID(r *http.Request) *int {
	v := r.FormValue("dev_id")
	if v == "" || v == "0" {
		return nil
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return nil
	}
	return &n
}

func optionalGrupoID(r *http.Request) *int {
	v := r.FormValue("grupo_id")
	if v == "" || v == "0" {
		return nil
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return nil
	}
	return &n
}

func HandleCreateLicenca(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		devID := optionalDevID(r)
		versao := r.FormValue("versao")
		serial := strings.TrimSpace(r.FormValue("serial"))
		tipo := r.FormValue("tipo")
		canal := r.FormValue("canal")
		if canal == "" {
			canal = "EDN"
		}
		hostname := strings.TrimSpace(r.FormValue("hostname"))
		ednLogin := strings.TrimSpace(r.FormValue("edn_login"))
		ednSenha := strings.TrimSpace(r.FormValue("edn_senha"))
		cadEfetuado := r.FormValue("cad_efetuado") == "true"
		obs := strings.TrimSpace(r.FormValue("obs"))

		var dataCad *time.Time
		if ds := r.FormValue("data_cad"); ds != "" {
			if t, err := time.Parse("2006-01-02", ds); err == nil {
				dataCad = &t
			}
		}
		grupoID := optionalGrupoID(r)
		if err := store.CreateLicenca(db, devID, grupoID, versao, serial, tipo, canal, hostname, ednLogin, ednSenha, cadEfetuado, dataCad, obs); err != nil {
			log.Printf("CreateLicenca: %v", err)
		} else {
			// pega o ID da licença recém-criada
			var newID int
			db.QueryRow(`SELECT id FROM licencas WHERE serial=$1 ORDER BY id DESC LIMIT 1`, serial).Scan(&newID)
			store.RegistrarAuditoria(db, "licenca", newID, "criar",
				currentUserName(db, r),
				fmt.Sprintf("Versão: %s | Serial: %s | Tipo: %s | Controle: %s | Hostname: %s", versao, serial, tipo, canal, hostname))
		}
		http.Redirect(w, r, "/licencas", http.StatusSeeOther)
	}
}

func HandleUpdateLicenca(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathID(r)
		old, _ := store.GetLicenca(db, id)
		devID := optionalDevID(r)
		versao := r.FormValue("versao")
		serial := strings.TrimSpace(r.FormValue("serial"))
		tipo := r.FormValue("tipo")
		canal := r.FormValue("canal")
		if canal == "" {
			canal = "EDN"
		}
		hostname := strings.TrimSpace(r.FormValue("hostname"))
		ednLogin := strings.TrimSpace(r.FormValue("edn_login"))
		ednSenha := strings.TrimSpace(r.FormValue("edn_senha"))
		cadEfetuado := r.FormValue("cad_efetuado") == "true"
		obs := strings.TrimSpace(r.FormValue("obs"))

		var dataCad *time.Time
		if ds := r.FormValue("data_cad"); ds != "" {
			if t, err := time.Parse("2006-01-02", ds); err == nil {
				dataCad = &t
			}
		}
		grupoID := optionalGrupoID(r)
		if err := store.UpdateLicenca(db, id, devID, grupoID, versao, serial, tipo, canal, hostname, ednLogin, ednSenha, cadEfetuado, dataCad, obs); err != nil {
			log.Printf("UpdateLicenca: %v", err)
		} else {
			oldCad := "Não"
			if old.CadEfetuado {
				oldCad = "Sim"
			}
			newCad := "Não"
			if cadEfetuado {
				newCad = "Sim"
			}
			detalhes := diffFields([][3]string{
				{"Versão", old.Versao, versao},
				{"Serial", old.Serial, serial},
				{"Tipo", old.Tipo, tipo},
				{"Controle", old.Canal, canal},
				{"Hostname", old.Hostname, hostname},
				{"EDN Login", old.EdnLogin, ednLogin},
				{"EDN Senha", old.EdnSenha, ednSenha},
				{"Cad. Efetuado", oldCad, newCad},
				{"Obs", old.Obs, obs},
			})
			store.RegistrarAuditoria(db, "licenca", id, "editar", currentUserName(db, r), detalhes)
		}
		http.Redirect(w, r, "/licencas", http.StatusSeeOther)
	}
}

func HandleVincularDevLicenca(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathID(r)
		old, _ := store.GetLicenca(db, id)
		devID := optionalDevID(r)
		if err := store.VincularDevLicenca(db, id, devID); err != nil {
			log.Printf("VincularDevLicenca: %v", err)
		} else {
			oldDev := "(livre)"
			if old.DevNome != "" {
				oldDev = old.DevNome
			}
			acao := "desvincular"
			detalhe := fmt.Sprintf("Dev: [%s] → [(livre)]", oldDev)
			if devID != nil {
				acao = "vincular"
				var devNome string
				db.QueryRow(`SELECT nome FROM devs WHERE id=$1`, *devID).Scan(&devNome)
				detalhe = fmt.Sprintf("Dev: [%s] → [%s]", oldDev, devNome)
			}
			store.RegistrarAuditoria(db, "licenca", id, acao, currentUserName(db, r), detalhe)
		}
		http.Redirect(w, r, "/licencas", http.StatusSeeOther)
	}
}

func HandleDeleteLicenca(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathID(r)
		serial := store.GetLicencaSerial(db, id)
		store.DeleteLicenca(db, id)
		store.RegistrarAuditoria(db, "licenca", id, "excluir",
			currentUserName(db, r),
			fmt.Sprintf("Serial excluído: %s", serial))
		http.Redirect(w, r, "/licencas", http.StatusSeeOther)
	}
}

func HandleCreateGrupo(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nome := strings.TrimSpace(r.FormValue("nome"))
		obs := strings.TrimSpace(r.FormValue("obs"))
		if nome != "" {
			if err := store.CreateGrupo(db, nome, obs); err != nil {
				log.Printf("CreateGrupo: %v", err)
			}
		}
		http.Redirect(w, r, "/licencas", http.StatusSeeOther)
	}
}

func HandleDeleteGrupo(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		store.DeleteGrupo(db, pathID(r))
		http.Redirect(w, r, "/licencas", http.StatusSeeOther)
	}
}

// ============================================================
// Auxiliares
// ============================================================

func HandleAuxiliares(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		render(w, "auxiliares", models.AuxiliaresPage{
			BasePage: basePage(db, r, "auxiliares", "Auxiliares"),
			Equipes:  store.ListAuxiliarObjs(db, "equipe"),
			Versoes:  store.ListAuxiliarObjs(db, "versao"),
			TiposLic: store.ListAuxiliarObjs(db, "tipo_lic"),
			Canais:   store.ListAuxiliarObjs(db, "canal"),
		})
	}
}

func HandleCreateAuxiliar(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tipo := r.FormValue("tipo")
		nome := strings.TrimSpace(r.FormValue("nome"))
		if tipo != "" && nome != "" {
			if err := store.CreateAuxiliar(db, tipo, nome); err != nil {
				log.Printf("CreateAuxiliar: %v", err)
			}
		}
		http.Redirect(w, r, "/auxiliares", http.StatusSeeOther)
	}
}

func HandleDeleteAuxiliar(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		store.DeleteAuxiliar(db, pathID(r))
		http.Redirect(w, r, "/auxiliares", http.StatusSeeOther)
	}
}

// ============================================================
// Componentes
// ============================================================

func HandleComponentes(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		comps, _ := store.ListComponentes(db)
		ferramentas := store.ListFerramentas(db)
		render(w, "componentes", models.ComponentesPage{
			BasePage:       basePage(db, r, "componentes", "Componentes Delphi"),
			Componentes:    comps,
			Ferramentas:    ferramentas,
			Licenciamentos: []string{"Pago", "Free", "Pago Por uso"},
		})
	}
}

func HandleCreateComponente(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nome := strings.TrimSpace(r.FormValue("nome"))
		if nome == "" {
			http.Redirect(w, r, "/componentes", http.StatusSeeOther)
			return
		}
		licenciamento := r.FormValue("licenciamento")
		store.CreateComponente(db,
			nome,
			strings.TrimSpace(r.FormValue("versao")),
			strings.TrimSpace(r.FormValue("ferramenta")),
			strings.TrimSpace(r.FormValue("informacoes")),
			strings.TrimSpace(r.FormValue("uso")),
			licenciamento,
			strings.TrimSpace(r.FormValue("site")),
			strings.TrimSpace(r.FormValue("obs")),
			strings.TrimSpace(r.FormValue("serial")),
			strings.TrimSpace(r.FormValue("usuario")),
			strings.TrimSpace(r.FormValue("senha")),
		)
		var newID int
		db.QueryRow(`SELECT id FROM componentes WHERE nome=$1 ORDER BY id DESC LIMIT 1`, nome).Scan(&newID)
		store.RegistrarAuditoria(db, "componente", newID, "criar",
			currentUserName(db, r),
			fmt.Sprintf("Nome: %s | Licenciamento: %s", nome, licenciamento))
		http.Redirect(w, r, "/componentes", http.StatusSeeOther)
	}
}

func HandleUpdateComponente(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathID(r)
		old, _ := store.GetComponente(db, id)
		nome := strings.TrimSpace(r.FormValue("nome"))
		versao := strings.TrimSpace(r.FormValue("versao"))
		ferramenta := strings.TrimSpace(r.FormValue("ferramenta"))
		informacoes := strings.TrimSpace(r.FormValue("informacoes"))
		uso := strings.TrimSpace(r.FormValue("uso"))
		licenciamento := r.FormValue("licenciamento")
		site := strings.TrimSpace(r.FormValue("site"))
		obs := strings.TrimSpace(r.FormValue("obs"))
		serial := strings.TrimSpace(r.FormValue("serial"))
		usuario := strings.TrimSpace(r.FormValue("usuario"))
		senha := strings.TrimSpace(r.FormValue("senha"))

		store.UpdateComponente(db, id, nome, versao, ferramenta, informacoes, uso, licenciamento, site, obs, serial, usuario, senha)
		detalhes := diffFields([][3]string{
			{"Nome", old.Nome, nome},
			{"Versão", old.Versao, versao},
			{"Ferramenta", old.Ferramenta, ferramenta},
			{"Uso", old.Uso, uso},
			{"Licenciamento", old.Licenciamento, licenciamento},
			{"Serial", old.Serial, serial},
			{"Usuário", old.Usuario, usuario},
			{"Senha", old.Senha, senha},
			{"Site", old.Site, site},
			{"Informações", old.Informacoes, informacoes},
			{"Obs", old.Obs, obs},
		})
		store.RegistrarAuditoria(db, "componente", id, "editar", currentUserName(db, r), detalhes)
		http.Redirect(w, r, "/componentes", http.StatusSeeOther)
	}
}

func HandleDeleteComponente(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathID(r)
		var nome string
		db.QueryRow(`SELECT nome FROM componentes WHERE id=$1`, id).Scan(&nome)
		store.DeleteComponente(db, id)
		store.RegistrarAuditoria(db, "componente", id, "excluir",
			currentUserName(db, r),
			fmt.Sprintf("Componente excluído: %s", nome))
		http.Redirect(w, r, "/componentes", http.StatusSeeOther)
	}
}

// ============================================================
// Usuarios (admin only)
// ============================================================

func HandleUsuarios(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		usuarios, _ := store.ListUsuarios(db)
		render(w, "usuarios", models.UsuariosPage{
			BasePage: basePage(db, r, "usuarios", "Usuarios"),
			Usuarios: usuarios,
		})
	}
}

func HandleCreateUsuario(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nome := strings.TrimSpace(r.FormValue("nome"))
		email := strings.TrimSpace(r.FormValue("email"))
		senha := r.FormValue("senha")
		admin := r.FormValue("admin") == "true"
		if err := store.CreateUsuario(db, nome, email, senha, admin); err != nil {
			log.Printf("CreateUsuario: %v", err)
		}
		http.Redirect(w, r, "/usuarios", http.StatusSeeOther)
	}
}

func HandleToggleUsuarioAtivo(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		store.ToggleUsuarioAtivo(db, pathID(r))
		http.Redirect(w, r, "/usuarios", http.StatusSeeOther)
	}
}

func HandleToggleUsuarioAdmin(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		store.ToggleUsuarioAdmin(db, pathID(r))
		http.Redirect(w, r, "/usuarios", http.StatusSeeOther)
	}
}

func HandleResetSenha(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathID(r)
		nova := r.FormValue("nova_senha")
		if nova == "" {
			nova = "senha123"
		}
		store.ResetSenha(db, id, nova)
		http.Redirect(w, r, "/usuarios", http.StatusSeeOther)
	}
}

func HandleMinhaSenha(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := currentUser(db, r)
		if u == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		atual := r.FormValue("senha_atual")
		nova := r.FormValue("nova_senha")

		var hash string
		db.QueryRow(`SELECT senha_hash FROM usuarios WHERE id=$1`, u.ID).Scan(&hash)
		if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(atual)); err != nil {
			http.Error(w, "Senha atual incorreta", http.StatusBadRequest)
			return
		}
		store.ResetSenha(db, u.ID, nova)
		fmt.Fprint(w, "Senha alterada com sucesso")
	}
}

// ============================================================
// Auditoria
// ============================================================

func HandleAuditoria(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		entidade := r.URL.Query().Get("entidade")
		entidadeID, _ := strconv.Atoi(r.URL.Query().Get("entidade_id"))
		pagina, _ := strconv.Atoi(r.URL.Query().Get("pagina"))
		if pagina < 1 {
			pagina = 1
		}
		porPagina := 50
		offset := (pagina - 1) * porPagina

		logs, total := store.ListAuditLogs(db, entidade, entidadeID, porPagina, offset)
		render(w, "auditoria", models.AuditoriaPage{
			BasePage:   basePage(db, r, "auditoria", "Auditoria"),
			Logs:       logs,
			Entidade:   entidade,
			EntidadeID: entidadeID,
			Total:      total,
			Pagina:     pagina,
			PorPagina:  porPagina,
		})
	}
}

// ============================================================
// Export Auditoria CSV
// ============================================================

func HandleExportAuditCSV(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		entidade := r.URL.Query().Get("entidade")
		entidadeID, _ := strconv.Atoi(r.URL.Query().Get("entidade_id"))

		logs, _ := store.ListAuditLogs(db, entidade, entidadeID, 10000, 0)

		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", "attachment; filename=auditoria.csv")
		w.Write([]byte("\xEF\xBB\xBF")) // BOM for Excel

		cw := csv.NewWriter(w)
		cw.Comma = ';'
		cw.Write([]string{"Data/Hora", "Usuário", "Entidade", "ID", "Ação", "Detalhes"})
		for _, l := range logs {
			cw.Write([]string{
				l.CriadoEm.Format("02/01/2006 15:04:05"),
				l.Usuario,
				l.Entidade,
				strconv.Itoa(l.EntidadeID),
				l.Acao,
				l.Detalhes,
			})
		}
		cw.Flush()
	}
}

// ============================================================
// Importar / Exportar
// ============================================================

func HandleImportExport(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		render(w, "importexport", struct {
			models.BasePage
			Erro    string
			Sucesso string
		}{
			BasePage: basePage(db, r, "importexport", "Importar / Exportar"),
		})
	}
}

func HandleExportCSV(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tipo := r.URL.Query().Get("tipo")

		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.csv", tipo))
		w.Write([]byte("\xEF\xBB\xBF"))

		cw := csv.NewWriter(w)
		cw.Comma = ';'

		switch tipo {
		case "desenvolvedores":
			cw.Write([]string{"Nome", "Equipe", "Status", "Obs"})
			devs, _ := store.ListDevs(db)
			for _, d := range devs {
				cw.Write([]string{d.Nome, d.Equipe, d.Status, d.Obs})
			}
		case "licencas":
			cw.Write([]string{"Dev", "Grupo", "Versão", "Serial", "Tipo", "Controle", "Hostname", "EDN Login", "EDN Senha", "Cad. Efetuado", "Data Cad.", "Obs"})
			lics, _ := store.ListLicencas(db)
			for _, l := range lics {
				cadEfetuado := "Não"
				if l.CadEfetuado {
					cadEfetuado = "Sim"
				}
				dataCad := ""
				if l.DataCad != nil {
					dataCad = l.DataCad.Format("02/01/2006")
				}
				cw.Write([]string{l.DevNome, l.GrupoNome, l.Versao, l.Serial, l.Tipo, l.Canal, l.Hostname, l.EdnLogin, l.EdnSenha, cadEfetuado, dataCad, l.Obs})
			}
		case "componentes":
			cw.Write([]string{"Nome", "Versão", "Ferramenta", "Uso", "Licenciamento", "Serial", "Usuário", "Senha", "Site", "Informações", "Obs"})
			comps, _ := store.ListComponentes(db)
			for _, c := range comps {
				cw.Write([]string{c.Nome, c.Versao, c.Ferramenta, c.Uso, c.Licenciamento, c.Serial, c.Usuario, c.Senha, c.Site, c.Informacoes, c.Obs})
			}
		default:
			http.Error(w, "Tipo inválido", http.StatusBadRequest)
			return
		}
		cw.Flush()
	}
}

func HandleImportCSV(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tipo := r.FormValue("tipo")
		file, _, err := r.FormFile("arquivo")
		if err != nil {
			renderImportExport(w, db, r, "Erro ao ler arquivo: "+err.Error(), "")
			return
		}
		defer file.Close()

		cr := csv.NewReader(file)
		cr.Comma = ';'
		cr.LazyQuotes = true

		// Skip header
		if _, err := cr.Read(); err != nil {
			renderImportExport(w, db, r, "Arquivo vazio ou inválido", "")
			return
		}

		userName := currentUserName(db, r)
		var count int
		var importErr string

		for {
			record, err := cr.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				importErr = fmt.Sprintf("Erro na linha %d: %v", count+2, err)
				break
			}

			switch tipo {
			case "desenvolvedores":
				if len(record) < 4 {
					importErr = fmt.Sprintf("Linha %d: esperado 4 colunas (Nome, Equipe, Status, Obs)", count+2)
				} else if err := store.CreateDev(db, record[0], record[1], record[2], record[3]); err != nil {
					importErr = fmt.Sprintf("Linha %d: %v", count+2, err)
				} else {
					store.RegistrarAuditoria(db, "dev", 0, "criar", userName, "Importado via CSV: "+record[0])
				}
			case "licencas":
				if len(record) < 12 {
					importErr = fmt.Sprintf("Linha %d: esperado 12 colunas (Dev; Grupo; Versão; Serial; Tipo; Controle; Hostname; EDN Login; EDN Senha; Cad. Efetuado; Data Cad.; Obs)", count+2)
				} else {
					devID := store.FindDevIDByNome(db, strings.TrimSpace(record[0]))
					grupoID := store.FindGrupoIDByNome(db, strings.TrimSpace(record[1]))
					cadEfetuado := strings.EqualFold(strings.TrimSpace(record[9]), "sim")
					var dataCad *time.Time
					if ds := strings.TrimSpace(record[10]); ds != "" {
						if t, err := time.Parse("02/01/2006", ds); err == nil {
							dataCad = &t
						}
					}
					canal := strings.TrimSpace(record[5])
					if canal == "" {
						canal = "EDN"
					}
					if err := store.CreateLicenca(db, devID, grupoID, strings.TrimSpace(record[2]), strings.TrimSpace(record[3]), strings.TrimSpace(record[4]), canal, strings.TrimSpace(record[6]), strings.TrimSpace(record[7]), strings.TrimSpace(record[8]), cadEfetuado, dataCad, strings.TrimSpace(record[11])); err != nil {
						importErr = fmt.Sprintf("Linha %d: %v", count+2, err)
					} else {
						store.RegistrarAuditoria(db, "licenca", 0, "criar", userName, "Importado via CSV: "+strings.TrimSpace(record[3]))
					}
				}
			case "componentes":
				if len(record) < 11 {
					importErr = fmt.Sprintf("Linha %d: esperado 11 colunas", count+2)
				} else if err := store.CreateComponente(db, record[0], record[1], record[2], record[9], record[3], record[4], record[8], record[10], record[5], record[6], record[7]); err != nil {
					importErr = fmt.Sprintf("Linha %d: %v", count+2, err)
				} else {
					store.RegistrarAuditoria(db, "componente", 0, "criar", userName, "Importado via CSV: "+record[0])
				}
			default:
				renderImportExport(w, db, r, "Tipo de importação inválido: "+tipo, "")
				return
			}
			if importErr != "" {
				break
			}
			count++
		}

		if importErr != "" {
			renderImportExport(w, db, r, importErr, "")
			return
		}
		renderImportExport(w, db, r, "", fmt.Sprintf("%d registros importados com sucesso!", count))
	}
}

func renderImportExport(w http.ResponseWriter, db *sql.DB, r *http.Request, erro, sucesso string) {
	render(w, "importexport", struct {
		models.BasePage
		Erro    string
		Sucesso string
	}{
		BasePage: basePage(db, r, "importexport", "Importar / Exportar"),
		Erro:     erro,
		Sucesso:  sucesso,
	})
}
