package pipelinetests

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

type RequestsSink interface {
	AllRequestsReceived() []http.Request
}

type httpSinkHandler struct {
	requests []http.Request
}

func (h *httpSinkHandler) AllRequestsReceived() []http.Request {
	return h.requests
}

func newHttpSink(ctx context.Context, port int) RequestsSink {
	sink := &httpSinkHandler{}
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: sink,
	}

	go func() {
		log.Printf("http sink listening on port %d", port)
		err := server.ListenAndServe()
		if err != nil {
			log.Printf("http sink stopped with error: %s", err)
		}
	}()

	go func() {
		<-ctx.Done()
		log.Println("shutting down http sink")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := server.Shutdown(ctx)
		if err != nil {
			log.Printf("error shutting down http sink: %s", err)
		}
	}()
	return sink
}

func (h *httpSinkHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	log.Printf("http sink got request: %v", request)
	h.requests = append(h.requests, *request)
	_, err := writer.Write([]byte("got it, thanks!"))
	if err != nil {
		log.Printf("http sink warning: error writing response: %v", err)
	}
}
