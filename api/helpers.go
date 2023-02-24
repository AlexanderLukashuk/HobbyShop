package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"fainal.net/internal/validator"
	"github.com/julienschmidt/httprouter"
)

type envelope map[string]interface{}

func (app *Application) readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}
	return id, nil
}

func (app *Application) writeJSON(w http.ResponseWriter, status int, data interface{}, headers http.Header) error {
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "Application/json")
	w.WriteHeader(status)
	w.Write(js)
	return nil
}

func (app *Application) ReadJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	err := json.NewDecoder(r.Body).Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		if errors.As(err, &syntaxError) {
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)
		} else if errors.As(err, &unmarshalTypeError) {
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", unmarshalTypeError.Offset)

		} else if errors.As(err, &invalidUnmarshalError) {
			panic(err) //If our program reaches a point where it cannot be recovered due to some major errors

		} else if errors.Is(err, io.ErrUnexpectedEOF) {
			return errors.New("body contains badly-formed JSON")

		} else if errors.Is(err, io.EOF) {
			return errors.New("body must not be empty")

		} else {
			return err
		}
	}

	return nil
}
func (app *Application) readString(qs url.Values, key string, defaultValue string) string {

	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}
	return s
}

func (app *Application) readCSV(qs url.Values, key string, defaultValue []string) []string {

	csv := qs.Get(key)
	if csv == "" {
		return defaultValue
	}
	return strings.Split(csv, ",")
}

func (app *Application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {

	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}
	return i
}
func (app *Application) background(fn func()) {
	app.wg.Add(1)

	go func() {
		defer app.wg.Done()
		defer func() {
			if err := recover(); err != nil {
				app.logger.PrintError(fmt.Errorf("%s", err), nil)
			}
		}()
		fn()
	}()

}
