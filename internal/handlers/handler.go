package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/salliko/gofemart/config"
	"github.com/salliko/gofemart/internal/databases"
	"github.com/salliko/gofemart/internal/datahashes"
)

func UserRegistration(cfg config.Config, db databases.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user databases.User

		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			log.Print(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		userID, err := datahashes.RandBytes(10)
		if err != nil {
			log.Print(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		user.UserID = userID

		user, err = db.CreateUser(user.Login, user.Password, user.UserID)
		if err != nil {
			if errors.Is(err, databases.ErrLoginConfict) {
				w.Header().Set("Content-Type", "application/json; charset=UTF-8")
				w.WriteHeader(http.StatusConflict)
				w.Write([]byte("логин уже занят"))
				return
			} else {
				log.Print(err.Error())
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}

		newCookie := &http.Cookie{
			Name:  "user_id",
			Value: user.UserID,
		}

		http.SetCookie(w, newCookie)

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("пользователь успешно зарегистрирован и аутентифицирован"))
	}
}
