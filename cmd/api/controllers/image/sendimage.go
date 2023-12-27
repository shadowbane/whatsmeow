package image

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

func SendImage(app *application.Application) httprouter.Handle {
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

		var imageDTO dto.ImageDTO
		err := json.NewDecoder(r.Body).Decode(&imageDTO)
		if err != nil {
			apiformattertrait.WriteErrorResponse(w, http.StatusUnprocessableEntity, err.Error())
			return
		}

		// convert phone number to acceptable format
		imageDTO.Destination = stripPhoneNumber(imageDTO.Destination)
		//recipient, err := validateMessageFields(request.Destination, &request.ContextInfo)
		if err != nil {
			zap.S().Debugf("Invalid destination number: %s", imageDTO.Destination)
			apiformattertrait.WriteMultipleErrorResponse(w, http.StatusUnprocessableEntity, "Invalid destination number")
			return
		}

		// validating request
		validationError := app.Validator.Validate(imageDTO)
		if validationError != nil {
			zap.S().Debugf("Error validating request: %+v", validationError)
			apiformattertrait.WriteMultipleErrorResponse(w, http.StatusUnprocessableEntity, validationError)
			return
		}

		if imageDTO.Base64Image[0:10] != "data:image" {
			zap.S().Debugf("Image data should start with \"data:image\"")
			apiformattertrait.WriteErrorResponse(w, http.StatusUnprocessableEntity, "Image data should start with \"data:image\"")
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
			Destination: imageDTO.Destination,
			Body:        imageDTO.Message,
			MessageType: "image",
		}

		message.File = storeFile(&imageDTO)

		// store
		result := app.Models.Create(&message)
		if result.Error != nil {
			zap.S().Debugf("Error creating message: %+v", result)
			apiformattertrait.WriteErrorResponse(w, http.StatusInternalServerError, result.Error.Error())
			return
		}

		// send message
		err = wmeow.ClientPointer[device.ID].SendImageMessage(newMessageId, &imageDTO)

		apiformattertrait.WriteResponse(w, &ReturnMessageDTO{
			MessageId:   newMessageId,
			Destination: imageDTO.Destination,
			Message:     imageDTO.Message,
		})
	}
}

// storeFile handle store file here.
// Returns path to file.
func storeFile(imageMessage *dto.ImageDTO) string {
	return ""
}
