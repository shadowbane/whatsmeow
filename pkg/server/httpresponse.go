package server

import (
	"encoding/json"
	"net/http"
)

type SimpleHttpResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

type DataMessage struct {
	Message string `json:"message"`
}

func SendHttpResp(w http.ResponseWriter, messageString string, statusCode int) {
	var requestStatus = true
	if statusCode != 200 {
		requestStatus = false
	}

	jsonResponse := SimpleHttpResponse{
		Status:  requestStatus,
		Message: messageString,
	}

	response, _ := json.Marshal(jsonResponse)

	w.WriteHeader(statusCode)
	w.Write(response)

	return
}
