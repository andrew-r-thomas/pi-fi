package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/andrew-r-thomas/pifi"
)

const port = 8080

func main() {
	// register serving of static files
	http.Handle(
		"/static/",
		http.StripPrefix(
			"/static/",
			http.FileServer(
				http.Dir(
					"./static",
				),
			),
		),
	)

	log.Printf("about to create server\n")
	// create and start main server
	s, err := pifi.NewServer(context.Background())
	if err != nil {
		log.Fatalf("error creating server: %v\n", err)
	}

	log.Printf("starting server...\n")
	err = s.Start(fmt.Sprintf(":%d", port))
	log.Printf("server stopped because: %v\n", err)
}
