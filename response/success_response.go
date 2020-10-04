package response

import (
	"encoding/json"
	"eventers-marketplace-backend/model"
	"net/http"
)

type SuccessResponse struct {
	Data       interface{} `json:"data"`
	StatusCode int         `json:"-"`
}

type Data struct {
	User        *model.User        `json:"user,omitempty"`
	PublicEvent *model.PublicEvent `json:"public_event"`
	Auth        *model.Auth        `json:"auth,omitempty"`
}

func (r SuccessResponse) Send(w http.ResponseWriter) {
	w.WriteHeader(r.StatusCode)
	json.NewEncoder(w).Encode(r)
}
