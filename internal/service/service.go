package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"log"

	"github.com/leonid-grubenkov/loyalty-system/internal/models"

	"github.com/leonid-grubenkov/loyalty-system/internal/storage"
	"github.com/leonid-grubenkov/loyalty-system/internal/utils"
)

type Service struct {
	db     *storage.Database
	orders chan int
}

const loginKey string = "login"

func NewService(db *storage.Database, orders chan int) *Service {
	return &Service{db: db, orders: orders}
}

func (s *Service) Ping() error {
	err := s.db.Ping()
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) RegisterUser(login, pass string) error {
	hashPass, err := utils.HashPassword(pass)
	if err != nil {
		return err
	}
	err = s.db.RegisterUser(login, hashPass)
	return err
}

func (s *Service) LoginUser(login, pass string) error {
	hashPass, err := s.db.GetHashPass(login)
	if err != nil {
		log.Println(err)
		return err
	}

	if !utils.CheckPasswordHash(pass, hashPass) {
		log.Println("wrong pass")
		return fmt.Errorf("wrong_pass")
	}
	return nil
}

func (s *Service) LoadOrder(ctx context.Context, order int) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	user, err := s.db.GetUserFromOrder(ctx, order)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	ctxLogin := ctx.Value(loginKey)
	if user != "" && user == ctxLogin {
		return fmt.Errorf("200")
	}
	if user != "" && user != ctxLogin {
		return fmt.Errorf("409")
	}

	err = s.db.InsertNewOrder(ctx, order)
	if err != nil {
		log.Println(err)
		return err
	}
	s.orders <- order
	return nil
}

func (s *Service) GetOrders(ctx context.Context) (*[]models.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	userLogin := ctx.Value(loginKey)
	if userLogin == "" {
		return nil, fmt.Errorf("no login")
	}

	orders, err := s.db.GetOrders(ctx, fmt.Sprint(userLogin))
	if err != nil {
		return nil, err
	}

	return &orders, nil
}

func (s *Service) CheckBalance(ctx context.Context) (*models.BalanceInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	userLogin, ok := ctx.Value(loginKey).(string)
	if !ok {
		return nil, fmt.Errorf("userLogin is not a string")
	}

	info, err := s.db.CheckBalance(ctx, userLogin)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return info, nil
}

func (s *Service) NewWithdrawn(ctx context.Context, order int, sum float64) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	userLogin, ok := ctx.Value(loginKey).(string)
	if !ok {
		return fmt.Errorf("userLogin is not a string")
	}

	balance, err := s.db.CheckBalance(ctx, userLogin)
	if err != nil {
		return err
	}

	if balance.Balance < sum {
		return fmt.Errorf("insufficient balance")
	}

	err = s.db.CreateWithdrawn(ctx, order, sum, userLogin)
	if err != nil {
		log.Println("error when create withdrawn: ", err)
		if strings.Contains(err.Error(), "duplicate") {
			return fmt.Errorf("already registred")
		}
		return err
	}

	err = s.db.RemoveBalance(ctx, userLogin, sum)
	if err != nil {
		log.Println("error when remove blance: ", err)
		return err
	}

	return nil
}

func (s *Service) GetWithdrawns(ctx context.Context) (*[]models.Withdraw, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	userLogin := ctx.Value(loginKey)
	if userLogin == "" {
		return nil, fmt.Errorf("no login")
	}

	withdrawans, err := s.db.GetWithdrawns(ctx, fmt.Sprint(userLogin))
	if err != nil {
		return nil, err
	}

	return &withdrawans, nil
}
