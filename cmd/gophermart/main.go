package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/salliko/gofemart/config"
	"github.com/salliko/gofemart/internal/databases"
	"github.com/salliko/gofemart/internal/handlers"
	"github.com/salliko/gofemart/internal/middlewares"
)

func NewRouter(cfg config.Config, db databases.Database) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Logger)

	r.Post("/api/user/register", handlers.UserRegistration(cfg, db))
	r.Post("/api/user/login", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("login"))
	})

	r.Route("/api/user", func(r chi.Router) {

		r.Use(middlewares.CheckCookie)

		r.Post("/orders", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("orders"))
		})

		r.Get("/orders", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("orders"))
		})

		r.Get("/withdrawals", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("withdrawals"))
		})

		r.Route("/balance", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("balance"))
			})
			r.Post("/withdraw", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("withdraw"))
			})
		})

	})

	return r

}

func main() {
	var cfg config.Config
	var db databases.Database

	if err := cfg.Parse(); err != nil {
		log.Fatal(err)
	}

	db, err := databases.NewPostgresqlDatabase(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := NewRouter(cfg, db)
	log.Printf("Server running on address: %s", cfg.RunAddress)
	log.Fatal(http.ListenAndServe(cfg.RunAddress, r))
}
