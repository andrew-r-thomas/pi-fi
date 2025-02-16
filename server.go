package pifi

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mewkiz/flac"
	"github.com/mewkiz/flac/meta"
)

type Server struct {
	ms MetaStore
}

func NewServer() (s Server, err error) {
	// setup meta store
	s.ms, err = NewMetaStore()
	if err != nil {
		return
	}

	// setup http handlers
	http.HandleFunc("/upload", s.upload)

	return
}

func (s *Server) Start(addr string) error {
	return http.ListenAndServe(addr, nil)
}

// TODO: maybe a buf pool
func (s *Server) upload(w http.ResponseWriter, r *http.Request) {
	// read body
	body := make([]byte, r.ContentLength)
	n, err := r.Body.Read(body)
	if err != nil {
		log.Fatalf("error reading body: %v\n")
	}
	if n != int(r.ContentLength) {
		log.Fatalf(
			"did not read full body, read %d/%d\n",
			n,
			r.ContentLength,
		)
	}

	// parse the body as flac, to extract metadata
	stream, err := flac.Parse(bytes.NewReader(body))
	if err != nil {
		log.Fatalf("error parsing body: %v\n", err)
	}

	var title string
	var artist string
	var album string
	var trackNumber int
	for _, block := range stream.Blocks {
		switch block.Type {
		case meta.TypeVorbisComment:
			b := block.Body.(*meta.VorbisComment)
			for _, tag := range b.Tags {
				switch tag[0] {
				case "TITLE":
					title = tag[1]
				case "ARTIST":
					artist = tag[1]
				case "ALBUM":
					album = tag[1]
				case "TRACKNUMBER":
					trackNumber, err = strconv.Atoi(tag[1])
					if err != nil {
						log.Fatalf(
							"error parsing track number: %v\n",
							err,
						)
					}
				}
			}
		default:
			log.Printf(
				"encountered non vorbis comment block: %s\n",
				block.Type.String(),
			)
		}
	}

	// file, err := os.Create(fmt.Sprintf("./tracks/%s.flac", id))
	// if err != nil {
	// 	log.Fatalf("error creating file: %v\n", err)
	// }
	// n, err := io.Copy(file, r.Body)
	// if err != nil {
	// 	log.Fatalf("error copying: %v\n", err)
	// }
	// if n != r.ContentLength {
	// 	log.Fatalf("did not copy full len!\n")
	// }
	// _, err = file.Seek(0, 0)
	// if err != nil {
	// 	log.Fatalf("error seeking file: %v\n", err)
	// }
	//
	// stream, err := flac.Parse(file)
	// if err != nil {
	// 	log.Fatalf("error parsing flac: %v\n", err)
	// }
	//
	// for _, block := range stream.Blocks {
	// 	log.Printf("block type: %s\n", block.Type.String())
	// }
}

const setupDB = `
	create table if not exists artists(
		id integer primary key,
		name text not null unique
	);

	create table if not exists albums(
		id integer primary key,
		title text not null,
	);

	create table if not exists tracks(
		id integer primary key,
		title text not null,
		album integer not null,
		track_number integer not null,
		foreign_key(album) references albums(id),
	);

	create table if not exists attributions(
		id integer primary key,
		artist_id integer not null,
		album_id integer,
		track_id integer,
		foreign_key(artist_id) references artists(id),
		foreign_key(album_id) references albums(id),
		foreign_key(track_id) references tracks(id)
	);
	`

const upsertArtist = `
	insert into artists(name) values(?)
	on conflict (name) do update set name = name
	returning id
	`

const upsertAlbum = `

	`
