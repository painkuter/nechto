package main

import (
	"net/http"

	"nechto/internal/app"
	"nechto/internal/pkg/apperr"
	"nechto/internal/pkg/log"
)

func LiveHandler(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("response"))
	apperr.Check(err)
	w.Header().Set("Content-Type", "application/json")
}

func main() {
	l := log.InitLogging()
	defer l.Close()

	server := app.NewServer()
	err := server.Start()
	if err != nil {
		panic(err)
	}
}
