package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/gocql/gocql"
)

type Column struct {
	Name string
	Type string
}

type Response struct {
	Status  string        `json:"status"`
	Message string        `json:"message"`
	Columns []Column      `json:"columns"`
	Data    []interface{} `json:"data"`
}

func main() {
	// cluster config
	cluster := gocql.NewCluster("172.16.51.118", "172.16.51.120", "172.16.51.121")

	cluster.Keyspace = "oc"
	cluster.Consistency = gocql.Quorum
	cluster.ProtoVersion = 4
	cluster.Timeout = 10 * time.Second

	// session
	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	// context
	ctx := context.Background()

	res := query(
		ctx,
		session,
		"SELECT date, store_id, SUM(play_count) FROM oc.quividi_people_hour_analyze WHERE date = '2023-06-01' GROUP BY date, store_id",
	)

	fmt.Println(res)

	// output
	b, err := json.Marshal(res)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println(string(b))
}

func query(ctx context.Context, session *gocql.Session, stmt string) Response {
	// // recover
	// defer func() {
	// 	r := recover()
	// 	if r != nil {
	// 		log.Println("[Recover]", r)
	// 	}
	// }()

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
			// log.Fatal(err)
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

	res := Response{}

	res.Columns = columns
	res.Data = data

	return res
}
