package server

type ErrorResponse struct {
	Message string
}

type GetResponse struct {
	Key   string
	Value string
}
