package server

import (
	"errors"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
	"log"
	"net/http"
)

type Server struct {
	srv *http.Server
}

type fwdToZapWriter struct {
	logger *zap.SugaredLogger
}

func (fw *fwdToZapWriter) Write(p []byte) (n int, err error) {
	fw.logger.Errorw(string(p))
	return len(p), nil
}

func Get() *Server {
	return &Server{
		srv: &http.Server{},
	}
}

func (s *Server) WithAddr(addr string) *Server {
	s.srv.Addr = addr
	return s
}

func (s *Server) WithErrLogger(l *zap.SugaredLogger) *Server {
	s.srv.ErrorLog = log.New(&fwdToZapWriter{l}, "", 0)
	return s
}

func (s *Server) WithRouter(router *httprouter.Router) *Server {
	s.srv.Handler = router
	return s
}

func (s *Server) Start() error {
	if len(s.srv.Addr) == 0 {
		return errors.New("server missing address")
	}

	if s.srv.Handler == nil {
		return errors.New("server missing handler")
	}

	return s.srv.ListenAndServe()
}

func (s *Server) Close() error {
	return s.srv.Close()
}
