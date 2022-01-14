package models

import (
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"time"
)

//User model
type User struct {
	UserID    uuid.UUID `json:"user_id" db:"user_id" validate:"omitempty"`
	Email     string    `json:"email" db:"email" validate:"omitempty,lte=60,email"`
	Username  string    `json:"username" db:"username" validate:"required,lte=30"`
	Avatar    *string   `json:"avatar" db:"avatar"`
	Password  string    `json:"password,omitempty" db:"password"`
	CreatedAt time.Time `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty" db:"updated_at"`
}

// Sanitize password
func (u *User) SanitizePassword() {
	u.Password = ""
}

// Hash user password with bcrypt
func (u *User) HashPassword() error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// Compare user password and payload
func (u *User) ComparePasswords(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}

// Prepare user for register
func (u *User) PrepareCreate() error {
	u.Email = strings.ToLower(strings.TrimSpace(u.Email))
	u.Password = strings.TrimSpace(u.Password)

	if err := u.HashPassword(); err != nil {
		return err
	}

	return nil
}

// Get avatar string
func (u *User) GetAvatar() string {
	if u.Avatar == nil {
		return ""
	}
	return *u.Avatar
}
