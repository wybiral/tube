package main

import (
	"fmt"
	"log"
	"os"

	"github.com/wybiral/tube/pkg/app"
)

func main() {
	cfg := app.DefaultConfig()
	err := cfg.ReadFile("config.json")
	if err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}
	a, err := app.NewApp(cfg)
	if err != nil {
		log.Fatal(err)
	}
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Local server: http://%s", addr)
	err = a.Run()
	if err != nil {
		log.Fatal(err)
	}
}
