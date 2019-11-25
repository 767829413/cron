package master

import (
	"net/http"
)

func SetRoute() (mux *http.ServeMux) {
	mux = http.NewServeMux()
	mux.HandleFunc("/job/save", GetHandleJobSaveFunc())
	mux.HandleFunc("/job/delete", GetHandleJobDeleteFunc())
	mux.HandleFunc("/job/list", GetHandleJobListFunc())
	mux.HandleFunc("/job/kill", GetHandleJobKillFunc())
	return
}
