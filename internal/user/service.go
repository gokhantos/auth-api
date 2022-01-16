package user

import (
	"auth-api/config"
	"auth-api/internal/session"
)

type usersService struct {
	cfg    *config.Config
	userUC UserUseCase
	sessUC session.SessionUseCase
}

// Auth service constructor
func NewAuthServerGRPC(cfg *config.Config, userUC UserUseCase, sessUC session.SessionUseCase) *usersService {
	return &usersService{cfg: cfg, userUC: userUC, sessUC: sessUC}
}
