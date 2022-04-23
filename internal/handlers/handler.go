package handlers

import (
	"encoding/json"
	"errors"
	"io"
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

func UserAuthentication(cfg config.Config, db databases.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user databases.User

		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			log.Print(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		user, err := db.SelectUser(user.Login, user.Password)
		if err != nil {
			if errors.Is(err, databases.ErrInvalidUsernamePassword) {
				w.Header().Set("Content-Type", "application/json; charset=UTF-8")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("неверная пара логин/пароль"))
				return
			} else {
				log.Print(err.Error())
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}

		if _, err := r.Cookie("user_id"); err != nil {
			log.Print("Создаем печеньку :)")
			newCookie := &http.Cookie{
				Name:  "user_id",
				Value: user.UserID,
			}
			http.SetCookie(w, newCookie)
		}

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("пользователь успешно аутентифицирован"))
	}
}

func CreateOrder(cfg config.Config, db databases.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		number, err := io.ReadAll(r.Body)

		if err != nil {
			log.Print(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		cookie, err := r.Cookie("user_id")
		if err != nil {
			log.Print(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = db.CreateOrder(string(number), cookie.Value)
		if err != nil {
			if errors.Is(err, databases.ErrOrderWasUploadedBefore) {
				http.Error(w, err.Error(), http.StatusOK)
				return
			}
			if errors.Is(err, databases.ErrOrderWasUploadedAnotherUser) {
				http.Error(w, err.Error(), http.StatusConflict)
				return
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("новый номер заказа принят в обработку"))
	}
}
