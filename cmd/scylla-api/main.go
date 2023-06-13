package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"time"

	"github.com/gocql/gocql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type QueryParams struct {
	Cql string
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type Column struct {
	Name string
	Type string
}

type Result struct {
	Columns []Column      `json:"columns"`
	Data    []interface{} `json:"data"`
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		panic("[os args] not setting conf file B")
	}

	// load env
	err := godotenv.Load(args[0])

	if err != nil {
		panic("[godotenv load] " + err.Error())
	}

	log.Println("[info] Loaded configuration file")

	r := mux.NewRouter()

	r.HandleFunc("/query", handler).Methods("POST")

	srv := &http.Server{
		Addr:    os.Getenv("HOST") + ":" + os.Getenv("POST"),
		Handler: r,
	}

	log.Println("[info] Listening on " + os.Getenv("HOST") + ":" + os.Getenv("POST"))

	log.Fatal(srv.ListenAndServe())
}

func handler(w http.ResponseWriter, r *http.Request) {
	// body
	body, _ := ioutil.ReadAll(r.Body)
	var params QueryParams
	json.Unmarshal(body, &params)

	// recover
	defer func() {
		r := recover()
		if r != nil {
			log.Printf("[Error] msg: %v, cql: %s\n", r, params.Cql)

			response := ErrorResponse{}
			response.Message = fmt.Sprintf("%v", r)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
		}
	}()

	// cluster config
	cluster := gocql.NewCluster("172.16.51.118", "172.16.51.120", "172.16.51.121")

	cluster.Keyspace = "oc"
	cluster.Consistency = gocql.Quorum
	cluster.ProtoVersion = 4
	cluster.Timeout = 10 * time.Second

	// session
	session, err := cluster.CreateSession()
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// context
	ctx := context.Background()

	res := query(
		ctx,
		session,
		params.Cql,
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func query(ctx context.Context, session *gocql.Session, stmt string) Result {
	iter := session.Query(stmt).WithContext(ctx).Iter()

	columns := make([]Column, 0)
	for _, columnInfo := range iter.Columns() {
		col := Column{}
		col.Name = columnInfo.Name
		col.Type = columnInfo.TypeInfo.Type().String()

		columns = append(columns, col)
	}

	data := make([]interface{}, 0)
	for {
		rd, err := iter.RowData()
		if err != nil {
			panic(err)
		}
		if !iter.Scan(rd.Values...) {
			break
		}

		rowData := make([]interface{}, 0)
		for _, val := range rd.Values {
			rowData = append(rowData, reflect.Indirect(reflect.ValueOf(val)).Interface())
		}

		data = append(data, rowData)
	}

	res := Result{}

	res.Columns = columns
	res.Data = data

	return res
}
