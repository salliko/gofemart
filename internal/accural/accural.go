package accural

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/salliko/gofemart/internal/databases"
)

var ErrAnother = errors.New(`прочая ошибка`)

func GetAccural(URL string) (databases.Order, error) {
	log.Print("accural: выполняем запрос на URL: ", URL)
	var order databases.Order
	resp, err := http.Get(URL)
	if err != nil {
		log.Println("ошибка на GET ", URL, err)
		return order, err
	}
	defer resp.Body.Close()
	bodyData, _ := io.ReadAll(resp.Body)
	log.Print("client GET ", URL, resp.StatusCode, string(bodyData))

	if resp.StatusCode == http.StatusOK {
		if err := json.NewDecoder(bytes.NewReader(bodyData)).Decode(&order); err != nil {
			log.Print("Получен заказ на ", URL)
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
