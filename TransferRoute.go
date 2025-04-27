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

type WalletTransfer struct {
	Receiver string `json:"payee"`
	Sender string `json:"payer"`
	Value int32 `json:"value"`
} 
type AuthorizationResponse struct {
	Status string `json:"status"`
	Data struct { 
		Authorization bool `json:"authorization"`
	} `json:"data"`
}

func isClientAuthorized() (isAuthorized bool, err error) {
	response, getErr := http.Get("https://util.devi.tools/api/v2/authorize")
	if getErr != nil {
		err = getErr
		return
	}
	body := json.NewDecoder(response.Body)
	authorizationResponse := &AuthorizationResponse{}
	decodeErr := body.Decode(authorizationResponse)
	if decodeErr != nil {
		err = decodeErr
		return
	}
	isAuthorized = authorizationResponse.Data.Authorization
	return
}

func GetClient(dbConn *pgx.Conn, docValueId string) (err error, client Client) {
	log.Println("[GetClient] - docValueId", docValueId)
	rows, execErr := dbConn.Query(context.Background(), "SELECT name, doc_type, doc_value, wallet FROM client WHERE doc_Value = $1;", docValueId)
	if execErr != nil {
		var pgErr *pgconn.PgError
		if errors.As(execErr, &pgErr) {
			log.Println("[GetClient] - pgErr: ", pgErr.Error())
			err = pgErr
			return
		}
		err =  fmt.Errorf("SQL Error", execErr.Error())
		return 
	}
	defer rows.Close()
	
	client = Client{}
	for rows.Next() {
		err = rows.Scan(&client.Name, &client.DocType, &client.DocValue, &client.Wallet)
		return
	}
	if rowsErr := rows.Err(); err != nil {
		err = fmt.Errorf("Error when iterate:", rowsErr.Error())
		return
	}

	return 
}

func TransferRoute(dbConn *pgx.Conn, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
		case "POST":
			body := json.NewDecoder(r.Body)
			walletTransferData := &WalletTransfer{}
			err := body.Decode(walletTransferData)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadGateway)
				log.Println("Error when decode:", err.Error())
				return
			}
			clientErr, sender := GetClient(dbConn, walletTransferData.Sender)
			if clientErr != nil {
				http.Error(w, clientErr.Error(), http.StatusNotFound)
				log.Println("client not found", walletTransferData.Sender, clientErr.Error())
				return
			}
			if sender.DocType == CNPJ {
				http.Error(w, "", http.StatusBadRequest)
				
				log.Println("CNPJ clients cannot send wallet points", walletTransferData.Sender)
				return
			}
			if sender.Wallet <= walletTransferData.Value {
				http.Error(w, "Insufficient credit", http.StatusBadRequest)
				log.Println("Insufficient credit", walletTransferData.Sender)
				return
			}
			isAuthorized, isAuthorizedErr := isClientAuthorized()
			if isAuthorizedErr != nil {
				http.Error(w, isAuthorizedErr.Error(), http.StatusInternalServerError)
				return
			}
			if !isAuthorized {
				http.Error(w, "client is not authorized", http.StatusUnauthorized)
				return
			}
			queryCtx, queryCtxErr := dbConn.Begin(context.Background())
			if queryCtxErr != nil {
				http.Error(w, queryCtxErr.Error(), http.StatusInternalServerError)
				log.Println("Error when decode:", queryCtxErr.Error())
				return
			}
			newSenderWalletValue := sender.Wallet - walletTransferData.Value
			queryCtx.Exec(context.Background(), "UPDATE client SET wallet = $1 WHERE doc_value = $2", newSenderWalletValue, sender.DocValue)
	
			receiverErr, receiver := GetClient(dbConn, walletTransferData.Receiver) 
			if receiverErr != nil {
				http.Error(w, receiverErr.Error(), http.StatusNotFound)
				log.Println("client not found", walletTransferData.Sender, receiverErr.Error())
				return
			}
			newReceiverWalletValue := receiver.Wallet + walletTransferData.Value
			queryCtx.Exec(context.Background(), "UPDATE client SET wallet = $1 WHERE doc_value = $2", newReceiverWalletValue, walletTransferData.Receiver)


			commitErr := queryCtx.Commit(context.Background())
			if commitErr != nil {
				http.Error(w, commitErr.Error(), http.StatusInternalServerError)
				log.Println("Error on exec transation: ", commitErr.Error())
				return
			}
			return
	}
}
