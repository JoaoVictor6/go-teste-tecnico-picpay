package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

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
	http.HandleFunc("/client", withDB(conn, ClientRoute))

	http.ListenAndServe(":8080", nil)
}
