package main

import (
	"log"
	"os"
	"time"

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
	log.Printf("channels: %d\n", stream.Info.NChannels)
	samples := stream.Info.NSamples
	sampleRate := stream.Info.SampleRate
	dur := time.Second * time.Duration(samples/uint64(sampleRate))
	log.Printf("duration: %v\n", dur)
	for i, block := range stream.Blocks {
		log.Printf("block %d: %v\n", i, block.Type)
		switch block.Type {
		case meta.TypeVorbisComment:
			vc := block.Body.(*meta.VorbisComment)
			log.Printf("vendor: %s\n", vc.Vendor)
			for _, tag := range vc.Tags {
				log.Printf("tag: {%s: %s}\n", tag[0], tag[1])
			}
		case meta.TypeSeekTable:
			st := block.Body.(*meta.SeekTable)
			for _, point := range st.Points {
				seekTime := time.Second * time.Duration(point.SampleNum/uint64(sampleRate))
				log.Printf("seek point: %v\n", seekTime)
			}
		}
	}
}
