/*

TODO:
- configure http verbs
- consider switching to cbor
- robust error handling
- images

*/

package pifi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/andrew-r-thomas/pifi/meta"

	_ "github.com/mattn/go-sqlite3"
)

type Server struct {
	ms  meta.MetaStore
	ctx context.Context
}

func NewServer(ctx context.Context) (s Server, err error) {
	// prep tracks file, just for dev for now
	err = os.Mkdir("tracks", 0755)
	if err != nil {
		if !errors.Is(err, os.ErrExist) {
			return
		}
	}

	// context
	s.ctx = ctx

	// setup meta store
	s.ms, err = meta.NewMetaStore(s.ctx)
	if err != nil {
		return
	}

	// setup http handlers
	http.HandleFunc("/upload-track", s.uploadTrack)
	http.HandleFunc("/album-meta", s.albumMeta)
	http.HandleFunc("/get-track", s.getTrack)
	http.HandleFunc("/get-library", s.getLibrary)
	http.HandleFunc("/get-image", s.getImage)

	return
}

func (s *Server) Start(addr string) error {
	return http.ListenAndServe(addr, nil)
}

func (s *Server) uploadTrack(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	id := params.Get("id")
	file, err := os.Create(fmt.Sprintf("tracks/%s.flac", id))
	if err != nil {
		log.Fatalf("error creating track file: %v\n", err)
	}
	io.Copy(file, r.Body)
}

func (s *Server) albumMeta(w http.ResponseWriter, r *http.Request) {
	var albumMeta meta.AlbumMeta
	err := json.NewDecoder(r.Body).Decode(&albumMeta)
	if err != nil {
		log.Fatalf("error decoding json: %v\n", err)
	}

	addAlbumResp, err := s.ms.AddAlbum(&albumMeta)
	if err != nil {
		log.Fatalf("error adding album: %v\n", err)
	}

	json.NewEncoder(w).Encode(&addAlbumResp)
}

// first we'll just get the whole file downloaded and playing
func (s *Server) getTrack(w http.ResponseWriter, r *http.Request) {
	vals := r.URL.Query()
	id := vals.Get("id")
	log.Printf("id: %s\n", id)
	f, err := os.Open(fmt.Sprintf("tracks/%s.flac", id))
	if err != nil {
		log.Fatalf("error opening file: %v\n", err)
	}

	_, err = io.Copy(w, f)
	if err != nil {
		log.Fatalf("error copying: %v\n", err)
	}
}

func (s *Server) getLibrary(w http.ResponseWriter, r *http.Request) {
	log.Printf("got get library request\n")
	resp, err := s.ms.GetLibrary()
	if err != nil {
		log.Fatalf("error getting library: %v\n", err)
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) getImage(w http.ResponseWriter, r *http.Request) {
	vals := r.URL.Query()
	id := vals.Get("id")
	http.ServeFile(w, r, fmt.Sprintf("images/%s.jpg", id))
}
