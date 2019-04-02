package main

import (
	"bytes"
	"net/http"

	"github.com/masudur-rahman/auditsink-prototype/event"

	"gopkg.in/macaron.v1"
)

func Welcome(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Welcome.."))
}

func ReceiveEvents(ctx *macaron.Context, w http.ResponseWriter, r *http.Request) {
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(r.Body)

	if err := event.ProcessEvents(buf.Bytes()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusCreated)
	_, _ = ctx.Resp.Write([]byte("Data write successful.."))

}

func main() {
	m := macaron.Classic()

	m.Get("/", Welcome)
	m.Post("/", ReceiveEvents)

	m.Run()
}
