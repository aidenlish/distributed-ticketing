package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/aidenlish/distributed-ticketing/internal/db"
	"github.com/aidenlish/distributed-ticketing/internal/ticketserver"
	"github.com/aidenlish/distributed-ticketing/internal/types"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// TODO
	sqlDB, err := sql.Open("mysql", "TODO")
	if err != nil {
		log.Fatal(err)
	}

	reserver := ticketserver.NewRangeReserver(db.New(sqlDB))

	http.HandleFunc("/reserve", func(w http.ResponseWriter, r *http.Request) {
		start, end, err := reserver.Reserve()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(types.RangeResponse{Start: start, End: end})
	})

	log.Println("ticket server listening on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
