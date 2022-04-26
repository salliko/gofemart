package accural

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/salliko/gofemart/internal/databases"
)

var ErrAnother = errors.New(`прочая ошибка`)

func GetAccural(URL string) (databases.Order, error) {
	var order databases.Order
	resp, err := http.Get(URL)
	if err != nil {
		return order, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		if err := json.NewDecoder(resp.Body).Decode(&order); err != nil {
			log.Print("tyta")
			return order, err
		}
	} else {
		log.Print(resp.StatusCode)
		log.Print(resp.Body)
		log.Print(URL)
		return order, ErrAnother
	}

	return order, nil
}
