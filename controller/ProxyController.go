package controller

import (
	"fmt"
	"io"
	"net/http"

	"github.com/SpeedVan/go-common/app/web"
	"github.com/SpeedVan/runtime-proxy/service"
)

// ProxyController 路由入口
type ProxyController struct {
	web.Controller
	ProxyCall service.Call
}

// GetRoute todo
func (s *ProxyController) GetRoute() web.RouteMap {
	items := []*web.RouteItem{
		&web.RouteItem{Path: "/{_dummy:.*}", HandleFunc: s.Call},
	}

	return web.NewRouteMap(items...)
}

// Call todo
func (s *ProxyController) Call(w http.ResponseWriter, r *http.Request) {
	resStatusCode, _, resHeader, resBody, err := s.ProxyCall.Call(r.Method, r.URL.Path, r.Header, r.Body)

	if err != nil {
		http.Error(w, fmt.Sprintf("Error : %s", err.Error()), http.StatusInternalServerError)
		return
	}
	for k, v := range resHeader {
		w.Header()[k] = v
	}
	w.WriteHeader(resStatusCode)
	_, err = io.Copy(w, resBody)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error : %s", err.Error()), http.StatusInternalServerError)
		return
	}
}
