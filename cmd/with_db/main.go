package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"github.com/gorilla/mux"
	"example/internal/pkg/db"
)

const port = ":8080"
const queryParamKey = "key"

type server1 struct {
	data map[string]string
}

type requestCustom struct {
	Key   string
	Value string
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	database, err := db.NewDB(ctx)
	if err != nil {
		log.Fatal(err)
	}

	defer database.GetPool(ctx).Close()
	
	implementation := server1{
		data: make(map[string]string)}

	http.Handle("/", createRouter(implementation))

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}

func createRouter(implementation server1) *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/article", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			implementation.Create(w, r)
		case http.MethodPut:
			implementation.Update(w, r)
		default:
			fmt.Println("error")
		}
	})

	router.HandleFunc(fmt.Sprintf("/article/{%s:[A-z]+}", queryParamKey), func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			implementation.Get(w, r)
		case http.MethodDelete:
			implementation.Delete(w, r)
		default:
			fmt.Println("error")
		}
	})
	return router
}

func (s *server1) Update(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Update")
}

func (s *server1) Delete(w http.ResponseWriter, r *http.Request) {
	key, ok := mux.Vars(r)[queryParamKey]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	_, ok = s.data[key]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	delete(s.data, key)
}

func (s *server1) Get(w http.ResponseWriter, r *http.Request) {
	key, ok := mux.Vars(r)[queryParamKey]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	value, ok := s.data[key]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if _, err := w.Write([]byte(value)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
func (s *server1) Create(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var unm requestCustom
	if err := json.Unmarshal(body, &unm); err != nil{
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if unm.Key == "" || unm.Value == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if _, ok := s.data[unm.Key]; ok{
		w.WriteHeader(http.StatusConflict)
		return
	}
	s.data[unm.Key] = unm.Value
}
