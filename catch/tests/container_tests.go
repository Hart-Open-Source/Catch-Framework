package main

import (
	"fmt"
	"io/ioutil"
	"os"
	_ "os/exec"
	"strconv"
	"strings"
)

// cp results file content to /osquerylogs/osquery_results on kolide container

func main() {
	hostcount := 500
	resultpath := "container_tests_results.txt"
	templatepath := "container_tests_template.txt"
	hostnames := generatehosts(hostcount)
	generateresults(templatepath, resultpath, hostnames)
}

func generatehosts(hostcount int) []string {
	var hostnames []string
	fmt.Println("[+] Generating hostnames")
	for i := 1; i < hostcount; i++ {
		hostnames = append(hostnames, fmt.Sprintf("host-%s", strconv.Itoa(i)))
	}

	return hostnames
}

func generateresults(templatepath string, resultpath string, hostnames []string) {
	fmt.Println("[+] Generating results file")
	if _, err := os.Stat(resultpath); os.IsNotExist(err) {
		newres, err := os.Create(resultpath)
		if err != nil {
			panic(err)
		}
		defer newres.Close()
	} else {
		newres, err := os.OpenFile(resultpath, os.O_TRUNC, 0600)
		if err != nil {
			panic(err)
		}
		defer newres.Close()
	}

	read, err := ioutil.ReadFile(templatepath)
	if err != nil {
		panic(err)
	}
	f, err := os.OpenFile(resultpath, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}

	defer f.Close()
	var filecontent string
	for _, hostname := range hostnames {

		filecontent = filecontent + "\n" + strings.Replace(string(read), "***", hostname, -1)

	}
	if _, err = f.WriteString(filecontent); err != nil {
		panic(err)
	}
}
