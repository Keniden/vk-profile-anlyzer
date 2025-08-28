package main

import (
	"context"
	"encoding/json"
	"errors"
	"example/internal/pkg/db"
	"example/internal/pkg/repository"
	"example/internal/pkg/repository/postgresql"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

const port = ":8080"
const queryParamKey = "key"

type server1 struct {
	repo *postgresql.ArticleRepo
}

type addArticleRequest struct {
	Name   string `json:"name"`
	Rating int64  `json:"rating"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	database, err := db.NewDB(ctx)
	if err != nil {
		log.Fatal(err)
	}

	defer database.GetPool(ctx).Close()

	articleRepo := postgresql.NewArticles(database)

	implementation := server1{
		repo: articleRepo}

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

	router.HandleFunc(fmt.Sprintf("/article/{%s:[0-9]+}", queryParamKey), func(w http.ResponseWriter, r *http.Request) {
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
	// key, ok := mux.Vars(r)[queryParamKey]
	// if !ok {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	return
	// }
	// _, ok = s.data[key]
	// if !ok {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	return
	// }

	// delete(s.data, key)
}

func (s *server1) Get(w http.ResponseWriter, r *http.Request) {
	key, ok := mux.Vars(r)[queryParamKey]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// value, ok := s.data[key]
	// if !ok {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	return
	// }

	// if _, err := w.Write([]byte(value)); err != nil {
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }
	keyInt, err := strconv.ParseInt(key, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	article, err := s.repo.GetByID(r.Context(), keyInt)
	if err != nil {
		if errors.As(err, &repository.ErrObjNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	articleJson, err := json.Marshal(article)
	if err != nil {
		fmt.Println(articleJson)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(articleJson)
}

func (s *server1) Create(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var unm addArticleRequest
	if err := json.Unmarshal(body, &unm); err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	articleRepo := &repository.Article{
		Name: unm.Name, Rating: unm.Rating,
	}
	id, err := s.repo.Add(r.Context(), articleRepo)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	articleRepo.ID = int64(id)
	articleJson, _ := json.Marshal(articleRepo)
	w.Write(articleJson)

	// if unm.Name == "" {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	return
	// }

	// if _, ok := s.data[unm.Key]; ok {
	// 	w.WriteHeader(http.StatusConflict)
	// 	return
	// }
	// s.data[unm.Key] = unm.Value
}
