package web

import (
	"embed"
	"io/fs"
	"log"
)

//go:embed static/*
var embeddedStatic embed.FS

func defaultStaticFS() fs.FS {
	sub, err := fs.Sub(embeddedStatic, "static")
	if err != nil {
		log.Fatalf("embed subfs: %v", err)
	}
	return sub
}
