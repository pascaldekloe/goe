package rest

import (
	"encoding/json"
	"fmt"
	"log"
	"mime"
	"net/http"
	"strconv"
)

var tailJSON = []byte{'\n'}

func serveJSON(w http.ResponseWriter, statusCode int, x interface{}) {
	bytes, err := json.MarshalIndent(x, "", "\t")
	if err != nil {
		log.Print("Can't serialize body: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h := w.Header()
	h.Set("Content-Type", "application/json;charset=UTF-8")
	h.Set("Content-Length", strconv.Itoa(len(bytes)+len(tailJSON)))
	w.WriteHeader(statusCode)

	if _, err := w.Write(bytes); err != nil {
		log.Print(err)
	}
	if _, err := w.Write(tailJSON); err != nil {
		log.Print(err)
	}
}

func receiveJSON(dst interface{}, r *http.Request, w http.ResponseWriter) bool {
	switch t, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); {
	case err != nil:
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return false
	case t != "application/json":
		http.Error(w, "want JSON", http.StatusUnsupportedMediaType)
		return false
	}

	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		http.Error(w, fmt.Sprintf("malformed request body: %s", err), http.StatusBadRequest)
		return false
	}

	return true
}
