package models

import "fmt"

type RoutePoint struct {
	Name           string `json:"name"`
	StartTimestamp int64  `json:"start_timestamp"`
	EndTimestamp   int64  `json:"end_timestamp"`
	RouteDetails   string `json:"route_details"`
}

func (rp *RoutePoint) SetDescription(id string, price float64) {
	rp.RouteDetails = fmt.Sprintf("Trip %s, Price: %.2f", id, price)
}
