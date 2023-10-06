package service

import (
	"context"
	"fmt"
	"time"

	"log"

	"github.com/leonid-grubenkov/loyalty-system/internal/models"
	"github.com/leonid-grubenkov/loyalty-system/internal/storage"
	"github.com/leonid-grubenkov/loyalty-system/internal/utils"
)

type Service struct {
	db *storage.Database
}

func NewService(db *storage.Database) *Service {
	return &Service{db: db}
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

	ctxLogin := ctx.Value("login")
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

	return nil
}

func (s *Service) GetOrders(ctx context.Context) (*[]models.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	userLogin := ctx.Value("login")
	if userLogin == "" {
		return nil, fmt.Errorf("no login")
	}

	orders, err := s.db.GetOrders(ctx, fmt.Sprint(userLogin))
	if err != nil {
		return nil, err
	}

	return &orders, nil
}
