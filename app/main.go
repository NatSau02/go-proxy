package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NatSau02/go-proxy/app/proxy"
	"github.com/sirupsen/logrus"
)

func main() {

	cache := proxy.NewCacheResponse()
	server := proxy.NewServerProxy(cache)
	http.HandleFunc("/", server.ServeHTTP)

	srv := &http.Server{
		Handler:      server,
		Addr:         "127.0.0.1:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			logrus.Fatalf("error occured while running http server: %s", err.Error())
		}
	}()

	logrus.Print("start proxy server")

	go func() {
		for {
			time.Sleep(5 * time.Minute)
			cache.Update()
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logrus.Println("shutting down")

	if err := srv.Shutdown(context.Background()); err != nil {
		logrus.Errorf("error occured on server shutting down: %s", err.Error())
	}
}
