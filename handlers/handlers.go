package handlers

import (
	"auth-api/models"
	userService "auth-api/proto"
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Register new user
func (u *usersService) Register(ctx context.Context, r *userService.RegisterRequest) (*userService.RegisterResponse, error) {

	user, err := u.registerReqToUserModel(r)
	if err != nil {
		u.logger.Errorf("registerReqToUserModel: %v", err)
		return nil, status.Errorf(grpc_errors.ParseGRPCErrStatusCode(err), "registerReqToUserModel: %v", err)
	}

	if err := utils.ValidateStruct(ctx, user); err != nil {
		u.logger.Errorf("ValidateStruct: %v", err)
		return nil, status.Errorf(grpc_errors.ParseGRPCErrStatusCode(err), "ValidateStruct: %v", err)
	}

	createdUser, err := u.userUC.Register(ctx, user)
	if err != nil {
		u.logger.Errorf("userUC.Register: %v", err)
		return nil, status.Errorf(grpc_errors.ParseGRPCErrStatusCode(err), "Register: %v", err)
	}

	return &userService.RegisterResponse{User: u.userModelToProto(createdUser)}, nil
}

// Login user with email and password
func (u *userService) Login(ctx context.Context, r *userService.LoginRequest) (*userService.LoginResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "user.Create")
	defer span.Finish()

	email := r.GetEmail()
	if !utils.ValidateEmail(email) {
		u.logger.Errorf("ValidateEmail: %v", email)
		return nil, status.Errorf(codes.InvalidArgument, "ValidateEmail: %v", email)
	}

	user, err := u.userUC.Login(ctx, email, r.GetPassword())
	if err != nil {
		u.logger.Errorf("userUC.Login: %v", err)
		return nil, status.Errorf(grpc_errors.ParseGRPCErrStatusCode(err), "Login: %v", err)
	}

	session, err := u.sessUC.CreateSession(ctx, &models.Session{
		UserID: user.UserID,
	}, u.cfg.Session.Expire)
	if err != nil {
		u.logger.Errorf("sessUC.CreateSession: %v", err)
		return nil, status.Errorf(grpc_errors.ParseGRPCErrStatusCode(err), "sessUC.CreateSession: %v", err)
	}

	return &userService.LoginResponse{User: u.userModelToProto(user), SessionId: session}, err
}
