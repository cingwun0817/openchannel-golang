package main

import (
	"encoding/hex"
	"encoding/json"
	"log"
	"oc-go/internal/crypt"
	"oc-go/internal/generate"
	"os"
	"runtime"
	"time"

	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
)

func main() {
	args := os.Args[1:]

	// load env
	err := godotenv.Load(args[0])

	if err != nil {
		panic("[godotenv load] " + err.Error())
	}

	log.Println("[info] Loaded configuration file")

	// nats
	nc, _ := nats.Connect(nats.DefaultURL)
	js, _ := nc.JetStream()

	log.Println("[info] Connect nats-server: " + nats.DefaultURL)

	js.QueueSubscribe(os.Getenv("NATS_SUBJ"), os.Getenv("NATS_QUEUE"), func(msg *nats.Msg) {
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

			msg.Ack()
		}()

		// os
		f, err := os.OpenFile(os.Getenv("LOG_PATH")+"/decrypt."+currentDate+".log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic("[os open file] " + err.Error())
		}

		// json
		var records [][]interface{}
		unmarshalErr := json.Unmarshal(msg.Data, &records)
		if unmarshalErr != nil {
			panic("[json unmarshal] " + unmarshalErr.Error() + ", string: " + string(msg.Data))
		}

		for _, record := range records {
			log := record[1].(map[string]interface{})

			// decrypt
			cipher, _ := hex.DecodeString(log["content"].(string))
			plainText := crypt.Decrypt(cipher, key)

			// append write
			f.WriteString(string(plainText) + " \n")
		}

		f.Close()
		msg.Ack()
	})

	runtime.Goexit()
}
