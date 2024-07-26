package dto

type CEPInput struct {
	CEP string
}

type CEPOutput struct {
	Logradouro string
	Bairro     string
	UF         string
	CIDADE     string
	Latitude   string
	Longitude  string
}
