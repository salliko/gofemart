package databases

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/salliko/gofemart/config"
)

var ErrLoginConfict = errors.New(`login conglict`)
var ErrInvalidUsernamePassword = errors.New(`invalid username/password pair`)

type Database interface {
	CreateUser(string, string, string) (User, error)
	SelectUser(string, string) (User, error)
	HasLogin(string) (bool, error)
	Close()
}

type User struct {
	UserID   string `json:"-"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

type PostgresqlDatabase struct {
	conn *pgxpool.Pool
}

func NewPostgresqlDatabase(cfg config.Config) (*PostgresqlDatabase, error) {
	conn, err := pgxpool.Connect(context.Background(), cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	rows, err := conn.Query(context.Background(), createDatabaseStruct)
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
