package pollmessage

type CreateValidator struct {
	Destination string `json:"destination" validate:"required,numeric,min=5,max=20"`
	PollId      string `json:"poll_id" validate:"required,min=5,max=255"`
}

type UpdateValidator struct {
	Destination string `json:"destination" validate:"required,numeric,min=5,max=20"`
	PollId      string `json:"poll_id" validate:"required,min=5,max=255"`
}

// Custom validation functions
