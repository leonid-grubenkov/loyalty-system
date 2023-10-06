package service

import (
	"fmt"

	"log"

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
