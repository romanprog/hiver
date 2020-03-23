package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
)

func myInfo() string {
	return "hinfo2"
}

var listenPort = flag.String("listen.port",
	getEnv("LISTEN_PORT", "80"),
	"Listened port.")

func main() {
	//	package main
	flag.Parse()
	startListen()

}

func startListen() {
	router := mux.NewRouter()
	router.HandleFunc("/", httpHeloHendler).Methods("GET")
	listenUrl := fmt.Sprintf("0.0.0.0:%s", *listenPort)
	log.Printf("Runing listener on %s", listenUrl)
	log.Fatal(http.ListenAndServe(listenUrl, router))
}

func getEnv(key string, defaultVal string) string {
	if envVal, ok := os.LookupEnv(key); ok {
		return envVal
	}
	return defaultVal
}

func httpHeloHendler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request from host: %s", r.Host)
	file, _ := ioutil.ReadFile("/print_this.txt")
	output := fmt.Sprintf("%s\n", myInfo())
	output = fmt.Sprintf("%s=============== ENV ===============\n%s\n", output, strings.Join(os.Environ(), "\n"))
	output = fmt.Sprintf("%s============= /print_this.txt ==============\n%s\n", output, string(file))

	w.Write([]byte(output))
}
