package controllers

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"go.uber.org/zap"
	"gomeow/pkg/application"
	"google.golang.org/protobuf/proto"
	"io"
	"log"
	"net/http"
)

type JsonErrorResponse struct {
	Error *ApiError `json:"error"`
}

type ApiError struct {
	Status int    `json:"status"`
	Title  string `json:"title"`
}

type returnData struct {
	Success bool            `json:"success"`
	Data    detailedMessage `json:"data"`
}

type detailedMessage struct {
	Message   string `json:"message"`
	To        string `json:"to"`
	MessageId string `json:"messageId"`
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

		newJid := types.NewJID(to, "s.whatsapp.net")

		newMessage := &waProto.Message{
			ExtendedTextMessage: &waProto.ExtendedTextMessage{
				Text: proto.String(message),
			},
		}

		zap.S().Debugf("Sending a message with ID: %s and content: %s to %s", newMessageId, newMessage.String(), to)
		_, err := app.Meow.Client.SendMessage(newJid, newMessageId, newMessage)
		if err != nil {
			zap.S().Errorf(err.Error())
			writeErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		formattedValues := returnData{
			Success: true,
			Data: detailedMessage{
				Message:   message,
				To:        to,
				MessageId: newMessageId,
			},
		}

		response, _ := json.Marshal(formattedValues)
		_, err = w.Write(response)
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
