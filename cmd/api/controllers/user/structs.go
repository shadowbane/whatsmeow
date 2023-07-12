package user

type CreateRequest struct {
	Name  string `json:"name"`
	Token string `json:"token"`
}
