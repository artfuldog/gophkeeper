package server

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/artfuldog/gophkeeper/internal/logger"
	"github.com/artfuldog/gophkeeper/internal/pb"
	"github.com/artfuldog/gophkeeper/internal/server/authorizer"
	"github.com/artfuldog/gophkeeper/internal/server/db"
	"github.com/artfuldog/gophkeeper/internal/server/grpcapi"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// Server represents server structure.
type Server struct {
	grpcServer *grpc.Server // gRPC server
	Address    string
	DB         db.DB
	Logger     logger.L
	LogLevel   logger.Level
}

//  NewSrv is a constructor used to initialize server and set up all parameters and components.
func NewServer(cfg *Config) (*Server, error) {
	s := new(Server)

	s.Address = cfg.Address

	var err error
	s.LogLevel, err = logger.GetLevelFromString(cfg.LogLevel)
	if err != nil {
		return nil, err
	}

	// Logger initialization
	s.Logger, err = logger.NewZLoggerConsole(s.LogLevel, "gk_server", logger.OutputStdoutPretty)
	if err != nil {
		return nil, err
	}

	if err := s.createDB(cfg); err != nil {
		return nil, err
	}
	if err := s.createGRPCServer(cfg); err != nil {
		return nil, err
	}

	return s, nil
}

// Run starts server and all it's components.
func (s *Server) Run(ctx context.Context, statusChan chan error) {
	componentName := "run"

	if err := s.DB.ConnectAndSetup(ctx); err != nil {
		statusChan <- err
		return
	}

	dbControlCh := make(db.CloseChannel)
	dbCtx, dbCancel := context.WithCancel(context.Background())
	go s.DB.Run(dbCtx, dbControlCh)
	defer dbCancel()

	grpcControlCh := make(db.CloseChannel)
	grpcCtx, grpcCancel := context.WithCancel(context.Background())
	go s.startGRPCServer(grpcCtx, grpcControlCh)
	defer grpcCancel()

	s.Logger.Info("server is started", componentName)

	select {
	case <-ctx.Done():
	case <-grpcControlCh:
		statusChan <- errors.New("GRPC-server unexpected error")
	case <-dbControlCh:
		statusChan <- errors.New("database unexpected error")
	}

	grpcCancel()
	<-grpcControlCh

	dbCancel()
	<-dbControlCh

	s.Logger.Info("server is stopped", componentName)
	close(statusChan)
}

// createAuthorizer is a helper function for initialization and configuration authorizer
func (s *Server) createAuthorizer(cfg *Config) (a authorizer.A, err error) {
	var authLogger logger.L
	if authLogger, err = logger.NewZLoggerConsole(s.LogLevel, "authenticator",
		logger.OutputStdoutPretty); err != nil {
		return
	}

	if a, err = authorizer.New(authorizer.TypePaseto, cfg.ServerKey,
		time.Duration(cfg.TokenValidPeriod)*time.Second, authLogger); err != nil {
		return
	}

	return
}

// createDB is a helper function for initialization and configuration database.
func (s *Server) createDB(cfg *Config) (err error) {
	var dbLogger logger.L
	if dbLogger, err = logger.NewZLoggerConsole(s.LogLevel, "db", logger.OutputStdoutPretty); err != nil {
		return
	}

	dbParams := db.NewDBParameters(cfg.DBDSN, cfg.DBUser, cfg.DBPassword, cfg.MaxSecretSize)

	if s.DB, err = db.New(cfg.DBType, dbParams, dbLogger); err != nil {
		return
	}

	return
}

// createGRPCServer is a helper function for initialization and configuration gRPC server.
func (s *Server) createGRPCServer(cfg *Config) (err error) {
	authorizer, err := s.createAuthorizer(cfg)
	if err != nil {
		return
	}

	grpcUnaryInterceptors := []grpc.UnaryServerInterceptor{
		grpcapi.IsAuthorized(authorizer),
		grpc_recovery.UnaryServerInterceptor(),
	}

	creds, err := s.getGRPCCredentials(cfg)
	if err != nil {
		s.Logger.Warn(err, "TLS error", "Server:createGRPCServer")
		return err
	}

	s.grpcServer = grpc.NewServer(
		grpc.Creds(creds),
		grpc.ChainUnaryInterceptor(grpcUnaryInterceptors...))

	var grpcLogger logger.L
	if grpcLogger, err = logger.NewZLoggerConsole(s.LogLevel, "grpc", logger.OutputStdoutPretty); err != nil {
		return
	}

	itemsService := grpcapi.NewItemsService(s.DB, grpcLogger)
	pb.RegisterItemsServer(s.grpcServer, itemsService)

	usersService := grpcapi.NewUsersService(s.DB, grpcLogger, authorizer)
	pb.RegisterUsersServer(s.grpcServer, usersService)

	return
}

// getGRPCCredentials is a helper function used to configure transport credentials for
// GRPC-server.
func (s *Server) getGRPCCredentials(cfg *Config) (credentials.TransportCredentials, error) {
	if cfg.TLSDisable {
		s.Logger.Warn(nil, "TLS is disabled, all comunications between client and server are insecured", "Server:getGRPCCredentials")
		return insecure.NewCredentials(), nil
	}
	creds, err := credentials.NewServerTLSFromFile(cfg.TLSCertFilepath, cfg.TLSKeyFilepath)
	if err != nil {
		return nil, err
	}
	return creds, nil
}

// startGRPCServer starts gRCP server.
//
// controlCh is a channel for control gRPC server status, received signal means that
// server is stopped.
func (s *Server) startGRPCServer(ctx context.Context, controlCh chan struct{}) {
	componentName := "startgrpc"

	grpcListener, err := net.Listen("tcp", s.Address)
	if err != nil {
		s.Logger.Error(err, "", componentName)
		close(controlCh)
		return
	}

	serverStatusCh := make(chan error)
	go func() {
		if err := s.grpcServer.Serve(grpcListener); err != nil {
			serverStatusCh <- err
		}
	}()
	s.Logger.Info("GRCP is running", componentName)

	select {
	case <-ctx.Done():
		s.grpcServer.GracefulStop()

		s.Logger.Info("GRCP is stopped", componentName)
		close(controlCh)
	case err := <-serverStatusCh:
		s.Logger.Error(err, "", componentName)
		close(controlCh)
		return
	}
}
