package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/m4rw3r/uuid"
)

var thoughtMap map[ThoughtsID]Thoughts

func init() {
	logrus.Info("Intiliazing the data structures.")
	thoughtMap = make(map[ThoughtsID]Thoughts)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9999"
	}

	r := new(mux.Router)
	r.HandleFunc("/api/ping", pingHandler)
	r.HandleFunc("/api/thoughts", thoughtsPostHandler).Methods("POST")
	r.HandleFunc("/api/thoughts", getAllThoughts).Methods("GET")
	r.HandleFunc("/api/thoughts/{id}", thoughtsGetHandler).Methods("GET")
	r.HandleFunc("/api/thoughts/{id}", editThoughtsHandler).Methods("PUT")

	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("public/"))))
	http.ListenAndServe(":"+port, r)
}

type Status struct {
	Status  string
	Service string
}

// Handles checking the basic availability of the application.
func pingHandler(rw http.ResponseWriter, r *http.Request) {

	response := Status{"pong", "scatter-brain"}
	resp, _ := json.Marshal(response)
	rw.Write(resp)
}

func editThoughtsHandler(rw http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.FromString(idStr)
	if err != nil {
		httpError(rw, http.StatusBadRequest, errors.New("unable to parse the identifier."))
		return
	}

	var thought Thoughts
	if _, ok := thoughtMap[ThoughtsID{id}]; !ok {
		httpError(rw, http.StatusNotFound, errors.New("unable to locate resource."))
		return
	}

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&thought)
	if err != nil {
		httpError(rw, http.StatusBadRequest, errors.New("unable to parse the resource"))
		return
	}

	thoughtMap[ThoughtsID{id}] = thought
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusNoContent)
}

func getAllThoughts(rw http.ResponseWriter, r *http.Request) {

	thoughts := make([]Thoughts, 0)
	for key := range thoughtMap {
		thoughts = append(thoughts, thoughtMap[key])
	}

	resp, err := json.Marshal(thoughts)
	if err != nil {
		httpError(rw, http.StatusInternalServerError, err)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	rw.Write(resp)
}

// Handles the addition of a new user thought in the system.
func thoughtsPostHandler(rw http.ResponseWriter, r *http.Request) {

	logrus.Info("Adding a new thought to the system.")
	decoder := json.NewDecoder(r.Body)

	thoughtsPost := ThoughtsPost{}
	err := decoder.Decode(&thoughtsPost)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("Json decoding failed.")

		httpError(rw, http.StatusBadRequest, err)
		return
	}

	id, _ := uuid.V4()
	thoughtsID := ThoughtsID{id}

	// Create a new thoughts object.
	thoughts := Thoughts{
		ID:          thoughtsID,
		CreatedTime: time.Now(),
		Title:       thoughtsPost.Title,
		Content:     thoughtsPost.Thought,
	}

	// Store the value in the thoughts map.
	thoughtMap[thoughts.ID] = thoughts

	resp, err := json.Marshal(thoughts)
	if err != nil {
		httpError(rw, http.StatusInternalServerError, err)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusCreated)
	rw.Write(resp)
}

// Handles the fetch of a particular thought from the map for now.
func thoughtsGetHandler(rw http.ResponseWriter, r *http.Request) {

	logrus.Info("Received a call to fetch the thoughts")

	vars := mux.Vars(r)
	idString := vars["id"]
	id, err := uuid.FromString(idString)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("Invalid url param.")

		httpError(rw, http.StatusBadRequest, err)
		return
	}

	// If the entry is being located successfully.
	if thought, ok := thoughtMap[ThoughtsID{id}]; ok {

		resp, err := json.Marshal(thought)
		if err != nil {
			httpError(rw, http.StatusInternalServerError, err)
			return
		}

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		rw.Write(resp)
		return
	}

	httpError(rw, http.StatusNotFound, errors.New("Unable to locate the resource."))
}
