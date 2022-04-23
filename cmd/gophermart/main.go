package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/salliko/gofemart/config"
)

func NewRouter() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Logger)

	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("register"))
		})

		r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("login"))
		})

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
	if err := cfg.Parse(); err != nil {
		log.Fatal(err)
	}

	r := NewRouter()
	log.Printf("Server running on address: %s", cfg.RunAddress)
	log.Fatal(http.ListenAndServe(cfg.RunAddress, r))
}
