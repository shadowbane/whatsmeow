package grpc

import (
	"go.uber.org/zap"
	"gomeow/cmd/api/router"
	"gomeow/pkg/application"
	"google.golang.org/grpc"
	"net"
)

type Server struct {
	srv      *grpc.Server
	listener net.Listener
	addr     string
}

type FwdToZapWriter struct {
	Logger *zap.SugaredLogger
}

func (fw *FwdToZapWriter) Write(p []byte) (n int, err error) {
	fw.Logger.Errorw(string(p))
	return len(p), nil
}

func Get() *Server {
	return &Server{
		srv: grpc.NewServer(),
	}
}

func (s *Server) WithAddr(addr string) *Server {
	s.addr = addr
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		zap.S().Fatalf("Cannot listen on %s", addr)
	}

	s.listener = listener

	return s
}

func (s *Server) WithErrLogger(l *zap.SugaredLogger) *Server {
	// ToDo: Add zap logger

	return s
}

func (s *Server) loadRoutes(srv *grpc.Server, app *application.Application) {
	router.Grpc(srv, app)
}

func (s *Server) Start(app *application.Application) error {
	s.loadRoutes(s.srv, app)

	return s.srv.Serve(s.listener)
}

func (s *Server) Close() error {
	s.srv.GracefulStop()

	return nil
}
