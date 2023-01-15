package main

import (
	"database/sql"
	"log"
	mdb "mailing-list/pkg/db"
	"mailing-list/pkg/grpcapi"
	"mailing-list/pkg/jsonapi"
	"sync"

	"github.com/alexflint/go-arg"
)

var args struct {
	DbPath   string `arg:"env:MAILINGLST_DB"`
	BindJSON string `arg:"env:MAILINGLST_BIND_JSON"`
	BindGRPC string `arg:"env:MAILINGLST_BIND_GRPC"`
}

func main() {
	arg.MustParse(&args)

	if args.DbPath == "" {
		args.DbPath = "data.sqlite"
	}

	if args.BindJSON == "" {
		args.BindJSON = ":8080"
	}

	if args.BindGRPC == "" {
		args.BindGRPC = ":8081"
	}

	log.Printf("using database '%v'", args.DbPath)

	db, err := sql.Open("sqlite3", args.DbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	mdb.TryCreate(db)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		log.Printf("Starting JSON API")
		jsonapi.Serve(db, args.BindJSON)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		log.Printf("Starting GRPC API")
		grpcapi.Serve(db, args.BindGRPC)
		wg.Done()
	}()

	wg.Wait()
}
