package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
)

const port = 8080

var libTempl = template.Must(template.ParseFiles("templates/index.html"))
var artistTempl = template.Must(template.ParseFiles("templates/artist.html"))

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/artist/", artist)
	http.HandleFunc("/", home)

	log.Printf("serving on port %d...\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	log.Printf("closing with err: %v\n", err)
}

type Library struct {
	artists []string
}

func home(w http.ResponseWriter, r *http.Request) {
	dir, err := os.ReadDir("./library")
	if err != nil {
		log.Fatalf("error reading dir: %v\n", err)
	}

	artists := make([]string, len(dir))
	for i, e := range dir {
		artists[i] = e.Name()
	}

	err = libTempl.Execute(w, artists)
	if err != nil {
		log.Fatalf("error executing template: %v\n", err)
	}
}

func artist(w http.ResponseWriter, r *http.Request) {
	segments := strings.Split(r.URL.Path, "/")
	artist := segments[len(segments)-1]
	dir, err := os.ReadDir(fmt.Sprintf("./library/%s", artist))
	if err != nil {
		log.Fatalf("error reading dir: %v\n", err)
	}

	tracks := make([]string, len(dir))
	for i, e := range dir {
		tracks[i] = e.Name()
	}

	err = artistTempl.Execute(w, tracks)
}

/*

ok so i think we want a sql db for all the meta data, maybe even the photos,
and then we just have all the tracks by uuid, and the raw flac files in a dir
somewhere. most "regular" queries will just go through db, and sound will come
from the flac files directly

*/
