package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Response struct {
	Status  bool
	Message string
	Result  []Media
}

type Media struct {
	ContentUid    int16
	ContentId     string
	ContentName   string
	ContentLength int16
	StartDate     string
	EndDate       string
}

func main() {
	db, _ := sql.Open("mysql", "analyze:Vu4wj/3ej9ej9@tcp(172.16.51.107:3306)/analyze")
	defer db.Close()

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	ctx := context.Background()

	today := time.Now()
	fmt.Printf("Run time: %s\n", today.Format("2006-01-02 15:04:05"))

	// dev
	devContentData := getContentData("https://dcms.sp88.tw/api/ContentList/List")
	insertMediaData(ctx, db, devContentData)
	fmt.Printf("dev: %d \n", len(devContentData.Result))
}

func getContentData(url string) Response {
	response, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	var result Response
	if err := json.Unmarshal(body, &result); err != nil {
		log.Fatal(err)
	}

	return result
}

func insertMediaData(ctx context.Context, db *sql.DB, data Response) {
	for _, content := range data.Result {
		_, err := db.ExecContext(
			ctx,
			"INSERT INTO `medias` (`media_id`, `name`, `length`, `start_date`, `end_date`) VALUES (?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE name = ?, length =?, start_date = ?, end_date = ?",
			content.ContentId,
			content.ContentName,
			content.ContentLength,
			content.StartDate,
			content.EndDate,
			content.ContentName,
			content.ContentLength,
			content.StartDate,
			content.EndDate,
		)
		if err != nil {
			log.Fatal(err)
		}
	}
}
