package main

import (
	"log"
	"os"

	"github.com/mewkiz/flac"
	"github.com/mewkiz/flac/meta"
)

func main() {
	path := os.Args[1]
	stream, err := flac.ParseFile(path)
	if err != nil {
		log.Fatalf("error inspecting file: %v\n", err)
	}
	defer stream.Close()

	log.Printf("unencoded audio md5sum: %032x\n", stream.Info.MD5sum[:])
	for i, block := range stream.Blocks {
		log.Printf("block %d: %v\n", i, block.Type)
		switch block.Type {
		case meta.TypeVorbisComment:
			vc := block.Body.(*meta.VorbisComment)
			log.Printf("vendor: %s\n", vc.Vendor)
			for _, tag := range vc.Tags {
				log.Printf("tag: {%s: %s}\n", tag[0], tag[1])
			}
		}
	}
}
