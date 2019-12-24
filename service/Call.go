package service

import (
	"io"
	"net/http"
)

// Call todo
type Call interface {
	Call(string, string, http.Header, io.Reader) (int, string, http.Header, io.ReadCloser, error)
}
