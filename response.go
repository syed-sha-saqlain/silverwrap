package silverwrap

import (
	"encoding/csv"
	"encoding/json"
	"net/http"
)

func WriteJson(w http.ResponseWriter, code int, data interface{}) error {
	w.Header().Set(ContentType, MimeApplicationJSON)
	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(data)
}

func WriteCSV(w http.ResponseWriter, code int, data [][]string) error {
	w.Header().Set(ContentType, MimeTextCSV)

	writer := csv.NewWriter(w)
	defer writer.Flush()

	w.WriteHeader(code)
	return writer.WriteAll(data)
}
