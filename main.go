package main

import (
	"github.com/k4ties/dystopia/dystopia"
	"log"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
)

func main() {
	l := slog.Default()
	conf := dystopia.MustReadConfig("config.yaml")

	startDebugServer()

	d := dystopia.New(l, conf)
	d.Start()
}

func startDebugServer() {
	go func() {
		log.Fatalf(http.ListenAndServe("localhost:1337", nil).Error())
	}()
}
