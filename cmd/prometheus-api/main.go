package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type JobParams struct {
	Targets []string
}

type JobContext struct {
	Targets []string `json:"targets"`
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

	r := mux.NewRouter()

	r.HandleFunc("/configuration/jobs/{name}", handler).Methods("POST")

	srv := &http.Server{
		Addr:    os.Getenv("HOST") + ":" + os.Getenv("POST"),
		Handler: r,
	}

	log.Println("[info] Listening on " + os.Getenv("HOST") + ":" + os.Getenv("POST"))

	log.Fatal(srv.ListenAndServe())
}

func handler(w http.ResponseWriter, r *http.Request) {
	// recover
	defer func() {
		r := recover()
		if r != nil {
			log.Println("[Recover]", r)

			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("{\"status\": \"failed\", \"message\": \"" + r.(string) + "\"}"))
		}
	}()

	// query
	vars := mux.Vars(r)

	if !strings.Contains(os.Getenv("JOBS_ITEMS"), vars["name"]) {
		panic("Only setting " + os.Getenv("JOBS_ITEMS") + " jobs, input name:" + vars["name"])
	}

	// body
	body, _ := ioutil.ReadAll(r.Body)
	var params JobParams
	json.Unmarshal(body, &params)

	// context
	var contexts []JobContext
	context := JobContext{}
	context.Targets = append(context.Targets, params.Targets...)
	contexts = append(contexts, context)
	b, _ := json.Marshal(contexts)

	// write
	os.WriteFile(os.Getenv("JOB_TARGET_PATH")+"/"+vars["name"]+"-targets.json", b, 0644)

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("{\"status\": \"ok\"}"))
}
