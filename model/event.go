package model

import (
	"time"
)

type PublicEvents struct {
	PublicEvent PublicEvent `json:"public_event"`
	EventTicket interface{} `json:"event_ticket"`
}

type PublicEvent struct {
	PublicEventID    int64      `json:"public_event_id,omitempty"`
	DateTime         *time.Time `json:"date_time,omitempty"`
	EventTitle       *string    `json:"event_title,omitempty"`
	EventDescription *string    `json:"event_description,omitempty"`
	EventImage       *string    `json:"event_image,omitempty"`
	TotalTickets     uint64     `json:"total_tickets,omitempty"`
	TicketPrice      uint64     `json:"ticket_price,omitempty"`
}

type EventTicket struct {
	EventTicketID     int64   `json:"event_ticket_id,omitempty"`
	BusinessUserID    int64   `json:"business_user_id,omitempty"`
	PublicEventID     int64   `json:"public_event_id,omitempty"`
	AssetID           uint64  `json:"asset_id,omitempty"`
	CurrentHolderID   int64   `json:"current_holder_id,omitempty"`
	Status            *string `json:"status,omitempty"`
	AvailableToResell *bool   `json:"available_to_resell,omitempty"`
	Price             uint64  `json:"price,omitempty"`
}

type Ticket struct {
	EventTicketID int64   `json:"event_ticket_id,omitempty"`
	PublicEventID int64   `json:"public_event_id,omitempty"`
	FromUserID    int64   `json:"from_user_id,omitempty"`
	ToUserID      int64   `json:"to_user_id,omitempty"`
	Status        *string `json:"status,omitempty"`
	PriceToResell int64   `json:"price_to_resell,omitempty"`
}
