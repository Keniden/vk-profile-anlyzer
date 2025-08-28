package main

import (
	"fmt"
	"log"
	"net/http"
)

const (
	customPort = ":8080"
	defaultPort = ":9000"
)



func main() {
	go func ()  {
		customMux()
	}()
	go func(){defaultMux()}()
}


func customMux(){
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("custom mux")
	})
	if err := http.ListenAndServe(defaultPort, mux); err != nil {
		log.Fatal(err)
	}
}

func defaultMux(){
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("hello def mux")
	})
	http.ListenAndServe(customPort, nil)
}

