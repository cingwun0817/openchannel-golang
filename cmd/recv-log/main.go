package main

import (
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"oc-go/internal/crypt"
	"oc-go/internal/generate"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/denisbrodbeck/machineid"
	"github.com/joho/godotenv"
)

type Log struct {
	Type    string `json:"type"`
	Date    string `json:"date"`
	Content string `json:"content"`
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		log.Fatalf("[Error] Not setting conf file")
	}

	// load env
	err := godotenv.Load(args[0])

	if err != nil {
		log.Fatalf("[Error] Error loading .env file")
	}

	log.Println("[Info] Loaded configuration file")

	// listen
	listener, err := net.Listen("tcp", os.Getenv("HOST")+":"+os.Getenv("POST"))
	if err != nil {
		log.Fatalf(err.Error())
	}
	defer listener.Close()

	log.Println("[Info] Listening on " + os.Getenv("HOST") + ":" + os.Getenv("POST"))
	log.Println("[Info] Log path " + os.Getenv("LOG_PATH"))

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf(err.Error())
		}

		go handleRequest(conn)

		go archiveExpiredFiles()
	}
}

func handleRequest(conn net.Conn) {
	buf := make([]byte, 16384)

	reqLen, err := conn.Read(buf)
	if err != nil {
		log.Fatalf(err.Error())
	}

	text := buf[1:]

	// prefix
	prefixName := "edge"
	encrypt := false
	if string(buf[0:1]) == "1" {
		prefixName = "enedge"
		encrypt = true
	}

	// date
	currentDate := time.Now().Format("20060102")

	// key
	key, _ := hex.DecodeString(generate.GetKey(currentDate))

	// machine id
	mId, err := machineid.ID()
	if err != nil {
		log.Fatalf(err.Error())
	}

	if reqLen != 0 {
		// os
		f, err := os.OpenFile(os.Getenv("LOG_PATH")+"/"+prefixName+"."+currentDate+".log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf(err.Error())
		}

		content := string(text)
		wText := content[0:len(content)-1] + `,"machine_id":"` + mId + `"}`

		if encrypt {
			// encrypt
			cipher := crypt.Encrypt([]byte(wText), key)
			cipherText := hex.EncodeToString(cipher)

			// json
			log := Log{Type: "encrypt", Date: currentDate, Content: cipherText}
			jsonLog, _ := json.Marshal(log)

			wText = string(jsonLog)
		}

		// append write
		f.WriteString(wText + " \n")

		f.Close()
	}

	// response
	conn.Write([]byte("ok"))

	conn.Close()
}

func archiveExpiredFiles() {
	expiryDate := time.Now().AddDate(0, 0, -14).Format("20060102")

	files, err := ioutil.ReadDir(os.Getenv("LOG_PATH"))
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		fileName := file.Name()

		firstPos := strings.Index(fileName, ".")

		if !file.IsDir() {
			prefix := fileName[0:firstPos]

			if prefix == "edge" || prefix == "enedge" {
				date := fileName[firstPos+1 : firstPos+9]

				intDate, _ := strconv.Atoi(date)
				intExpiryDate, _ := strconv.Atoi(expiryDate)

				if intDate-intExpiryDate < 0 {
					os.Remove(os.Getenv("LOG_PATH") + "/" + fileName)

					log.Println("[Info] remove " + fileName)
				}
			}
		}
	}
}
