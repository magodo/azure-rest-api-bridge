package mockserver

import (
	"net/http"
	"time"

	"github.com/magodo/azure-rest-api-bridge/log"
)

func Serve(addr string) (*http.Server, chan struct{}, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"location": "westeurope", "tags": {"foo": "bar"}}`))
	})
	srv := &http.Server{
		Addr:         addr,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      mux,
	}
	shutdownCh := make(chan struct{})
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Error("HTTP server ListenAndServe", "error", err)
		}
		close(shutdownCh)
	}()
	return srv, shutdownCh, nil
}
