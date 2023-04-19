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
	newProxy := proxy.NewServerProxy(cache)
	reverseProxy, err := newProxy.ConfigureProxy("")
	if err != nil {
		panic(err)
	}

	// handle all requests to your server using the proxy
	http.HandleFunc("/", newProxy.ProxyRequestHandler(reverseProxy))

	srv := &http.Server{
		Handler:      reverseProxy,
		Addr:         "127.0.0.1:8088",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			logrus.Fatalf("error occured while running http server: %s", err.Error())
		}
	}()

	logrus.Print("proxy server started")

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
