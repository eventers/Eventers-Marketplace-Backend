package response

import (
	"encoding/json"
	"eventers-marketplace-backend/model"
	"net/http"
)

type SuccessResponse struct {
	Data       *Data `json:"data"`
	StatusCode int `json:"-"`
}

type Data struct {
	User *model.User `json:"user,omitempty"`
	Auth *model.Auth `json:"auth,omitempty"`
}

func (r SuccessResponse) Send(w http.ResponseWriter) {
	w.WriteHeader(r.StatusCode)
	json.NewEncoder(w).Encode(r)
}
