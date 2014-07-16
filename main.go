package main

import (
	"flag"
	"log"
	"net/http"
	"time"
)

type MyHandler struct{}

func (*MyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// log.Println(r.RequestURI, "Serve")
	time.Sleep(1e3)
}

var port = ":9090"

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	max_per_second_msg = *flag.Int("c", 3, "max message count")
	period_second_num = *flag.Int("m", 1, "measure the unit(second)")
	port = *flag.String("p", ":9090", "listen http port")
}

func main() {
	flag.Parse()
	httpSrv := http.Server{
		Addr:           port,
		Handler:        &MyHandler{},
		ReadTimeout:    time.Duration(time.Second * time.Duration(30)),
		MaxHeaderBytes: 1,
		ConnState:      onConnState}

	go func() {
		log.Printf("Start Server Listening %v", httpSrv.Addr)
		err := httpSrv.ListenAndServe()
		if err != nil {
			log.Fatalf("ListenAndServe: %v", err)
		}
	}()
	select {}
}
