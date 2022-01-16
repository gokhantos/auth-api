package user

import (
	"auth-api/internal/models"
	userService "auth-api/proto"
	"auth-api/utils"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (u *usersService) Register(ctx context.Context, r *userService.RegisterRequest) (*userService.RegisterResponse, error) {

	user, err := u.registerReqToUserModel(r)
	if err != nil {
		fmt.Printf("registerReqToUserModel: %v", err)
		return nil, errors.New("register request to user model error")
	}

	if err := utils.ValidateStruct(ctx, user); err != nil {
		fmt.Printf("ValidateStruct: %v", err)
		return nil, errors.New("struct validation error")
	}

	createdUser, err := u.userUC.Register(ctx, user)
	if err != nil {
		fmt.Printf("userUC.Register: %v", err)
		return nil, errors.New("user create error")
	}

	return &userService.RegisterResponse{User: u.userModelToProto(createdUser)}, nil
}

func (u *usersService) FindByEmail(ctx context.Context, r *userService.FindByEmailRequest) (*userService.FindByEmailResponse, error) {
	email := r.GetEmail()
	if !utils.ValidateEmail(email) {
		fmt.Printf("ValidateEmail: %v", email)
		return nil, status.Errorf(codes.InvalidArgument, "ValidateEmail: %v", email)
	}

	user, err := u.userUC.FindByEmail(ctx, email)
	if err != nil {
		fmt.Printf("userUC.FindByEmail: %v", err)
		return nil, status.Errorf(utils.ParseGRPCErrStatusCode(err), "userUC.FindByEmail: %v", err)
	}
	return &userService.FindByEmailResponse{User: u.userModelToProto(user)}, err
}

func (u *usersService) FindByEmailOrUsername(ctx context.Context, r *userService.FindByEmailOrUsernameRequest) (*userService.FindByEmailOrUsernameResponse, error) {
	email := r.GetEmail()
	username := r.GetUsername()
	if !utils.ValidateEmail(email) {
		fmt.Printf("ValidateEmail: %v", email)
		return nil, status.Errorf(codes.InvalidArgument, "ValidateEmail: %v", email)
	}
	user, err := u.userUC.FindByEmailOrUsername(ctx, email, username)
	if err != nil {
		fmt.Printf("userUC.FindByEmailOrUsername: %v", err)
		return nil, status.Errorf(utils.ParseGRPCErrStatusCode(err), "userUC.FindByEmailOrUsername: %v", err)
	}
	return &userService.FindByEmailOrUsernameResponse{User: u.userModelToProto(user)}, err
}

// Find user by uuid
func (u *usersService) FindByID(ctx context.Context, r *userService.FindByIDRequest) (*userService.FindByIDResponse, error) {

	userUUID, err := uuid.Parse(r.GetUuid())
	if err != nil {
		fmt.Printf("uuid.Parse: %v", err)
		return nil, status.Errorf(utils.ParseGRPCErrStatusCode(err), "uuid.Parse: %v", err)
	}

	user, err := u.userUC.FindById(ctx, userUUID)
	if err != nil {
		fmt.Printf("userUC.FindById: %v", err)
		return nil, status.Errorf(utils.ParseGRPCErrStatusCode(err), "userUC.FindById: %v", err)
	}

	return &userService.FindByIDResponse{User: u.userModelToProto(user)}, nil
}

// Get session id from, ctx metadata, find user by uuid and returns it
func (u *usersService) GetMe(ctx context.Context, r *userService.GetMeRequest) (*userService.GetMeResponse, error) {

	sessID, err := u.getSessionIDFromCtx(ctx)
	if err != nil {
		fmt.Printf("getSessionIDFromCtx: %v", err)
		return nil, status.Errorf(utils.ParseGRPCErrStatusCode(err), "sessUC.getSessionIDFromCtx: %v", err)
	}

	session, err := u.sessUC.GetSessionByID(ctx, sessID)
	if err != nil {
		fmt.Printf("sessUC.GetSessionByID: %v", err)
		if errors.Is(err, redis.Nil) {
			return nil, status.Errorf(codes.NotFound, "sessUC.GetSessionByID: %v", utils.ErrNotFound)
		}
		return nil, status.Errorf(utils.ParseGRPCErrStatusCode(err), "sessUC.GetSessionByID: %v", err)
	}

	user, err := u.userUC.FindById(ctx, session.UserID)
	if err != nil {
		fmt.Printf("userUC.FindById: %v", err)
		return nil, status.Errorf(utils.ParseGRPCErrStatusCode(err), "userUC.FindById: %v", err)
	}

	return &userService.GetMeResponse{User: u.userModelToProto(user)}, nil
}

// Logout user, delete current session
func (u *usersService) Logout(ctx context.Context, request *userService.LogoutRequest) (*userService.LogoutResponse, error) {

	sessID, err := u.getSessionIDFromCtx(ctx)
	if err != nil {
		fmt.Printf("getSessionIDFromCtx: %v", err)
		return nil, err
	}

	if err := u.sessUC.DeleteByID(ctx, sessID); err != nil {
		fmt.Printf("sessUC.DeleteByID: %v", err)
		return nil, status.Errorf(utils.ParseGRPCErrStatusCode(err), "sessUC.DeleteByID: %v", err)
	}

	return &userService.LogoutResponse{}, nil
}

// Login user with email and password
func (u *usersService) Login(ctx context.Context, r *userService.LoginRequest) (*userService.LoginResponse, error) {
	email := r.GetEmail()
	if !utils.ValidateEmail(email) {
		fmt.Printf("ValidateEmail: %v", email)
		return nil, status.Errorf(codes.InvalidArgument, "ValidateEmail: %v", email)
	}

	user, err := u.userUC.Login(ctx, email, r.GetPassword())
	if err != nil {
		fmt.Printf("userUC.Login: %v", err)
		return nil, status.Errorf(utils.ParseGRPCErrStatusCode(err), "Login: %v", err)
	}

	session, err := u.sessUC.CreateSession(ctx, &models.Session{
		UserID: user.UserID,
	}, u.cfg.Session.Expire)
	if err != nil {
		fmt.Printf("sessUC.CreateSession: %v", err)
		return nil, status.Errorf(utils.ParseGRPCErrStatusCode(err), "sessUC.CreateSession: %v", err)
	}

	return &userService.LoginResponse{User: u.userModelToProto(user), SessionId: session}, err
}

func (u *usersService) registerReqToUserModel(r *userService.RegisterRequest) (*models.User, error) {
	candidate := &models.User{
		UserID:    uuid.New(),
		Email:     r.GetEmail(),
		Username:  r.GetUsername(),
		Password:  r.GetPassword(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := candidate.PrepareCreate(); err != nil {
		return nil, err
	}

	return candidate, nil
}

func (u *usersService) userModelToProto(user *models.User) *userService.User {
	userProto := &userService.User{
		Uuid:      user.UserID.String(),
		Username:  user.Username,
		Password:  user.Password,
		Email:     user.Email,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}
	return userProto
}

func (u *usersService) getSessionIDFromCtx(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Errorf(codes.Unauthenticated, "metadata.FromIncomingContext: %v", utils.ErrNoCtxMetaData)
	}

	sessionID := md.Get("session_id")
	if sessionID[0] == "" {
		return "", status.Errorf(codes.PermissionDenied, "md.Get sessionId: %v", utils.ErrInvalidSessionId)
	}

	return sessionID[0], nil
}
