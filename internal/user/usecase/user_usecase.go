package usecase

import (
	"auth-api/internal/models"
	"auth-api/internal/user"
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

const (
	userByIdCacheDuration = 3600
)

// User UseCase
type userUseCase struct {
	userCBRepo    user.UserCouchbaseRepository
	userRedisRepo user.UserRedisRepository
}

// New User UseCase
func NewUserUseCase(userRepo user.UserCouchbaseRepository, redisRepo user.UserRedisRepository) *userUseCase {
	return &userUseCase{userCBRepo: userRepo, userRedisRepo: redisRepo}
}

// Register new user
func (u *userUseCase) Register(ctx context.Context, user *models.User) (*models.User, error) {
	existsUser, err := u.userCBRepo.FindByEmailOrUsername(ctx, user.Email, user.Username)
	if existsUser != nil || err == nil {
		return nil, errors.New("email exists")
	}

	return u.userCBRepo.CreateUser(ctx, user)
}

// Find use by email address
func (u *userUseCase) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	findByEmail, err := u.userCBRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("User not found by email")
	}
	fmt.Println(findByEmail)
	findByEmail.SanitizePassword()
	return findByEmail, nil
}

func (u *userUseCase) FindByEmailOrUsername(ctx context.Context, email string, username string) (*models.User, error) {
	findByEmailOrUsername, err := u.userCBRepo.FindByEmailOrUsername(ctx, email, username)
	if err != nil {
		return nil, errors.New("User not found by email or username")
	}
	return findByEmailOrUsername, nil
}

// Find use by uuid
func (u *userUseCase) FindById(ctx context.Context, userID uuid.UUID) (*models.User, error) {

	cachedUser, err := u.userRedisRepo.GetByIDCtx(ctx, userID.String())
	if err != nil && !errors.Is(err, redis.Nil) {
		fmt.Println("redisRepo.GetByIDCtx", err)
	}
	if cachedUser != nil {
		return cachedUser, nil
	}

	foundUser, err := u.userCBRepo.FindById(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, "userPgRepo.FindById")
	}

	if err := u.userRedisRepo.SetUserCtx(ctx, foundUser.UserID.String(), userByIdCacheDuration, foundUser); err != nil {
		fmt.Println("redisRepo.SetUserCtx", err)
	}

	return foundUser, nil
}

// Login user with email and password
func (u *userUseCase) Login(ctx context.Context, email string, password string) (*models.User, error) {
	foundUser, err := u.userCBRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, errors.Wrap(err, "userPgRepo.FindByEmail")
	}

	if err := foundUser.ComparePasswords(password); err != nil {
		return nil, errors.Wrap(err, "user.ComparePasswords")
	}

	return foundUser, err
}
