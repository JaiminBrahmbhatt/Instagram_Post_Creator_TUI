package output

import (
	"encoding/json"
	"fmt"
	"io"
)

type Response struct {
	Status  string      `json:"status"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// OK writes a success JSON response to w.
func OK(w io.Writer, data interface{}, pretty bool) {
	write(w, Response{Status: "ok", Data: data}, pretty)
}

// Err writes an error JSON response to w.
func Err(w io.Writer, msg string, pretty bool) {
	write(w, Response{Status: "error", Message: msg}, pretty)
}

func write(w io.Writer, r Response, pretty bool) {
	var b []byte
	if pretty {
		b, _ = json.MarshalIndent(r, "", "  ")
	} else {
		b, _ = json.Marshal(r)
	}
	fmt.Fprintln(w, string(b))
}
