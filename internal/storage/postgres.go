package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/leonid-grubenkov/loyalty-system/internal/models"
)

type Database struct {
	DB *sql.DB
}

func GetDB(dsn string) *Database {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		fmt.Println(err)
	}
	database := &Database{DB: db}

	err = database.Ping()
	if err != nil {
		fmt.Println(err)
	}

	err = database.createTables()
	if err != nil {
		fmt.Println(err)
	}
	return database
}

func (d *Database) createTables() error {
	query := `CREATE TABLE IF NOT EXISTS users(login text primary key unique, pass_hash text, balance DOUBLE PRECISION, withdrawn DOUBLE PRECISION);
				CREATE TABLE IF NOT EXISTS orders(order_id bigint primary key unique, status text, accrual DOUBLE PRECISION, login text, uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := d.DB.ExecContext(ctx, query)
	if err != nil {
		log.Printf("Error %s when creating users table", err)
		return err
	}

	log.Println("table created")
	return nil
}

func (d *Database) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := d.DB.PingContext(ctx); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (d *Database) RegisterUser(login, hashPass string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	query := `
			INSERT INTO users(login, pass_hash)
			VALUES ($1, $2)
			ON CONFLICT (login) DO NOTHING`

	res, err := d.DB.ExecContext(ctx, query, login, hashPass)
	if err != nil {
		log.Printf("Error %s when creating or updating counter metrics", err)
		return err
	}
	rows, err := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("user_already_exist")
	}
	return nil
}

func (d *Database) GetHashPass(login string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var hashPass string

	err := d.DB.QueryRowContext(ctx, "SELECT pass_hash FROM users WHERE login = $1", login).Scan(&hashPass)
	if err != nil {
		return "", err
	}
	return hashPass, nil
}

func (d *Database) GetUserFromOrder(ctx context.Context, order int) (string, error) {
	var user string
	err := d.DB.QueryRowContext(ctx, "SELECT login FROM orders WHERE order_id = $1", order).Scan(&user)
	if err != nil {
		return "", err
	}
	return user, nil
}

func (d *Database) InsertNewOrder(ctx context.Context, order int) error {
	query := `
			INSERT INTO orders(order_id, status, login)
			VALUES ($1, $2, $3)`

	user := ctx.Value("login")
	_, err := d.DB.ExecContext(ctx, query, order, "NEW", user)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) GetOrders(ctx context.Context, login string) ([]models.Order, error) {
	var orders []models.Order

	rows, err := d.DB.QueryContext(ctx, "SELECT order_id, status, accrual, uploaded_at FROM orders WHERE login = $1 ORDER BY uploaded_at DESC", login)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var order models.Order
		var accrual sql.NullFloat64

		if err := rows.Scan(&order.Number, &order.Status, &accrual, &order.UploadedAt); err != nil {
			return nil, err
		}
		if accrual.Valid {
			order.Accrual = accrual.Float64
		}

		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (d *Database) CheckBalance(ctx context.Context) (*models.BalanceInfo, error) {
	var balance sql.NullFloat64
	var withdrawn sql.NullFloat64
	var info models.BalanceInfo

	usetLogin := ctx.Value("login")
	err := d.DB.QueryRowContext(ctx, "SELECT balance, withdrawn FROM users WHERE login = $1", usetLogin).Scan(&balance, &withdrawn)
	if err != nil {
		return nil, err
	}
	if balance.Valid {
		info.Balance = balance.Float64
	}
	if withdrawn.Valid {
		info.Withdrawn = withdrawn.Float64
	}
	return &info, nil
}

func (d *Database) ChangeStatus(ctx context.Context, order int, status string) error {
	query := `
		UPDATE orders
		SET status = $2
		WHERE order_id = $1`

	_, err := d.DB.ExecContext(ctx, query, order, status)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) ChangeAccrual(ctx context.Context, order int, status string, accrual float64) error {
	query := `
		UPDATE orders
		SET status = $2, accrual = $3
		WHERE order_id = $1`

	_, err := d.DB.ExecContext(ctx, query, order, status, accrual)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) AddBallance(ctx context.Context, login string, accrual float64) error {
	query := `
	UPDATE users
	SET balance = balance + $2
	WHERE login = $1`

	_, err := d.DB.ExecContext(ctx, query, login, float64(accrual))
	if err != nil {
		return err
	}
	return nil
}
