package main

import (
	"encoding/hex"
	"log"
	"oc-go/internal/crypt"
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
		log.Fatalf("Error loading .env file")
	}

	// date
	currentDate := time.Now().Format("20060102")

	// crypt
	key, _ := hex.DecodeString(os.Getenv("KEY"))

	// nats
	nc, _ := nats.Connect(nats.DefaultURL)
	js, _ := nc.JetStream()

	js.QueueSubscribe(os.Getenv("NATS_SUBJ"), os.Getenv("NATS_QUEUE"), func(msg *nats.Msg) {
		// os
		f, err := os.OpenFile(os.Getenv("LOG_PATH")+"/decrypt."+currentDate+".log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf(err.Error())
		}

		content := string(msg.Data)

		// decrypt
		cipher, _ := hex.DecodeString(content)
		plainText := crypt.Decrypt(cipher, key)

		// append write
		f.WriteString(string(plainText) + " \n")

		f.Close()
		msg.Ack()
	})

	runtime.Goexit()
}
