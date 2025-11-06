package entity

import "github.com/google/uuid"

type Job struct {
	CreatedAt      int64
	Id             uuid.UUID
	Interval       *string
	LastFinishedAt int64
	Once           *string
	Payload        map[string]interface{}
	Status         string
}
