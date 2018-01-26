package server

// Response represents a server request response
type Response struct {
	Status int         `json:"status"`
	Msg    string      `json:"msg"`
	Body   interface{} `json:"body"`
}
