package router

import (
	"gomeow/cmd/api/rpc"
	"gomeow/cmd/common/dto"
	"gomeow/pkg/application"
	"google.golang.org/grpc"
)

func Grpc(srv *grpc.Server, app *application.Application) {
	dto.RegisterSessionServiceServer(srv, rpc.SessionService{
		App: app,
	})
}
