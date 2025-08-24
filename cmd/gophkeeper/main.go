package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net"

	"github.com/PiskarevSA/goph-keeper/config"
	gophkeeperv1 "github.com/PiskarevSA/goph-keeper/gen/gophkeeper/v1"
	"github.com/PiskarevSA/goph-keeper/internal/app/gophkeeper"
	"github.com/PiskarevSA/goph-keeper/internal/service"
	"github.com/PiskarevSA/goph-keeper/internal/storage"
	"google.golang.org/grpc"

	_ "github.com/lib/pq"
)

func main() {
	log := slog.Default()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Error("failed to load config", "err", err)
		return
	}
	log.Info("loaded config", "config", cfg.SafeString())

	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		log.Error("db connect failed", "err", err)
		return
	}
	defer db.Close()

	store := storage.NewPostgresStorage(db)
	userSvc := service.NewUserService(store)
	secretSvc := service.NewSecretsService(store)

	grpcServer := grpc.NewServer()
	gophkeeperv1.RegisterUserServiceServer(
		grpcServer, gophkeeper.NewUserServer(userSvc))
	gophkeeperv1.RegisterSecretsServiceServer(
		grpcServer, gophkeeper.NewSecretsServer(secretSvc))

	addr := net.JoinHostPort(
		cfg.Server.Host,
		fmt.Sprintf("%d", cfg.Server.Port))
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Error("listen failed", "err", err)
		return
	}
	log.Info("server started", "addr", addr)

	if err := grpcServer.Serve(listener); err != nil {
		log.Error("serve failed", "err", err)
	}
}
