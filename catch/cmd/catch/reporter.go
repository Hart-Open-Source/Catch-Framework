package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"net/http"
	"strings"
	"time"
)

func printresults(globalconfig *Config) []string {
	//Print results for workstations or servers from mongo DB
	mongoclient := getclient()
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err := mongoclient.Connect(ctx)
	if err != nil {
		panic(err)
	}

	err = mongoclient.Ping(ctx, readpref.Primary())
	if err != nil {
		panic(err)
	}
	defer mongoclient.Disconnect(ctx)

	db := mongoclient.Database("hitrust")
	var collection *mongo.Collection
	var tablelist []string

	collection = db.Collection("hosts")
	tablelist = buildtable(ctx, collection)

	return tablelist
}

func createjiratickets(globalconfig *Config) {
	mongoclient := getclient()
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err := mongoclient.Connect(ctx)
	if err != nil {
		panic(err)
	}

	err = mongoclient.Ping(ctx, readpref.Primary())
	if err != nil {
		panic(err)
	}
	defer mongoclient.Disconnect(ctx)

	db := mongoclient.Database("hitrust")
	var collection *mongo.Collection
	collection = db.Collection("hosts")

	cursor, err := collection.Find(ctx, bson.M{"controlmap": bson.M{"$ne": bson.M{}}})
	if err != nil {
		panic(err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		payload := `{
			"fields": {
				"project": {
					"key": "CAT"
				},
				"issuetype": {
					"name": "Bug"
				},
				"summary": "HITRUST Results: %s",
				"description": "%s"
				}
			}`

		var hostreport HostReport
		err = cursor.Decode(&hostreport)
		if err != nil {
			panic(err)
		}

		issues := ""

		for control, test := range hostreport.Controlmap {
			if test[1] == "fail" {
				issues = issues + fmt.Sprintf("*HITRUST Control Failed:* %s\\n*Description:* %s\\n---", control, test[0])
			}
		}

		payload = fmt.Sprintf(payload, hostreport.Hostname, issues)
		req, err := http.NewRequest("POST", globalconfig.Jiraurl, bytes.NewBuffer([]byte(payload)))
		req.SetBasicAuth(globalconfig.Jirauser, globalconfig.Jirapass)
		req.Header.Add("Content-Type", "application/json")
		client := &http.Client{}

		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
	}
}

func buildtable(ctx context.Context, collection *mongo.Collection) []string {
	// Construct a table using mongo collection of documents
	cursor, err := collection.Find(ctx, bson.M{"controlmap": bson.M{"$ne": bson.M{}}})

	if err != nil {
		panic(err)
	}
	defer cursor.Close(ctx)
	var tablelist []string

	for cursor.Next(ctx) {

		tablestring := &strings.Builder{}
		table := tablewriter.NewWriter(tablestring)
		table.SetHeader([]string{"Hostname", "HITRUST Control", "Description", "Pass/Fail", "Value"})
		var hostreport HostReport
		err = cursor.Decode(&hostreport)
		if err != nil {
			panic(err)
		}

		for control, test := range hostreport.Controlmap {
			value := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(test[1], " ", ""), "\t", ""), "\n", "")
			table.Append([]string{hostreport.Hostname, control, test[0], test[2], value})
		}

		table.Render()
		tablelist = append(tablelist, tablestring.String())
	}
	return tablelist
}
