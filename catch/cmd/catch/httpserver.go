package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"strings"
	"time"
)

var globalconfig *Config

func audit(w http.ResponseWriter, req *http.Request) {

	filter := req.URL.Query().Get("filter")
	jira := req.URL.Query().Get("jira")

	switch filter {
	case "workstations":
		printout("Gathering audit results for workstations")
		loadquerypacks(globalconfig, "workstations")
	case "servers":
		printout("Gathering audit results for servers")
		loadquerypacks(globalconfig, "servers")
	}
	resultshandler(globalconfig)
	tablelist := printresults(globalconfig)
	joinedtable := strings.Join(tablelist, "\n")

	if jira == "1" {
		createjiratickets(globalconfig)
	}

	fmt.Fprint(w, joinedtable)
}

func runhttp(gconfig *Config) {
	globalconfig = gconfig
	router := mux.NewRouter()
	router.Handle("/metrics", promhttp.Handler())
	router.HandleFunc("/audit", audit)

	srv := &http.Server{
		Handler:      router,
		Addr:         "0.0.0.0:9090",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	srv.ListenAndServe()
}
