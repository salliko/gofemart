package databases

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/salliko/gofemart/config"
)

var ErrLoginConfict = errors.New(`login conglict`)
var ErrInvalidUsernamePassword = errors.New(`invalid username/password pair`)
var ErrOrderWasUploadedBefore = errors.New(`номер заказа уже был загружен этим пользователем`)
var ErrOrderWasUploadedAnotherUser = errors.New(`номер заказа уже был загружен другим пользователем`)
var ErrNotFoundOrders = errors.New(`нет данных для ответа`)
var ErrInsufficientFunds = errors.New(`на счету недостаточно средств`)
var ErrInvalidOrderNumber = errors.New(`неверный номер заказа`)
var ErrNotFoundOperations = errors.New(`нет ни одного списания`)

type Database interface {
	CreateUser(string, string, string) (User, error)
	SelectUser(string, string) (User, error)
	HasLogin(string) (bool, error)
	CreateOrder(string, string) error
	SelectOrders(string) ([]Order, error)
	UpdateOrder(string, AccuralOrder) error
	SelectUserBalance(string) (Balance, error)
	CreateDebit(string, Withdrawn) error
	SelectUserOperations(string) ([]Withdrawn, error)
	Close()
}

type Withdrawn struct {
	Order      string    `json:"order"`
	Sum        float64   `json:"sum"`
	UploadedAt time.Time `json:"uploaded_at" format:"RFC3339"`
}

type User struct {
	UserID   string `json:"-"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Order struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accural    float64   `json:"accural,omitempty"`
	UploadedAt time.Time `json:"uploaded_at" format:"RFC3339"`
}

type AccuralOrder struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accural float64 `json:"accrual,omitempty"`
}

type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
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

	rows, err = conn.Query(context.Background(), createTableUserBalances)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rows, err = conn.Query(context.Background(), createTableOperations)
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

	rows, err := p.conn.Query(context.Background(), createDefaultUserBalance, userID)
	if err != nil {
		return user, err
	}
	defer rows.Close()

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

func (p *PostgresqlDatabase) SelectOrders(userID string) ([]Order, error) {
	var orders []Order

	rows, err := p.conn.Query(context.Background(), selectOrders, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var order Order
		err := rows.Scan(&order.Number, &order.Status, &order.Accural, &order.UploadedAt)
		if err != nil {
			return nil, err
		}

		orders = append(orders, order)
	}

	if len(orders) == 0 {
		return nil, ErrNotFoundOrders
	}

	return orders, nil
}

func (p *PostgresqlDatabase) UpdateOrder(userID string, order AccuralOrder) error {
	rows, err := p.conn.Query(context.Background(), updateOrder, order.Status, order.Accural, order.Order)
	if err != nil {
		return err
	}
	defer rows.Close()

	var balance float64
	err = p.conn.QueryRow(context.Background(), selectUserOnlyBalance, userID).Scan(&balance)

	if err != nil {
		return err
	}

	credit := balance + float64(order.Accural)

	rows, err = p.conn.Query(context.Background(), updateUserBalance, credit, userID)
	if err != nil {
		return err
	}
	defer rows.Close()

	return nil
}

func (p *PostgresqlDatabase) SelectUserBalance(userID string) (Balance, error) {
	var balance Balance
	err := p.conn.QueryRow(context.Background(), selectUserBalance, userID).Scan(&balance.Current, &balance.Withdrawn)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return balance, err
		}
	}

	return balance, nil
}

func (p *PostgresqlDatabase) CreateDebit(userID string, withdrawn Withdrawn) error {
	var numberOrder string
	var balance float64
	err := p.conn.QueryRow(
		context.Background(),
		selectUserBalanceAndOrder,
		userID,
		withdrawn.Order).Scan(&balance, &numberOrder)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInvalidOrderNumber
		} else {
			return err
		}
	}

	debit := balance - withdrawn.Sum

	if debit < 0 {
		return ErrInsufficientFunds
	}

	rows, err := p.conn.Query(context.Background(), updateUserBalance, debit, userID)
	if err != nil {
		return err
	}
	defer rows.Close()

	rows, err = p.conn.Query(context.Background(), insertOperation, userID, numberOrder, withdrawn.Sum)
	if err != nil {
		return err
	}
	defer rows.Close()

	return nil
}

func (p *PostgresqlDatabase) SelectUserOperations(userID string) ([]Withdrawn, error) {
	var withdraws []Withdrawn

	rows, err := p.conn.Query(context.Background(), selectUserOperations, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var withdraw Withdrawn
		err := rows.Scan(&withdraw.Order, &withdraw.Sum, &withdraw.UploadedAt)
		if err != nil {
			return nil, err
		}

		withdraws = append(withdraws, withdraw)
	}

	if len(withdraws) == 0 {
		return nil, ErrNotFoundOperations
	}

	return withdraws, nil
}
