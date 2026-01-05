package models

import "time"

type Seating struct {
	Version int    `json:"version"`
	Seats   []Seat `json:"seats"`
}

func (s Seating) Capacity() int {
	return len(s.Seats)
}

type Seat struct {
	ID         int       `json:"id"`
	Row        string    `json:"row"`
	SeatNumber int       `json:"seat_number"`
	HallID     int       `json:"hall_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
