package entity

type Job struct {
	CreatedAt      int64
	Id             string
	Interval       *string
	LastFinishedAt int64
	Once           *string
	Payload        map[string]interface{}
	Status         string
}
