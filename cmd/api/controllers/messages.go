package controllers

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"go.mau.fi/whatsmeow"
	"go.uber.org/zap"
	"gomeow/cmd/models"
	"gomeow/pkg/application"
	"io"
	"log"
	"net/http"
	"time"
)

type JsonErrorResponse struct {
	Error *ApiError `json:"error"`
}

type ApiError struct {
	Status int    `json:"status"`
	Title  string `json:"title"`
}

type returnData struct {
	Status  bool                       `json:"status"`
	Message string                     `json:"message"`
	Data    application.PendingMessage `json:"data"`
}

func MessageIndex(app *application.Application) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Fatal(err)
			}
		}(r.Body)

		to := r.URL.Query().Get("destination")
		message := r.URL.Query().Get("message")
		newMessageId := whatsmeow.GenerateMessageID()

		if (len(to) == 0) || (len(message) == 0) {
			writeErrorResponse(w, 422, "Invalid Parameter Supplied")
			return
		}

		zap.S().Debugf("Queueing message with ID: %s and content: %s to %s", newMessageId, message, to)

		pendingMessage := application.PendingMessage{
			To:        to,
			MessageId: newMessageId,
			Message:   message,
		}

		// store to message store
		storeToMessageStore(
			app,
			app.Meow.DeviceStore.ID.String(),
			to,
			message,
			newMessageId,
		)

		// add to queue
		app.Queue.Add(pendingMessage)

		formattedValues := returnData{
			Status:  true,
			Message: "Message queued",
			Data:    pendingMessage,
		}

		response, _ := json.Marshal(formattedValues)
		_, err := w.Write(response)
		if err != nil {
			zap.S().Errorf(err.Error())
			writeErrorResponse(w, http.StatusInternalServerError, "Internal Server Error")
		}
	}
}

// Writes the error response as a Standard API JSON response with a response code
func writeErrorResponse(w http.ResponseWriter, errorCode int, errorMsg string) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(errorCode)
	err := json.
		NewEncoder(w).Encode(&JsonErrorResponse{Error: &ApiError{Status: errorCode, Title: errorMsg}})

	if err != nil {
		zap.S().Fatalf(err.Error())
	}
}

func storeToMessageStore(app *application.Application, jid string, destination string, body string, messageId string) {
	storedMessage := models.Message{
		JID:         jid,
		Destination: destination,
		MessageId:   messageId,
		Body:        body,
		CreatedAt:   time.Time{},
		UpdatedAt:   time.Time{},
	}

	app.Models.Create(&storedMessage)
}
