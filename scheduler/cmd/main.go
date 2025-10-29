package main

import (
	"fmt"
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/config"
	"github.com/AnikinSimon/Distributed-scheduler/scheduler/internal/app"
	"log"
	"os"
)

func main() {
	log.Printf("djbhjbahjdajhdawda")
	fmt.Println(os.Getwd())

	//for {
	//
	//}

	cfg := config.MustLoad()

	// TODO: config

	if err := app.Start(cfg); err != nil {
		panic(err)
	}

}
