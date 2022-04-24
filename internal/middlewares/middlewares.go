package middlewares

import (
	"log"
	"net/http"
)

func CheckCookie(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := r.Cookie("user_id"); err != nil {
			log.Print("пользователь не аутентифицирован")
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusUnauthorized)
			http.Error(w, "пользователь не аутентифицирован", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
