package file

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
	apiformattertrait "gomeow/cmd/api/controllers/traits"
	"gomeow/cmd/dto"
	"gomeow/cmd/models"
	"gomeow/pkg/application"
	"gomeow/pkg/wmeow"
	"net/http"
)

func SendFile(app *application.Application) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		device := r.Context().Value("device").(models.Device)

		// check if device is connected
		if wmeow.ClientPointer[device.ID] == nil {
			apiformattertrait.WriteErrorResponse(w, http.StatusBadRequest, "Device not connected")
			return
		}

		// check if device is logged in
		client := wmeow.ClientPointer[device.ID].WAClient
		if !client.IsConnected() || !client.IsLoggedIn() {
			apiformattertrait.WriteErrorResponse(w, http.StatusBadRequest, "Device is not logged in")
			return
		}

		var fileDto dto.FileDTO
		err := json.NewDecoder(r.Body).Decode(&fileDto)
		if err != nil {
			apiformattertrait.WriteErrorResponse(w, http.StatusUnprocessableEntity, err.Error())
			return
		}

		// convert phone number to acceptable format
		fileDto.Destination = stripPhoneNumber(fileDto.Destination)
		//recipient, err := validateMessageFields(request.Destination, &request.ContextInfo)
		if err != nil {
			zap.S().Debugf("Invalid destination number: %s", fileDto.Destination)
			apiformattertrait.WriteMultipleErrorResponse(w, http.StatusUnprocessableEntity, "Invalid destination number")
			return
		}

		// validating request
		validationError := app.Validator.Validate(fileDto)
		if validationError != nil {
			zap.S().Debugf("Error validating request: %+v", validationError)
			apiformattertrait.WriteMultipleErrorResponse(w, http.StatusUnprocessableEntity, validationError)
			return
		}

		// Do everything else here

		// generate new messageID
		newMessageId := client.GenerateMessageID()

		// create message object
		message := models.Message{
			JID:         device.JID.String,
			DeviceId:    device.ID,
			MessageId:   newMessageId,
			Destination: fileDto.Destination,
			MessageType: "file",
		}

		message.File = storeFile(&fileDto)

		// store
		result := app.Models.Create(&message)
		if result.Error != nil {
			zap.S().Debugf("Error creating message: %+v", result)
			apiformattertrait.WriteErrorResponse(w, http.StatusInternalServerError, result.Error.Error())
			return
		}

		// send message
		err = wmeow.ClientPointer[device.ID].SendFileMessage(newMessageId, &fileDto)

		apiformattertrait.WriteResponse(w, &ReturnMessageDTO{
			MessageId:   newMessageId,
			Destination: fileDto.Destination,
		})
	}
}

// storeFile handle store file here.
// Returns path to file.
func storeFile(imageMessage *dto.FileDTO) string {
	return ""
}
