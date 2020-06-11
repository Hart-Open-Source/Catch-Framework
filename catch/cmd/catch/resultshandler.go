package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"net/http"
	"regexp"
	"strings"
	"time"
)

func resultshandler(globalconfig *Config) error {
	// Init mongo DB client
	// Manage acquisition of results for workstations and/or servers
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
	cleardatabase(mongoclient, &ctx)
	getresults(globalconfig, &ctx, mongoclient)

	return nil
}

func gethosts(globalconfig *Config) []string {
	var hostlist []string
	indexname := "osquery-result-%s"
	currenttime := time.Now()
	date := currenttime.Format("2006.01.02")
	indexname = fmt.Sprintf(indexname, date)
	rawjson := `{"size": 0,
	"aggs" : {
		"langs" : {
			"terms" : { "field" : "hostIdentifier.keyword",  "size" : 500 }
		}
	}}`

	url := fmt.Sprintf("http://%s:%s/%s/_search", globalconfig.Esserver, globalconfig.Esport, indexname)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(rawjson)))
	req.SetBasicAuth(globalconfig.Esuser, globalconfig.Espass)
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		var e map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&e); err != nil {
			panic(err)
		}
		printout("Could not decode hosts search response json from ES!")
		fmt.Println(e)
	} else {
		var r map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
			panic(err)
		}

		aggs := r["aggregations"].(map[string]interface{})
		langs := aggs["langs"].(map[string]interface{})
		buckets := langs["buckets"].([]interface{})
		for _, bucket := range buckets {
			key := bucket.(map[string]interface{})["key"]
			hostlist = append(hostlist, key.(string))
		}
		printout("Hosts found and aggregated")
	}
	return hostlist
}
func getresults(globalconfig *Config, ctx *context.Context, mongoclient *mongo.Client) error {
	// Get a list of all hosts (workstations and servers) from ES
	// foreach host
	//		get all results for host
	// 			foreach result
	// 				resulttest
	indexname := "osquery-result-%s"
	currenttime := time.Now()
	date := currenttime.Format("2006.01.02")
	indexname = fmt.Sprintf(indexname, date)
	hostlist := gethosts(globalconfig)

	for _, host := range hostlist {
		rawjson := `{"query": {
			"bool": {"filter": 
			[{
				"match": {
					"decorations.hostname": "%s"}},{
						"range": {
							"@timestamp": {
								"gte": "now-5m", "lte": "now"}
								}}]}}}`

		rawjson = fmt.Sprintf(rawjson, host)
		url := fmt.Sprintf("http://%s:%s/%s/_search", globalconfig.Esserver, globalconfig.Esport, indexname)
		req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(rawjson)))
		req.SetBasicAuth(globalconfig.Esuser, globalconfig.Espass)
		req.Header.Add("Content-Type", "application/json")
		client := &http.Client{}

		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			var e map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&e); err != nil {
				panic(err)
			}
			printout("Could not decode results search response json from ES!")
		} else {
			var r map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
				panic(err)
			}

			for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
				doc := hit.(map[string]interface{})
				source := doc["_source"]
				querypathname := source.(map[string]interface{})["name"]
				querypathslice := strings.Split(querypathname.(string), "/")
				queryname := querypathslice[len(querypathslice)-1]
				snapshot := source.(map[string]interface{})["snapshot"]
				snapshotdata := getsnapshotdata(snapshot)

				resulthandler(globalconfig, host, snapshotdata, queryname, mongoclient, ctx)
			}
		}
	}
	return nil
}

func getsnapshotdata(snapshot interface{}) string {
	if snapshot != nil {
		result := snapshot.(interface{})
		if result != nil {
			resultslice := result.([]interface{})
			if len(resultslice) == 1 {
				resultmap := resultslice[0].(map[string]interface{})
				for _, value := range resultmap {
					return value.(string)
				}
			}
		}
	}

	return ""
}

func resulthandler(globalconfig *Config, hostname string, snapshotdata string, queryname string, mongoclient *mongo.Client, ctx *context.Context) error {
	for _, querypack := range globalconfig.querypacks {
		resulttest(globalconfig, querypack, hostname, snapshotdata, queryname, mongoclient, ctx)
	}
	return nil
}

func resulttest(globalconfig *Config, querypack map[string]Query, hostname string, snapshotdata string, queryname string, mongoclient *mongo.Client, ctx *context.Context) error {
	// Check if a previous host report exists for a hostname in mongo DB and use that if it does exist
	// Otherwise create a new host report
	// For each query in globalconfig's querypack pertaining to that host type
	//		Use querypack query's regex to extract result data from sourcedata
	//		Use querypack query's test regex against result to see if it passed of failed the test
	//		Save pass/fail to mongo DB for that query's test

	var collection *mongo.Collection
	var extractmatchlist []string
	var successmatchlist []string
	var controllist []string
	var impllist []string

	extractmatchlist = querypack[queryname].Matches
	successmatchlist = querypack[queryname].Successconditions
	controllist = querypack[queryname].Hitrustcontrols
	impllist = querypack[queryname].Implementations

	db := mongoclient.Database("hitrust")
	collection = db.Collection("hosts")

	var hostreport HostReport
	filter := bson.D{{"hostname", hostname}}
	err := collection.FindOne(*ctx, filter).Decode(&hostreport)

	if err != nil { //no previous hostreport existed, so populate a new one
		hostreport.Hostname = hostname
		var controlmap map[string][]string
		controlmap = make(map[string][]string)
		hostreport.Controlmap = controlmap
	}

	for i := range extractmatchlist {
		extractpat := extractmatchlist[i]
		successpat := successmatchlist[i]
		control := controllist[i]
		implementation := impllist[i]

		re := regexp.MustCompile(extractpat)
		resultslice := re.FindStringSubmatch(snapshotdata)
		var controlresult []string

		if len(resultslice) == 2 {
			matcheddata := resultslice[1]
			re = regexp.MustCompile(successpat)
			passfailbool := re.MatchString(matcheddata)

			if passfailbool == true {
				controlresult = []string{implementation, matcheddata, "pass"}
				printout(fmt.Sprintf("[Host: %s] [Test Type: %s] [Value: %s] [Result: pass]", hostname, queryname, matcheddata))
			} else {
				controlresult = []string{implementation, matcheddata, "fail"}
				printout(fmt.Sprintf("[Host: %s] [Test Type: %s] [Value: %s] [Result: pass]", hostname, queryname, matcheddata))
			}
		} else {
			controlresult = []string{implementation, snapshotdata, "result not found"}
			printout(fmt.Sprintf("[Host: %s] [Test Type: %s] [Value: not found] [Result: N/A]", hostname, queryname))
		}
		hostreport.Controlmap[control] = controlresult
	}

	_, err = collection.UpdateOne(
		*ctx,
		bson.M{"hostname": hostname},
		bson.M{
			"$set": bson.M{
				"controlmap": hostreport.Controlmap,
			},
		},
		options.Update().SetUpsert(true),
	)

	if err != nil {
		panic(err)
	}
	return nil
}

func cleardatabase(mongoclient *mongo.Client, ctx *context.Context) error {
	// Delete everything from hitrust mongo DB to start a fresh testing session
	db := mongoclient.Database("hitrust")
	collection := db.Collection("hosts")
	err := collection.Drop(*ctx)
	if err != nil {
		panic(err)
	}
	return nil
}

func getclient() *mongo.Client {
	// Init mongo DB client
	clientOptions := options.Client().ApplyURI("mongodb://mongodb:27017")
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		panic(err)
	}
	return client
}

type HostReport struct {
	ID       primitive.ObjectID `json:"ID" bson:"_id,omitempty"`
	Hostname string             `json: "hostname" bson: "hostname"`
	// Platform   string              `json: "platform" bson: "platform"`
	Controlmap map[string][]string `json: "controlmap" bson: "controlmap"`
}

type SearchQuery struct {
	ESQuery ESQuery `json:"query"`
}
type Match struct {
	Name string `json:"name"`
}
type Timestamp struct {
	Gte string `json:"gte"`
	Lte string `json:"lte"`
}
type Range struct {
	Timestamp Timestamp `json:"@timestamp"`
}
type Filter struct {
	Match Match `json:"match,omitempty"`
	Range Range `json:"range,omitempty"`
}
type ESBool struct {
	Filters []Filter `json:"filter"`
}
type ESQuery struct {
	ESBool ESBool `json:"bool"`
}
