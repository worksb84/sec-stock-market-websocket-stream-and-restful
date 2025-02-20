package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	clientName string
	address    string
	port       string
	db         int

	uri    string
	apiKey string
)

func main() {
	// runtime.GOMAXPROCS(2)
	flag.StringVar(&clientName, "clientName", "", "clientName")
	flag.StringVar(&address, "address", "", "address")
	flag.StringVar(&port, "port", "", "port")
	flag.IntVar(&db, "db", 0, "db")

	flag.StringVar(&uri, "Uri", "", "uri")
	flag.StringVar(&apiKey, "apiKey", "", "apiKey")
	flag.Parse()

	valkeyClient := NewValkeyClient(
		Config{
			Address:    address,
			Port:       port,
			DB:         db,
			ClientName: clientName,
		},
	)

	requester := NewRequester(valkeyClient, uri, apiKey)

	go requester.Run()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	<-ctx.Done()
}
