package localhttpcall

import (
	"io"
	"net/http"
)

// LocalHTTPCall todo
type LocalHTTPCall struct {
	Port       string
	HTTPClient *http.Client
}

// New todo
func New(port string, httpClient *http.Client) *LocalHTTPCall {
	return &LocalHTTPCall{
		Port:       port,
		HTTPClient: httpClient,
	}
}

// Call todo
func (s *LocalHTTPCall) Call(
	method, urlPath string,
	reqHeader http.Header,
	reqBody io.Reader,
) (int, string, http.Header, io.ReadCloser, error) {

	// d, _ := ioutil.ReadAll(reqBody)
	// log.Printf("RequestEventBodyReader ReadAll: %s", string(d))

	req, _ := http.NewRequest(method, "http://127.0.0.1:"+s.Port+urlPath, reqBody)
	req.Header = reqHeader
	res, err := s.HTTPClient.Do(req)
	if err != nil {
		return -1, "", nil, nil, err
	}
	return res.StatusCode, res.Status, res.Header, res.Body, nil
}
