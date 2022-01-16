package repository

import (
	"auth-api/config"
	"auth-api/internal/models"
	"auth-api/internal/session"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

const (
	basePrefix = "sessions:"
)

// Session repository
type sessionRepo struct {
	redisClient *redis.Client
	basePrefix  string
	cfg         *config.Config
}

func NewSessionRepository(redisClient *redis.Client, cfg *config.Config) session.SessRepository {
	return &sessionRepo{redisClient: redisClient, basePrefix: basePrefix, cfg: cfg}
}

// Create session in redis
func (s *sessionRepo) CreateSession(ctx context.Context, sess *models.Session, expire int) (string, error) {

	sess.SessionID = uuid.New().String()
	sessionKey := s.createKey(sess.SessionID)

	sessBytes, err := json.Marshal(&sess)
	if err != nil {
		return "", errors.WithMessage(err, "sessionRepo.CreateSession.json.Marshal")
	}
	if err = s.redisClient.Set(ctx, sessionKey, sessBytes, time.Second*time.Duration(expire)).Err(); err != nil {
		return "", errors.Wrap(err, "sessionRepo.CreateSession.redisClient.Set")
	}
	return sess.SessionID, nil
}

// Get session by id
func (s *sessionRepo) GetSessionByID(ctx context.Context, sessionID string) (*models.Session, error) {

	sessBytes, err := s.redisClient.Get(ctx, s.createKey(sessionID)).Bytes()
	if err != nil {
		return nil, errors.Wrap(err, "sessionRep.GetSessionByID.redisClient.Get")
	}

	sess := &models.Session{}
	if err = json.Unmarshal(sessBytes, &sess); err != nil {
		return nil, errors.Wrap(err, "sessionRepo.GetSessionByID.json.Unmarshal")
	}
	return sess, nil
}

// Delete session by id
func (s *sessionRepo) DeleteByID(ctx context.Context, sessionID string) error {

	if err := s.redisClient.Del(ctx, s.createKey(sessionID)).Err(); err != nil {
		return errors.Wrap(err, "sessionRepo.DeleteByID")
	}
	return nil
}

func (s *sessionRepo) createKey(sessionID string) string {
	return fmt.Sprintf("%s: %s", s.basePrefix, sessionID)
}
