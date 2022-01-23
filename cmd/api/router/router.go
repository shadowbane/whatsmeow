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

	return mux
}
