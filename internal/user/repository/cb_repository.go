package repository

import (
	"auth-api/internal/models"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/couchbase/gocb/v2"
	"github.com/google/uuid"
)

type UserCouchbaseRepository interface {
	CreateUser(ctx context.Context, user *models.User) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
}

type UserRepository struct {
	cb *gocb.Cluster
}

// User repository constructor
func NewUserCBRepository(db *gocb.Cluster) *UserRepository {
	return &UserRepository{cb: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	bucket := r.cb.Bucket("Users")
	collection := bucket.DefaultCollection()

	_, err := collection.Insert(uuid.New().String(), &user, &gocb.InsertOptions{})
	if err != nil {
		log.Println("User create error")
		return nil, errors.New("user create error")
	}
	return &models.User{}, nil
}

func (r *UserRepository) FindByEmailOrUsername(ctx context.Context, email string, username string) (*models.User, error) {
	results, error := r.cb.Query(fmt.Sprintf("SELECT user_id, email, username, created_at, updated_at FROM Users WHERE email='%s' OR username='%s'", email, username), &gocb.QueryOptions{})
	if error != nil {
		return nil, errors.New("query error")
	}
	var result interface{}
	for results.Next() {
		err := results.Row(&result)
		if err != nil {
			panic(err)
		}
	}
	if result == nil {
		return nil, errors.New("user not found with this email or username")
	}
	var user models.User
	jsonString, _ := json.Marshal(result)
	json.Unmarshal(jsonString, &user)
	return &user, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	results, err := r.cb.Query(fmt.Sprintf("SELECT _password, user_id, email, username, created_at, updated_at FROM Users WHERE email='%s'", email), &gocb.QueryOptions{})
	if err != nil {
		return nil, errors.New("user did not find error")
	}
	var result interface{}
	for results.Next() {
		err := results.Row(&result)
		if err != nil {
			panic(err)
		}
	}
	if result == nil {
		return nil, errors.New("user not found with this email")
	}
	var user models.User
	jsonString, _ := json.Marshal(result)
	json.Unmarshal(jsonString, &user)
	return &user, nil
}

// Find user by uuid
func (r *UserRepository) FindById(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	results, err := r.cb.Query(fmt.Sprintf("SELECT _password, user_id, email, username, created_at, updated_at FROM Users WHERE user_id='%s'", userID), &gocb.QueryOptions{})
	if err != nil {
		return nil, errors.New("user did not find error")
	}
	var result interface{}
	for results.Next() {
		err := results.Row(&result)
		if err != nil {
			panic(err)
		}
	}
	if result == nil {
		return nil, errors.New("user not found with this email")
	}
	var user models.User
	jsonString, _ := json.Marshal(result)
	json.Unmarshal(jsonString, &user)
	return &user, nil
}
