package poll

type CreateValidator struct {
	Question string `json:"question" validate:"required,min=5,max=255"`
}

type UpdateValidator struct {
	Question string `json:"question" validate:"required,min=5,max=255"`
}

// Custom validation functions
