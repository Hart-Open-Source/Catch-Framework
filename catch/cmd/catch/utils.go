package main

import (
	"fmt"
	"os"
	"time"
)

func getenvironmentalvars(globalconfig *Config) error {
	esserver := os.Getenv("ELASTIC_SERVER")
	esport := os.Getenv("ELASTIC_PORT")
	esuser := os.Getenv("ELASTIC_USER")
	espass := os.Getenv("ELASTIC_PASSWORD")
	jiraurl := os.Getenv("JIRA_URL")
	jirauser := os.Getenv("JIRA_USER")
	jirapass := os.Getenv("JIRA_PASSWORD")

	if esserver == "" || esport == "" || esuser == "" || espass == "" {
		printout("Environmental variables for Elasticsearch server missing. Configure ES_SERVER, ES_PORT, ES_USER and ES_PASS")
		os.Exit(1)
	} else {
		globalconfig.Esserver = esserver
		globalconfig.Esport = esport
		globalconfig.Esuser = esuser
		globalconfig.Espass = espass
		globalconfig.Jiraurl = jiraurl
		globalconfig.Jirauser = jirauser
		globalconfig.Jirapass = jirapass
	}
	return nil
}

func printout(message string) {
	currenttime := time.Now()
	fmt.Println(fmt.Sprintf("[%s] %s", currenttime.Format("2006.01.02 15:04:05"), message))
}
