package traits

import (
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
)

type ResponseData struct {
	Success bool         `json:"success"`
	Data    *interface{} `json:"data"`
}

func (r *ResponseData) SetData(data interface{}) {
	r.Data = &data
}

func (r *ResponseData) ToJson() []byte {
	response, _ := json.Marshal(r)

	return response
}

func WriteResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")

	responseData := ResponseData{
		Success: true,
		Data:    nil,
	}

	responseData.SetData(data)

	_, err := w.Write(responseData.ToJson())
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
	}
}

// WriteErrorResponse writes an error response to the client in the following format:
//
//	 {
//	   "success": false,
//	   "data": {
//		     "message": "ERRINFO"
//		   }
//		}
//
// Basically, the error message is ResponseData with ResponseData.Success = false,
// and ResponseData.Data.Message = is the error message.
func WriteErrorResponse(w http.ResponseWriter, errorCode int, errorMsg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errorCode)

	responseData := &ResponseData{
		Success: false,
	}

	responseData.SetData(map[string]string{
		"message": errorMsg,
	})

	err := json.
		NewEncoder(w).Encode(responseData)

	if err != nil {
		zap.S().Fatalf(err.Error())
	}
}
