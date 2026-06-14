package main

import (
	"log"
	"net/http"
	"os"

	"delphilic/internal/auth"
	"delphilic/internal/handler"
	"delphilic/internal/store"
)

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	db := store.NewDB()
	defer db.Close()

	if err := store.RunMigrations(db); err != nil {
		log.Fatal("migrate:", err)
	}
	if err := store.EnsureAdmin(db); err != nil {
		log.Fatal("ensure admin:", err)
	}

	handler.InitTemplates()
	mux := http.NewServeMux()

	// Estaticos
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Login / Logout
	mux.HandleFunc("GET /login", handler.HandleLogin(db))
	mux.HandleFunc("POST /login", handler.HandleLogin(db))
	mux.HandleFunc("POST /logout", handler.HandleLogout(db))

	// Dashboard
	mux.HandleFunc("GET /", auth.Protected(db, handler.HandleDashboard(db)))

	// Desenvolvedores
	mux.HandleFunc("GET /desenvolvedores", auth.Protected(db, handler.HandleDevs(db)))
	mux.HandleFunc("POST /desenvolvedores", auth.Protected(db, handler.HandleCreateDev(db)))
	mux.HandleFunc("POST /desenvolvedores/{id}/update", auth.Protected(db, handler.HandleUpdateDev(db)))
	mux.HandleFunc("POST /desenvolvedores/{id}/delete", auth.Protected(db, handler.HandleDeleteDev(db)))

	// Licencas Delphi por Dev
	mux.HandleFunc("GET /licencas", auth.Protected(db, handler.HandleLicencas(db)))
	mux.HandleFunc("POST /licencas", auth.Protected(db, handler.HandleCreateLicenca(db)))
	mux.HandleFunc("POST /licencas/{id}/update", auth.Protected(db, handler.HandleUpdateLicenca(db)))
	mux.HandleFunc("POST /licencas/{id}/vincular", auth.Protected(db, handler.HandleVincularDevLicenca(db)))
	mux.HandleFunc("POST /licencas/{id}/delete", auth.Protected(db, handler.HandleDeleteLicenca(db)))

	// Grupos de Licenca
	mux.HandleFunc("POST /grupos", auth.Protected(db, handler.HandleCreateGrupo(db)))
	mux.HandleFunc("POST /grupos/{id}/delete", auth.Protected(db, handler.HandleDeleteGrupo(db)))

	// Auxiliares
	mux.HandleFunc("GET /auxiliares", auth.Protected(db, handler.HandleAuxiliares(db)))
	mux.HandleFunc("POST /auxiliares", auth.Protected(db, handler.HandleCreateAuxiliar(db)))
	mux.HandleFunc("POST /auxiliares/{id}/delete", auth.Protected(db, handler.HandleDeleteAuxiliar(db)))

	// Componentes Delphi
	mux.HandleFunc("GET /componentes", auth.Protected(db, handler.HandleComponentes(db)))
	mux.HandleFunc("POST /componentes", auth.Protected(db, handler.HandleCreateComponente(db)))
	mux.HandleFunc("POST /componentes/{id}/update", auth.Protected(db, handler.HandleUpdateComponente(db)))
	mux.HandleFunc("POST /componentes/{id}/delete", auth.Protected(db, handler.HandleDeleteComponente(db)))

	// Auditoria
	mux.HandleFunc("GET /auditoria", auth.Protected(db, handler.HandleAuditoria(db)))
	mux.HandleFunc("GET /auditoria/export", auth.Protected(db, handler.HandleExportAuditCSV(db)))

	// Importar / Exportar
	mux.HandleFunc("GET /importexport", auth.Protected(db, handler.HandleImportExport(db)))
	mux.HandleFunc("GET /importexport/export", auth.Protected(db, handler.HandleExportCSV(db)))
	mux.HandleFunc("POST /importexport/import", auth.Protected(db, handler.HandleImportCSV(db)))

	// Usuarios (admin only)
	mux.HandleFunc("GET /usuarios", auth.AdminOnly(db, handler.HandleUsuarios(db)))
	mux.HandleFunc("POST /usuarios", auth.AdminOnly(db, handler.HandleCreateUsuario(db)))
	mux.HandleFunc("POST /usuarios/{id}/toggle-ativo", auth.AdminOnly(db, handler.HandleToggleUsuarioAtivo(db)))
	mux.HandleFunc("POST /usuarios/{id}/toggle-admin", auth.AdminOnly(db, handler.HandleToggleUsuarioAdmin(db)))
	mux.HandleFunc("POST /usuarios/{id}/reset-senha", auth.AdminOnly(db, handler.HandleResetSenha(db)))

	// Alterar propria senha
	mux.HandleFunc("POST /minha-senha", auth.Protected(db, handler.HandleMinhaSenha(db)))

	port := getEnv("PORT", "8081")
	log.Printf("delphiLic iniciado em http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
