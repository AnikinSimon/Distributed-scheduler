package main

import (
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/config"
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/app"
)

func main() {
	// TODO: config

	if err := app.Start(config.Config{}); err != nil {
		panic(err)
	}
}
