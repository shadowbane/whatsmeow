package router

import (
	"github.com/julienschmidt/httprouter"
	"gomeow/cmd/api/controllers"
	"gomeow/pkg/application"
	"gomeow/pkg/middleware"

	messagecontroller "gomeow/cmd/api/controllers/message"
	sessioncontroller "gomeow/cmd/api/controllers/session"
	usercontroller "gomeow/cmd/api/controllers/user"
)

func Get(app *application.Application) *httprouter.Router {
	m := middleware.InitMiddlewareList(app.Models)

	mux := httprouter.New()

	// index
	mux.GET("/api/v1/messages", controllers.MessageIndex(app))

	// show

	// store

	// update

	// delete

	// solo.wablas.com Compatible API
	mux.POST("/api/v2/send-message", m.Chain(controllers.MessageSend(app), "auth", "default"))

	// Users
	// index
	mux.GET("/api/v1/user", m.Chain(usercontroller.Index(app), "auth"))
	// store
	mux.POST("/api/v1/user", m.Chain(usercontroller.Store(app), "auth"))
	// show
	mux.GET("/api/v1/user/:id", m.Chain(usercontroller.Show(app), "auth"))
	// update
	mux.PUT("/api/v1/user/:id", m.Chain(usercontroller.Update(app), "auth"))
	// delete
	mux.DELETE("/api/v1/user/:id", m.Chain(usercontroller.Delete(app), "auth"))

	// Polls
	// index
	// store
	// show
	// update
	// delete

	// Poll Details
	// index
	// store
	// show
	// update
	// delete

	// Sessions
	// connect
	mux.POST("/api/v1/session/connect", m.Chain(sessioncontroller.Connect(app), "auth"))
	// Get QR Code
	mux.GET("/api/v1/session/qr-code", m.Chain(sessioncontroller.GetQRCode(app), "auth"))
	// Disconnect
	mux.POST("/api/v1/session/disconnect", m.Chain(sessioncontroller.Disconnect(app), "auth"))
	// Logout
	mux.POST("/api/v1/session/logout", m.Chain(sessioncontroller.Logout(app), "auth"))

	// Messages
	// index
	mux.GET("/api/v1/message/text", m.Chain(messagecontroller.Index(app), "auth"))
	// send text
	mux.POST("/api/v1/message/text", m.Chain(messagecontroller.SendText(app), "auth"))

	return mux
}
