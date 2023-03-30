package main

import (
	"context"
	"github.com/NatSau02/go-proxy/app/proxy"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	server := proxy.NewServerProxy()
	http.HandleFunc("/", server.ServeHTTP)

	//
	//go func() {
	//	err := app.Listen(":8089")
	//	if err != nil {
	//		logrus.Fatalf("error occured while running http server: %s", err.Error())
	//	}
	//}()

	//	router := mux.NewRouter()
	//r := chi.NewRouter()

	//srv := &http.Server{
	//	Handler:      router,
	//	Addr:         "127.0.0.1:8000",
	//	WriteTimeout: 15 * time.Second,
	//	ReadTimeout:  15 * time.Second,
	//}

	//proxyS := goproxy.NewProxyHttpServer()
	//proxyS.Verbose = true
	//
	//proxyS.NonproxyHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	//	if req.Host == "" {
	//		fmt.Fprintln(w, "Cannot handle requests without Host header, e.g., HTTP 1.0")
	//		return
	//	}
	//	req.URL.Scheme = "http"
	//	req.URL.Host = req.Host
	//	proxyS.ServeHTTP(w, req)
	//})

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

	logrus.Print("proxy server started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logrus.Println("shutting down")

	if err := srv.Shutdown(context.Background()); err != nil {
		logrus.Errorf("error occured on server shutting down: %s", err.Error())
	}
}
