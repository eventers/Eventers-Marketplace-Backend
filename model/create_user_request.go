package model

type CreateUser struct {
	Data Data `json:"data"`
}

type Data struct {
	User *User `json:"user,omitempty" validate:"required"`
	Auth *Auth `json:"auth,omitempty" validate:"required"`
}
