package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/StonerF/posts/internal/config"
	"github.com/StonerF/posts/internal/graph"
	"github.com/StonerF/posts/internal/resolver"
	"github.com/StonerF/posts/internal/storage"
	"github.com/StonerF/posts/internal/storage/inmemory"
	"github.com/StonerF/posts/internal/storage/postgres"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5"
)

const port = "8888"

func main() {

	Cfg, err := config.MustloadConfig()
	if err != nil {
		log.Fatalln("Can't load config", err)
	}
	fmt.Println("Loaded config", Cfg)
	var repository storage.Repository
	if Cfg.Storage == "PostgreSQL" {
		fmt.Println("Connstring", fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable", Cfg.DB_User, Cfg.DB_Password, Cfg.DB_Name, Cfg.DB_Host, Cfg.DB_Port))
		db, err := pgx.Connect(context.Background(), fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable", Cfg.DB_User, Cfg.DB_Password, Cfg.DB_Name, Cfg.DB_Host, Cfg.DB_Port))

		if err != nil {
			log.Fatalln("Can`t connect to db", err)
		}
		repository = postgres.NewPostgresRep(db)
	}
	if Cfg.Storage == "IN_MEMORY" {
		repository = inmemory.NewInMemoryRepository()
	}

	if repository == nil {
		log.Fatal("repository nil error")
	}

	Resolver := resolver.NewResolver(repository)

	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: Resolver}))

	srv.AddTransport(&transport.Websocket{
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		KeepAlivePingInterval: 10 * time.Second,
	})

	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.MultipartForm{})

	srv.Use(extension.Introspection{})

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
