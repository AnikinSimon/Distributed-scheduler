package main

import (
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/config"
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/app"
)

func main() {
	cfg := config.MustLoad()

	if err := app.Start(cfg); err != nil {
		panic(err)
	}
}
