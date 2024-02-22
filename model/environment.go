package model

import "time"

type Environment struct {
	Name   string    `json:"name,omitempty" bson:"name"`
	Status string    `json:"status,omitempty" bson:"status"`
	Owner  string    `json:"owner,omitempty" bson:"owner"`
	Date   time.Time `json:"date" bson:"date"`
}
