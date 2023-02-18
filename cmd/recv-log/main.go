package main

import (
	"encoding/hex"
	"log"
	"net"
	"oc-go/internal/crypt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	args := os.Args[1:]

	// load env
	err := godotenv.Load(args[0])

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// listen
	listener, err := net.Listen("tcp", os.Getenv("HOST")+":"+os.Getenv("POST"))
	if err != nil {
		log.Fatalf(err.Error())
	}
	defer listener.Close()

	log.Println("Listening on " + os.Getenv("HOST") + ":" + os.Getenv("POST"))

	// crypt
	key, _ := hex.DecodeString(os.Getenv("KEY"))

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf(err.Error())
		}

		go handleRequest(conn, key)
	}
}

func handleRequest(conn net.Conn, key []byte) {
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

	if reqLen != 0 {
		// date
		currentDate := time.Now().Format("20060102")

		// os
		f, err := os.OpenFile(os.Getenv("LOG_PATH")+"/"+prefixName+"."+currentDate+".log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf(err.Error())
		}

		wText := text
		if encrypt {
			// encrypt
			cipherText := crypt.Encrypt(text, key)

			wText = cipherText
		}

		// append write
		f.WriteString(string(wText) + " \n")

		f.Close()
	}

	// response
	conn.Write([]byte("ok"))

	conn.Close()
}
