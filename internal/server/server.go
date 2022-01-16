package server

import (
	"auth-api/config"
	sessRepoRedis "auth-api/internal/session/repository"
	sessionUseCase "auth-api/internal/session/usecase"
	"auth-api/internal/user"
	userRepo "auth-api/internal/user/repository"
	userUseCase "auth-api/internal/user/usecase"
	userService "auth-api/proto"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/couchbase/gocb/v2"
	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

// GRPC Auth Server
type Server struct {
	cfg         *config.Config
	db          *gocb.Cluster
	redisClient *redis.Client
}

func NewAuthServer(cfg *config.Config, db *gocb.Cluster, redisClient *redis.Client) *Server {
	return &Server{cfg: cfg, db: db, redisClient: redisClient}
}

func (s *Server) Run() error {

	userRepoCB := userRepo.NewUserCBRepository(s.db)
	userRepoRedis := userRepo.NewUserRedisRepo(s.redisClient)
	sessRepo := sessRepoRedis.NewSessionRepository(s.redisClient, s.cfg)
	userUC := userUseCase.NewUserUseCase(userRepoCB, userRepoRedis)
	sessUC := sessionUseCase.NewSessionUseCase(sessRepo, s.cfg)

	l, err := net.Listen("tcp", s.cfg.Server.Port)
	if err != nil {
		return err
	}
	defer l.Close()

	server := grpc.NewServer(grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle: s.cfg.Server.MaxConnectionIdle * time.Minute,
		Timeout:           s.cfg.Server.Timeout * time.Second,
		MaxConnectionAge:  s.cfg.Server.MaxConnectionAge * time.Minute,
		Time:              s.cfg.Server.Timeout * time.Minute,
	}),
	)

	if s.cfg.Server.Mode != "Production" {
		reflection.Register(server)
	}
	authGRPCServer := user.NewAuthServerGRPC(s.cfg, userUC, sessUC)
	userService.RegisterUserServiceServer(server, authGRPCServer)

	go func() {
		if err := server.Serve(l); err != nil {
			log.Fatal(err)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	server.GracefulStop()
	fmt.Println("Server Exited Properly")
	return nil
}
