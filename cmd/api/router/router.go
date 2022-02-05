package router

import (
	"github.com/julienschmidt/httprouter"
	"gomeow/cmd/api/controllers"
	"gomeow/pkg/application"
)

func Get(app *application.Application) *httprouter.Router {
	mux := httprouter.New()

	// index
	mux.GET("/api/v1/messages", controllers.MessageIndex(app))

	// show

	// store

	// update

	// delete

	// solo.wablas.com Compatible API
	mux.POST("/api/v2/send-message", controllers.MessageSend(app))

	return mux
}
