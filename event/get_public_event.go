package event

import (
	"database/sql"
	"eventers-marketplace-backend/model"
	"fmt"
)

func (u *Event) GetPublicEvents(db *sql.DB) ([]model.PublicEvents, error) {
	q := `SELECT pe.public_event_id, date_time, event_title, event_description, event_image, ticket_price, 
		  COUNT(et.status) AS total_tickets FROM Public_Event pe
		  INNER JOIN Event_Tickets et ON pe.public_event_id = et.public_event_id
		  WHERE business_user_id=current_holder_id AND  et.status='ACTIVE' GROUP BY pe.public_event_id;`

	st, rows, err := query(db, q, nil)
	if err != nil {
		return nil, fmt.Errorf("getPublicEvents: error querying public events: %w", err)
	}
	pes, err := publicEventsScanner(st, rows)
	if err != nil {
		return nil, fmt.Errorf("getPublicEvents: error getting public events: %w", err)
	}

	var publicEvents []model.PublicEvents

	for i := range pes {
		publicEventID := pes[i].PublicEventID

		q = `SELECT event_ticket_id, public_event_id, status, price FROM Event_Tickets WHERE status='RESELL' AND public_event_id = ?;`
		st, rows, err := query(db, q, []interface{}{publicEventID})
		if err != nil {
			return nil, fmt.Errorf("getPublicEvents: error querying event tickets: %w", err)
		}
		ets, err := eventTicketsScanner(st, rows)
		if err != nil {
			return nil, fmt.Errorf("getPublicEvents: error getting event tickets: %w", err)
		}

		publicEvents = append(publicEvents, model.PublicEvents{
			PublicEvent: pes[i],
			EventTicket: ets,
		})
	}

	return publicEvents, nil
}

func (u *Event) GetPublicEvent(db *sql.DB, userID int64) ([]model.PublicEvents, error) {
	q := `SELECT event_ticket_id, public_event_id, status, price FROM Event_Tickets WHERE current_holder_id = ?;`
	st, rows, err := query(db, q, []interface{}{userID})
	if err != nil {
		return nil, fmt.Errorf("getPublicEvent: error querying event tickets: %w", err)
	}
	ets, err := eventTicketsScanner(st, rows)
	if err != nil {
		return nil, fmt.Errorf("getPublicEvent: error getting event tickets: %w", err)
	}

	q = `SELECT public_event_id, date_time, event_title, event_description, event_image FROM Public_Event
		  WHERE public_event_id = ?;`
	var pes []model.PublicEvents
	for i := range ets {
		publicEventID := ets[i].PublicEventID
		st, rows, err := query(db, q, []interface{}{publicEventID})
		if err != nil {
			return nil, fmt.Errorf("getPublicEvent: error querying public events: %w", err)
		}
		pe, ok, err := publicEventScanner(st, rows)
		if err != nil {
			return nil, fmt.Errorf("getPublicEvent: error getting public events: %w", err)
		}

		if !ok {
			return nil, fmt.Errorf("getPublicEvent: public event not found for the ID: %d", publicEventID)
		}

		pes = append(pes, model.PublicEvents{
			PublicEvent: *pe,
			EventTicket: ets[i],
		})
	}

	return pes, nil
}

func publicEventsScanner(st *sql.Stmt, rows *sql.Rows) ([]model.PublicEvent, error) {
	defer st.Close()

	var pes []model.PublicEvent
	for rows.Next() {
		pe := model.PublicEvent{}
		err := rows.Scan(
			&pe.PublicEventID,
			&pe.DateTime,
			&pe.EventTitle,
			&pe.EventDescription,
			&pe.EventImage,
			&pe.TicketPrice,
			&pe.TotalTickets,
		)
		if err != nil {
			return nil, fmt.Errorf("publicEventsScanner: error scanning public events: %w", err)
		}

		pes = append(pes, pe)
	}
	defer rows.Close()

	return pes, nil
}

func publicEventScanner(st *sql.Stmt, rows *sql.Rows) (*model.PublicEvent, bool, error) {
	defer st.Close()
	defer rows.Close()

	pe := model.PublicEvent{}
	if rows.Next() {
		err := rows.Scan(
			&pe.PublicEventID,
			&pe.DateTime,
			&pe.EventTitle,
			&pe.EventDescription,
			&pe.EventImage,
		)
		if err != nil {
			return nil, false, fmt.Errorf("publicEventScanner: error scanning public events: %w", err)
		}

		return &pe, true, nil
	}

	return nil, false, nil
}

func eventTicketsScanner(st *sql.Stmt, rows *sql.Rows) ([]model.EventTicket, error) {
	defer st.Close()

	var ets []model.EventTicket
	for rows.Next() {
		et := model.EventTicket{}
		err := rows.Scan(
			&et.EventTicketID,
			&et.PublicEventID,
			&et.Status,
			&et.Price,
		)
		if err != nil {
			return nil, fmt.Errorf("eventTicketsScanner: error scanning event tickets: %w", err)
		}

		ets = append(ets, et)
	}
	defer rows.Close()

	return ets, nil
}


func query(db *sql.DB, query string, args []interface{}) (*sql.Stmt, *sql.Rows, error) {
	st, err := db.Prepare(query)
	if err != nil {
		return nil, nil, fmt.Errorf("query: unable to prepare query: %s", err)
	}

	rows, err := st.Query(args...)
	if err != nil {
		return nil, nil, fmt.Errorf("query: error querying db: %s", err)
	}

	return st, rows, nil
}