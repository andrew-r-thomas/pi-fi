package pifi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/andrew-r-thomas/pifi/meta"

	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mewkiz/flac"
	flacMeta "github.com/mewkiz/flac/meta"
)

type Server struct {
	ms  meta.MetaStore
	ctx context.Context
}

func NewServer(ctx context.Context) (s Server, err error) {
	s.ctx = ctx

	// setup meta store
	s.ms, err = meta.NewMetaStore(s.ctx)
	if err != nil {
		return
	}

	// setup http handlers
	http.HandleFunc("/upload-track", s.uploadTrack)
	http.HandleFunc("/album", s.getAlbum)

	return
}

func (s *Server) Start(addr string) error {
	return http.ListenAndServe(addr, nil)
}

func (s *Server) uploadTrack(w http.ResponseWriter, r *http.Request) {
	log.Printf("upload track hit\n")

	// read body
	// TODO: maybe a buf pool
	body := make([]byte, r.ContentLength)
	n := 0
	for {
		nn, err := r.Body.Read(body[n:])
		n += nn
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else {
				log.Fatalf("error reading body: %v\n", err)
			}
		}
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
		case flacMeta.TypeVorbisComment:
			b := block.Body.(*flacMeta.VorbisComment)
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

	albumId, err := s.ms.AddTrack(title, artist, album, trackNumber)
	if err != nil {
		if errors.Is(err, sqlite3.ErrConstraintUnique) {
			// TODO: handle track dup situation
		} else {
			log.Fatalf("error adding track to meta db: %v\n", err)
		}
	}

	fmt.Fprintf(w, "%d", albumId)
}

func (s *Server) getAlbum(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	id, err := strconv.Atoi(params.Get("id"))
	if err != nil {
		log.Fatalf("error converting id to string: %v\n", err)
	}

	album, err := s.ms.GetAlbum(id)
	if err != nil {
		log.Fatalf("error getting album: %v\n", err)
	}

	err = json.NewEncoder(w).Encode(&album)
	if err != nil {
		log.Fatalf("error encoding album json: %v\n", err)
	}
}
