package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"snowflake"
	"syscall"
	"time"
)

func main() {
	w_id, err := snowflake.GenerateWorkerId()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("snowflake worker id %d\n", w_id)
	sf, err := snowflake.NewSnowflake(w_id)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/id", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		sid := snowflake.ParseId(sf.NextId())
		json.NewEncoder(w).Encode(sid)
	})

	PORT := 8000
	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", PORT),
		Handler:        nil,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	go func() {
		log.Fatal(s.ListenAndServe())
	}()

	log.Printf("Server is listning on http://:%d\n", PORT)

	sig := <-sigs
	log.Println("Signal: ", sig)
}
