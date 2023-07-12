package user

type CreateValidator struct {
	Name  string `json:"name" validate:"required,alpha,lowercase,min=5,max=20,unique=users/name"`
	Token string `json:"token" validate:"required,min=15,max=20"`
}

type UpdateValidator struct {
	Token string `json:"token" validate:"required,min=15,max=20"`
}

// Custom validation functions
