package pifi

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

const GetArtistsQuery = "select * from artists"
const DBSetup = `
	create table if not exists artists(
		id text primary key,
		name text not null
	);
`

var homeTmpl = template.Must(template.ParseFiles("./templates/index.html"))

type Server struct {
	db *sql.DB
}

func NewServer() (s Server, err error) {
	// setup db
	db, err := sql.Open("sqlite3", "./.db")
	if err != nil {
		err = fmt.Errorf("error openning db: %v\n", err)
		return
	}
	_, err = db.Exec(DBSetup)
	if err != nil {
		err = fmt.Errorf("error setting up db: %v\n", err)
		return
	}
	s.db = db

	// setup http handlers
	http.HandleFunc("/", s.home)

	return
}

func (s *Server) Start(addr string) error {
	return http.ListenAndServe(addr, nil)
}

func (s *Server) home(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.Query(GetArtistsQuery)
	if err != nil {
		log.Fatalf("error querying artists: %v\n", err)
	}
	defer rows.Close()

	artists := []string{}
	for rows.Next() {
		var artist string
		err = rows.Scan(&artist)
		if err != nil {
			log.Fatalf("error scanning artist: %v\n", err)
		}
		artists = append(artists, artist)
	}
	err = rows.Err()
	if err != nil {
		log.Fatalf("rows error: %v\n", err)
	}

	err = homeTmpl.Execute(w, artists)
	if err != nil {
		log.Fatalf("error executing home template: %v\n", err)
	}
}
