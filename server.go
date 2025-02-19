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
	err = os.RemoveAll("tracks")
	if err != nil {
		return
	}
	err = os.Mkdir("tracks", 0755)
	if err != nil {
		return
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
