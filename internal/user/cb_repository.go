package user

import (
	"auth-api/internal/models"
	"context"

	"github.com/google/uuid"
)

type UserCouchbaseRepository interface {
	CreateUser(ctx context.Context, user *models.User) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByEmailOrUsername(ctx context.Context, email string, username string) (*models.User, error)
	FindById(ctx context.Context, userID uuid.UUID) (*models.User, error)
}
