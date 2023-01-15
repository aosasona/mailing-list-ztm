package jsonapi

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	mdb "mailing-list/pkg/db"
	"net/http"
)

func setJSONHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
}

func fromJSON[T any](body io.Reader, target T) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(body)
	json.Unmarshal(buf.Bytes(), &target)
}

func returnJSON[T any](w http.ResponseWriter, withData func() (T, error)) {
	setJSONHeader(w)

	data, serverError := withData()

	if serverError != nil {
		w.WriteHeader(500)
		serverErrorJSON, err := json.Marshal(&serverError)
		if err != nil {
			log.Print(err)
			return
		}
		w.Write(serverErrorJSON)
		return
	}

	dataJSON, err := json.Marshal(&data)
	if err != nil {
		log.Print(err)
		w.WriteHeader(500)
		return
	}

	w.Write(dataJSON)
}

func returnErr(w http.ResponseWriter, err error, code int) {
	returnJSON(w, func() (interface{}, error) {
		errorMsg := struct {
			Err string
		}{
			Err: err.Error(),
		}
		w.WriteHeader(code)
		return errorMsg, nil
	})
}

func CreateEmail(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			return
		}

		entry := mdb.EmailEntry{}
		fromJSON(r.Body, &entry)

		if err := mdb.CreateEmail(db, entry.Email); err != nil {
			returnErr(w, err, 400)
			return
		}

		returnJSON(w, func() (interface{}, error) {
			log.Printf("CreateEmail: %v\n", entry.Email)
			return mdb.GetEmail(db, entry.Email)
		})
	})
}

func GetEmail(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			return
		}

		entry := mdb.EmailEntry{}
		fromJSON(r.Body, &entry)

		returnJSON(w, func() (interface{}, error) {
			log.Printf("GetEmail: %v\n", entry.Email)
			return mdb.GetEmail(db, entry.Email)
		})
	})
}

func UpdateEmail(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			return
		}

		entry := mdb.EmailEntry{}
		fromJSON(r.Body, &entry)

		if err := mdb.UpdateEmail(db, entry); err != nil {
			returnErr(w, err, 400)
			return
		}

		returnJSON(w, func() (interface{}, error) {
			log.Printf("UpdateEmail: %v\n", entry.Email)
			return mdb.GetEmail(db, entry.Email)
		})
	})
}

func DeleteEmail(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			return
		}

		entry := mdb.EmailEntry{}
		fromJSON(r.Body, &entry)

		if err := mdb.DeleteEmail(db, entry.Email); err != nil {
			returnErr(w, err, 400)
			return
		}

		returnJSON(w, func() (interface{}, error) {
			log.Printf("DeleteEmail: %v\n", entry.Email)
			return mdb.GetEmail(db, entry.Email)
		})
	})
}

func GetEmails(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			return
		}

		queryOptions := mdb.GetEmailsParams{}
		fromJSON(r.Body, &queryOptions)

		if queryOptions.Count <= 0 || queryOptions.Page <= 0 {
			returnErr(
				w,
				errors.New("page and count fields are required and must be greater than 0"),
				400,
			)
			return
		}

		returnJSON(w, func() (interface{}, error) {
			log.Printf("GetEmails: %v\n", queryOptions)
			return mdb.GetEmails(db, queryOptions)
		})
	})
}

func Serve(db *sql.DB, bind string) {
	http.Handle("/email/create", CreateEmail(db))
	http.Handle("/email/get", GetEmail(db))
	http.Handle("/email/get_batch", GetEmails(db))
	http.Handle("/email/update", UpdateEmail(db))
	http.Handle("/email/delete", DeleteEmail(db))

	log.Printf("JSON API listening on %v\n", bind)

	err := http.ListenAndServe(bind, nil)
	if err != nil {
		log.Fatalf("JSON server error: %v", err)
	}
}
