package main

type Config struct {
	Esserver     string
	Esport       string
	Esuser       string
	Espass       string
	Wsquerypack  map[string]Query
	Srvquerypack map[string]Query
	querypacks   []map[string]Query
	Jiraurl      string
	Jirauser     string
	Jirapass     string
	Logpath      string
	Maxlogsize   string
}
