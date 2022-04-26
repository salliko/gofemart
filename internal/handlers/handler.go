package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/salliko/gofemart/config"
	"github.com/salliko/gofemart/internal/accural"
	"github.com/salliko/gofemart/internal/databases"
	"github.com/salliko/gofemart/internal/datahashes"
)

func checkLuhn(number string) (bool, error) {
	sum, err := strconv.Atoi(string(number[len(number)-1]))
	if err != nil {
		return false, err
	}

	parity := len(number) % 2

	for i := len(number) - 2; i >= 0; i-- {
		summand, err := strconv.Atoi(string(number[i]))
		if err != nil {
			return false, err
		}
		if i%2 == parity {
			product := summand * 2
			if product > 9 {
				summand = product - 9
			} else {
				summand = product
			}
		}
		sum += summand
	}

	res := (sum % 10) == 0
	return res, nil
}

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
		} else {
			log.Print("Создаем печеньку с индефикатором пользователя")
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

		isValidNumber, err := checkLuhn(string(number))
		if err != nil {
			log.Print(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if !isValidNumber {
			log.Print("неверный формат номера заказа")
			http.Error(w, "неверный формат номера заказа", http.StatusUnprocessableEntity)
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

		go func(userID, number string, cfg config.Config, db databases.Database) {
			URL := fmt.Sprintf("%s/api/orders/%s", cfg.ActualSystemAddress, number)
			order, err := accural.GetAccural(URL)
			if err != nil {
				log.Print(err.Error())
				return
			}

			err = db.UpdateOrder(userID, order)
			if err != nil {
				log.Print(err.Error())
				return
			}

		}(cookie.Value, string(number), cfg, db)

		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("новый номер заказа принят в обработку"))
	}
}

func SelectOrders(cfg config.Config, db databases.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("user_id")
		if err != nil {
			log.Print(err.Error())
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		orders, err := db.SelectOrders(cookie.Value)
		if err != nil {
			if errors.Is(err, databases.ErrNotFoundOrders) {
				w.Header().Set("Content-Type", "application/json; charset=UTF-8")
				w.WriteHeader(http.StatusNoContent)
				http.Error(w, err.Error(), http.StatusNoContent)
				return
			}
			log.Print(err.Error())
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusInternalServerError)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data, err := json.Marshal(orders)
		if err != nil {
			log.Print(err.Error())
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusInternalServerError)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

func SelectUserBalance(cfg config.Config, db databases.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("user_id")
		if err != nil {
			log.Print(err.Error())
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusInternalServerError)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		balance, err := db.SelectUserBalance(cookie.Value)
		if err != nil {
			log.Print(err.Error())
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusInternalServerError)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data, err := json.Marshal(balance)
		if err != nil {
			log.Print(err.Error())
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusInternalServerError)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

func CreateDebit(cfg config.Config, db databases.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var withdrawn databases.Withdrawn

		if err := json.NewDecoder(r.Body).Decode(&withdrawn); err != nil {
			log.Print(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		cookie, err := r.Cookie("user_id")
		if err != nil {
			log.Print(err.Error())
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusInternalServerError)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = db.CreateDebit(cookie.Value, withdrawn)
		if err != nil {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")

			switch {
			case errors.Is(err, databases.ErrInsufficientFunds):
				w.WriteHeader(http.StatusPaymentRequired)
				http.Error(w, err.Error(), http.StatusPaymentRequired)
				return
			case errors.Is(err, databases.ErrInvalidOrderNumber):
				w.WriteHeader(http.StatusUnprocessableEntity)
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			default:
				w.WriteHeader(http.StatusInternalServerError)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("успешная обработка запроса"))

	}
}

func SelectUserOperations(cfg config.Config, db databases.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("user_id")
		if err != nil {
			log.Print(err.Error())
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusInternalServerError)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		withdrawns, err := db.SelectUserOperations(cookie.Value)
		if err != nil {
			log.Print(err.Error())
			if errors.Is(err, databases.ErrNotFoundOperations) {
				w.Header().Set("Content-Type", "application/json; charset=UTF-8")
				w.WriteHeader(http.StatusNoContent)
				http.Error(w, err.Error(), http.StatusNoContent)
				return
			}

			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusInternalServerError)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data, err := json.Marshal(withdrawns)
		if err != nil {
			log.Print(err.Error())
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusInternalServerError)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}
