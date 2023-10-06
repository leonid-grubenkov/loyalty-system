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
	query := `CREATE TABLE IF NOT EXISTS users(login text primary key unique, pass_hash text);
				CREATE TABLE IF NOT EXISTS orders(order_id bigint primary key unique, status text, accrual int, login text, uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);`

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
		var accrual sql.NullInt64

		if err := rows.Scan(&order.Number, &order.Status, &accrual, &order.UploadedAt); err != nil {
			return nil, err
		}
		if accrual.Valid {
			order.Accrual = int(accrual.Int64)
		}

		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (d *Database) selectCounter(name string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var delta int64

	err := d.DB.QueryRowContext(ctx, "SELECT delta FROM metrics WHERE id = $1", name).Scan(&delta)
	if err != nil {
		return 0, err
	}
	return delta, nil
}

func (d *Database) selectGauge(name string) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var value float64

	err := d.DB.QueryRowContext(ctx, "SELECT value FROM metrics WHERE id = $1", name).Scan(&value)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func (d *Database) insertCounter(ctx context.Context, name string, value int64) error {
	query := `
			INSERT INTO metrics(id, type, delta)
			VALUES ($1, $2, $3)
			ON CONFLICT (id) DO UPDATE
			SET delta = metrics.delta + EXCLUDED.delta`

	_, err := d.DB.ExecContext(ctx, query, name, "counter", value)
	if err != nil {
		log.Printf("Error %s when creating or updating counter metrics", err)
		return err
	}
	return nil
}

func (d *Database) insertGauge(ctx context.Context, name string, value float64) error {
	query := `
			INSERT INTO metrics(id, type, value)
			VALUES ($1, $2, $3)
			ON CONFLICT (id) DO UPDATE
			SET value = EXCLUDED.value`

	_, err := d.DB.ExecContext(ctx, query, name, "gauge", value)
	if err != nil {
		log.Printf("Error %s when creating or updating gauge metrics", err)
		return err
	}
	return nil
}
