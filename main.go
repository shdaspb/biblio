package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"biblio/internal/application"
	"biblio/internal/repository"

	"github.com/julienschmidt/httprouter"
)

func main() {
	ctx := context.Background()

	dbpool, err := repository.InitDBConn(ctx)
	if err != nil {
		log.Fatalf("%w failed to init DB connection", err)
	}
	defer dbpool.Close()

	a := application.NewApp(ctx, dbpool)
	r := httprouter.New()
	a.Routes(r)

	srv := &http.Server{Addr: "0.0.0.0:9090", Handler: r}
	fmt.Println("It is alive! Try http://localhost:9090")
	srv.ListenAndServe()
}
