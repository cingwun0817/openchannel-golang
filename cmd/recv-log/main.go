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
		panic("[os args] not setting conf file")
	}

	// load env
	err := godotenv.Load(args[0])

	if err != nil {
		panic("[godotenv load] " + err.Error())
	}

	log.Println("[info] Loaded configuration file")

	// listen
	listener, err := net.Listen("tcp", os.Getenv("HOST")+":"+os.Getenv("POST"))
	if err != nil {
		panic("[nat listen] " + err.Error())
	}
	defer listener.Close()

	log.Println("[info] Listening on " + os.Getenv("HOST") + ":" + os.Getenv("POST"))
	log.Println("[info] Log path " + os.Getenv("LOG_PATH"))

	for {
		conn, err := listener.Accept()
		if err != nil {
			panic("[listener accept] " + err.Error())
		}

		go handleRequest(conn)

		go archiveExpiredFiles()
	}
}

func handleRequest(conn net.Conn) {
	// date
	currentDate := time.Now().Format("20060102")

	// key
	key, _ := hex.DecodeString(generate.GetKey(currentDate))

	// recover
	defer func() {
		r := recover()
		if r != nil {
			log.Println("[Recover]", r, ", key: ", key)
		}

		// response
		conn.Write([]byte("err"))

		conn.Close()
	}()

	buf := make([]byte, 65536)

	reqLen, err := conn.Read(buf)
	if err != nil {
		panic("[conn read] " + err.Error())
	}

	text := buf[1:reqLen]

	// prefix
	prefixName := "edge"
	encrypt := false
	if string(buf[0:1]) == "1" {
		prefixName = "enedge"
		encrypt = true
	}

	// machine id
	mId, err := machineid.ID()
	if err != nil {
		panic("[machine id] " + err.Error())
	}

	if reqLen != 0 {
		// os
		f, err := os.OpenFile(os.Getenv("LOG_PATH")+"/"+prefixName+"."+currentDate+".log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic("[open file] " + err.Error())
		}

		// add machine_id
		var jsonData map[string]interface{}
		unmarshalErr := json.Unmarshal(text, &jsonData)
		if unmarshalErr != nil {
			panic("[json unmarshal] " + unmarshalErr.Error() + ", string: " + string(text))
		}
		jsonData["machine_id"] = mId

		wText, err := json.Marshal(jsonData)
		if err != nil {
			panic("[json marshal] " + err.Error())
		}

		if encrypt {
			// encrypt
			cipher := crypt.Encrypt(wText, key)
			cipherText := hex.EncodeToString(cipher)

			// json
			log := Log{Type: "encrypt", Date: currentDate, Content: cipherText}
			jsonLog, _ := json.Marshal(log)

			wText = jsonLog
		}

		// append write
		f.WriteString(string(wText) + " \n")

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
		panic("[ioutil read dir] " + err.Error())
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

					log.Println("[info] remove " + fileName)
				}
			}
		}
	}
}
