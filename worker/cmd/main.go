package main

import (
	"github.com/AnikinSimon/Distributed-scheduler/worker/config"
	"github.com/AnikinSimon/Distributed-scheduler/worker/internal/app"
)

func main() {
	cfg := config.MustLoad()

	if err := app.Start(cfg); err != nil {
		panic(err)
	}
}
