package main
import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func ClientRoute(dbConn *pgx.Conn, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
		case "POST": 
			body := json.NewDecoder(r.Body)
			client := &Client{}
			err := body.Decode(client)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				log.Println("Error when decode:", err.Error())
				return
			}

			_, queryErr := dbConn.Exec(context.Background(), "INSERT INTO client (name, doc_type, doc_value, password, wallet) VALUES ($1, $2, $3, $4, $5)", client.Name, client.DocType, client.DocValue, client.Password, client.Wallet)
			if queryErr != nil {
				var pgErr *pgconn.PgError
				if errors.As(queryErr, &pgErr) {
					// TODO: Create a validation of pg erro to improve feedback message for the client
					fmt.Println(pgErr.Message)
					fmt.Println(pgErr.Code)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				w.WriteHeader(http.StatusInternalServerError)
				log.Println("\nErro on save new user \nerror:", queryErr.Error())
				return
			}

			w.Header().Set("Content-Type", "application/json")
			encondeErr := json.NewEncoder(w).Encode(client)
			if encondeErr != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "JSON encoding error"})
				return
			}
	}
}
