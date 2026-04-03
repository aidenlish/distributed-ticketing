package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/aidenlish/distributed-ticketing/internal/allocator"
	"github.com/aidenlish/distributed-ticketing/internal/types"
)

func refillFromTicketServer() (start, end int64, err error) {
	resp, err := http.Post("http://localhost:8081/reserve", "application/json", nil)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	var r types.RangeResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return 0, 0, err
	}
	return r.Start, r.End, nil
}

func main() {
	alloc := allocator.New(refillFromTicketServer)

	http.HandleFunc("/id", func(w http.ResponseWriter, r *http.Request) {
		id, err := alloc.Next()
		if err != nil {
			log.Printf("app server - alloc.Next: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "%d\n", id)
	})

	log.Println("app server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
