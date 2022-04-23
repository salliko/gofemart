package databases

import (
	"context"
	"errors"
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/salliko/gofemart/config"
)

var ErrLoginConfict = errors.New(`login conglict`)
var ErrInvalidUsernamePassword = errors.New(`invalid username/password pair`)
var ErrOrderWasUploadedBefore = errors.New(`номер заказа уже был загружен этим пользователем`)
var ErrOrderWasUploadedAnotherUser = errors.New(`номер заказа уже был загружен другим пользователем`)

type Database interface {
	CreateUser(string, string, string) (User, error)
	SelectUser(string, string) (User, error)
	HasLogin(string) (bool, error)
	CreateOrder(string, string) error
	Close()
}

type User struct {
	UserID   string `json:"-"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Order struct {
	Number     string `json:"number"`
	Status     string `json:"status"`
	Accural    int    `json:"accural"`
	UploadedAt string `json:"uploaded_at"`
}

type PostgresqlDatabase struct {
	conn *pgxpool.Pool
}

func NewPostgresqlDatabase(cfg config.Config) (*PostgresqlDatabase, error) {
	conn, err := pgxpool.Connect(context.Background(), cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	rows, err := conn.Query(context.Background(), createTableUsers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rows, err = conn.Query(context.Background(), createTableOrders)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return &PostgresqlDatabase{conn: conn}, nil
}

func (p *PostgresqlDatabase) Close() {
	p.conn.Close()
}

func (p *PostgresqlDatabase) HasLogin(login string) (bool, error) {
	var check bool
	err := p.conn.QueryRow(context.Background(), checkLogin, login).Scan(&check)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return check, nil
		}

		return check, err
	}

	return check, nil

}

func (p *PostgresqlDatabase) CreateUser(login, password, userID string) (User, error) {
	var user User
	hasLogin, err := p.HasLogin(login)
	if err != nil {
		return user, err
	}

	if hasLogin {
		return user, ErrLoginConfict
	}

	err = p.conn.QueryRow(context.Background(), createUser, login, password, userID).Scan(&user.UserID, &user.Login, &user.Password)
	if err != nil {
		return user, err
	}

	return user, nil
}

func (p *PostgresqlDatabase) SelectUser(login, password string) (User, error) {
	var user User
	err := p.conn.QueryRow(context.Background(), selectUser, login, password).Scan(&user.UserID, &user.Login, &user.Password)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user, ErrInvalidUsernamePassword
		}

		return user, err
	}

	return user, nil
}

func (p *PostgresqlDatabase) CreateOrder(number, userID string) error {
	log.Print(number)
	log.Print(userID)
	var uploadedUserID string
	err := p.conn.QueryRow(context.Background(), checkUploadOrder, number).Scan(&uploadedUserID)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
	}

	if uploadedUserID != "" {
		if uploadedUserID == userID {
			return ErrOrderWasUploadedBefore
		} else {
			return ErrOrderWasUploadedAnotherUser
		}
	}

	rows, err := p.conn.Query(context.Background(), createOrder, number, userID)
	if err != nil {
		return err
	}
	defer rows.Close()

	return nil
}
