package services

import "auth-api/config"

type usersService struct {
	logger logger.Logger
	cfg    *config.Config
	userUC user.UserUseCase
	sessUC session.SessionUseCase
}

// Auth service constructor
func NewAuthServerGRPC(logger logger.Logger, cfg *config.Config, userUC user.UserUseCase, sessUC session.SessionUseCase) *usersService {
	return &usersService{logger: logger, cfg: cfg, userUC: userUC, sessUC: sessUC}