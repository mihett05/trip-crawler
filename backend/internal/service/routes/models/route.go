package models

import "fmt"

type RoutePoint struct {
	Name           string   `json:"name"`
	StartTimestamp int64    `json:"start_timestamp"`
	EndTimestamp   int64    `json:"end_timestamp"`
	Details        string   `json:"details"`
	Price          float64  `json:"price,omitempty"` // ← добавить
	Latitude       *float64 `json:"latitude,omitempty"`
	Longitude      *float64 `json:"longitude,omitempty"`
}

func (rp *RoutePoint) SetDescription(id string, price float64) {
	rp.Details = fmt.Sprintf("Trip %s, Price: %.2f", id, price)
}
