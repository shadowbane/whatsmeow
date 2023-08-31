package poll_detail

type CreateValidator struct {
	Option string `json:"option" validate:"required,min=1,max=255"`
}

type UpdateValidator struct {
	Option string `json:"option" validate:"required,min=1,max=255"`
}

// Custom validation functions
