package model

type PublicEventRequest struct {
	Data struct {
		PublicEvent *PublicEvent `json:"public_event,omitempty" validate:"required"`
		Auth        *Auth        `json:"auth,omitempty" validate:"required"`
	} `json:"data"`
}

type PublicEventUpdateReq struct {
	Data struct {
		Ticket *Ticket `json:"ticket,omitempty" validate:"required"`
		Auth   *Auth   `json:"auth,omitempty" validate:"required"`
	} `json:"data"`
}

type PublicEventResponse struct {
	Data interface{} `json:"data"`
}
