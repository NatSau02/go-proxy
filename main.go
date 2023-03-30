package main

import (
	"bufio"
	"bytes"
	"context"
	"github.com/NatSau02/go-proxy/app/proxy"
	"github.com/sirupsen/logrus"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// set to false, if you don't want the body to be cached
const dumpBody = true

func main() {
	server := proxy.NewServerProxy()
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		req, err := http.NewRequest("GET", "https://www.google.com/", nil)

		if err != nil {
			log.Fatal(err)
		}

		req.AddCookie(&http.Cookie{Name: "c", Value: "ccc"})

		resp, err := http.DefaultClient.Do(req)

		if err != nil {
			log.Fatal(err)
		}

		body, err := httputil.DumpResponse(resp, dumpBody)

		if err != nil {
			log.Fatal(err)
		}

		r := bufio.NewReader(bytes.NewReader(body))

		resp, err = http.ReadResponse(r, nil)

		if err != nil {
			log.Fatal(err)
		}

		io.Copy(writer, resp.Body)
	})

	srv := &http.Server{
		Handler:      server,
		Addr:         "127.0.0.1:8006",
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
