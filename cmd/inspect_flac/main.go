package main

import (
	"log"

	"github.com/mewkiz/flac"
	"github.com/mewkiz/flac/meta"
)

func main() {
	stream, err := flac.ParseFile("../../tracks/peggy/JPEGMAFIA - I LAY DOWN MY LIFE FOR YOU (DIRECTOR'S C - 04 SIN MIEDO.flac")
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
