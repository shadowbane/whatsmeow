package router

import (
	"github.com/julienschmidt/httprouter"
	"gomeow/pkg/application"
	"gomeow/pkg/middleware"

	devicecontroller "gomeow/cmd/api/controllers/device"
	filecontroller "gomeow/cmd/api/controllers/file"
	imagecontroller "gomeow/cmd/api/controllers/image"
	messagecontroller "gomeow/cmd/api/controllers/message"
	pollcontroller "gomeow/cmd/api/controllers/poll"
	polldetailcontroller "gomeow/cmd/api/controllers/poll_detail"
	pollmessagecontroller "gomeow/cmd/api/controllers/pollmessage"
	sessioncontroller "gomeow/cmd/api/controllers/session"
)

func Api(app *application.Application) *httprouter.Router {
	m := middleware.InitMiddlewareList(app.Models)

	mux := httprouter.New()

	// Devices
	// index
	mux.GET("/api/v1/device", m.Chain(devicecontroller.Index(app), "auth"))
	// store
	mux.POST("/api/v1/device", m.Chain(devicecontroller.Store(app), "auth"))
	// show
	mux.GET("/api/v1/device/:id", m.Chain(devicecontroller.Show(app), "auth"))
	// update
	mux.PUT("/api/v1/device/:id", m.Chain(devicecontroller.Update(app), "auth"))
	// delete
	mux.DELETE("/api/v1/device/:id", m.Chain(devicecontroller.Delete(app), "auth"))

	// Polls
	// index
	mux.GET("/api/v1/poll", m.Chain(pollcontroller.Index(app), "auth"))
	// store
	mux.POST("/api/v1/poll", m.Chain(pollcontroller.Store(app), "auth"))
	// show
	// update
	// delete

	// Poll Details
	// index
	mux.GET("/api/v1/poll/:pollId", m.Chain(polldetailcontroller.Index(app), "auth"))
	// store
	mux.POST("/api/v1/poll/:pollId", m.Chain(polldetailcontroller.Store(app), "auth"))
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

	// Messages - Text
	// index
	mux.GET("/api/v1/message/text", m.Chain(messagecontroller.Index(app), "auth"))
	// send text
	mux.POST("/api/v1/message/text", m.Chain(messagecontroller.SendText(app), "auth"))

	// Messages - Poll
	// index
	mux.GET("/api/v1/message/poll", m.Chain(pollmessagecontroller.Index(app), "auth"))
	// send poll
	mux.POST("/api/v1/message/poll", m.Chain(pollmessagecontroller.SendPoll(app), "auth"))

	// Messages - Image
	// send image
	mux.POST("/api/v1/message/image", m.Chain(imagecontroller.SendImage(app), "auth"))

	// Messages - File
	// send image
	mux.POST("/api/v1/message/file", m.Chain(filecontroller.SendFile(app), "auth"))

	return mux
}
