package main

import (
	"log"

	"github.com/wybiral/tube/pkg/app"
)

const addr = "127.0.0.1:40404"

func main() {
	a, err := app.NewApp()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Serving at http://%s", addr)
	err = a.Run(addr)
	if err != nil {
		log.Fatal(err)
	}
}
