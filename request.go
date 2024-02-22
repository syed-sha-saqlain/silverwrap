package silverwrap

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	validator "github.com/go-playground/validator/v10"
	"github.com/gorilla/schema"
)

var (
	ErrMalformedRequest   = errors.New("malformed request")
	ErrInvalidContentType = errors.New("invalid content-type")
)

// HandlerType is the type of the business logic handler.
type HandlerType func(w http.ResponseWriter, r *http.Request) (int, any, error)

// HandleCall is the wrapper around business logic handler for dealing with boilerplate logic.
func HandleCall(w http.ResponseWriter, r *http.Request, data any, handler HandlerType) {

	if err := bind(r, data); err != nil {
		if err == ErrInvalidContentType {
			WriteJson(w, http.StatusNotAcceptable, err)
			return
		}

		WriteJson(w, http.StatusBadRequest, err)
		return
	}

	if data != nil {
		if err := validator.New().Struct(data); err != nil {
			WriteJson(w, http.StatusBadRequest, err)
			return
		}
	}

	code, res, err := handler(w, r)
	if err != nil {
		WriteJson(w, code, err)
		return
	}

	switch w.Header().Get(ContentType) {
	case MimeTextCSV:
		WriteCSV(w, code, res.([][]string))

	default:
		WriteJson(w, code, res)
	}
}

func bind(r *http.Request, data any) (err error) {
	if data == nil {
		return
	}

	switch r.Method {
	case http.MethodGet, http.MethodDelete, http.MethodHead:
		return bindQueryParams(r, data)
	default:
		return bindBody(r, data)
	}
}

func bindQueryParams(r *http.Request, data any) error {
	decoder := schema.NewDecoder()
	if err := decoder.Decode(data, r.URL.Query()); err != nil {
		return err
	}

	return nil
}

func bindBody(r *http.Request, data any) error {

	if r.ContentLength == 0 {
		return nil
	}

	contentType := strings.ToLower(r.Header.Get(ContentType))

	switch {
	case strings.Contains(contentType, MimeApplicationJSON):
		if err := bindJSON(r, data); err != nil {
			return err
		}
	case strings.Contains(contentType, "application/x-www-form-urlencoded"), strings.Contains(contentType, "multipart/form-data"):
		if err := bindFormData(r, data); err != nil {
			return err
		}

	default:
		return ErrInvalidContentType
	}

	return nil
}

func bindJSON(r *http.Request, data any) error {
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return err
	}

	return nil
}

func bindFormData(r *http.Request, data any) error {
	err := r.ParseMultipartForm(32 << 20) // 32 MB
	if err != nil {
		return err
	}

	// todo: decode passed files

	decoder := schema.NewDecoder()
	if err := decoder.Decode(data, r.Form); err != nil {
		return err
	}

	return nil
}
