package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/joho/godotenv"
)
type DocTypeEnum string

const (
	CPF DocTypeEnum = "CPF"
	CNPJ DocTypeEnum = "CNPJ"
)
type Client struct {
	Name string `json:"name"`
	DocType DocTypeEnum `json:"doc_type"`
	DocValue string `json:"doc_value"`
	Password string `json:"password"`
	Wallet int32 `json:"wallet"`
}
func clientRoute(dbConn *pgx.Conn, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
		case "POST": 
			body := json.NewDecoder(r.Body)
			client := &Client{}
			err := body.Decode(client)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				log.Fatal("Error when decode:", err.Error())
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

func withDB (dbConn *pgx.Conn, handler func (*pgx.Conn, http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		handler(dbConn, w, r)
	}
}
func main() {
	err := godotenv.Load()
	if err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao carregar o arquivo .env")
	}
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to dabase: %v\n", err)
	}
	defer conn.Close(context.Background())

	http.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello")
	})
	http.HandleFunc("/client", withDB(conn, clientRoute))

	http.ListenAndServe(":8080", nil)
}
