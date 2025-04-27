package main

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
