package main

import (
	"flag"
	"github.com/cuixin/httptraffic"
	"log"
	"net"
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
	httptraffic.MaxPerSecondMsg = *flag.Int("c", 3, "max message count")
	httptraffic.PeriodSecondNum = *flag.Int("m", 1, "measure the unit(second)")
	port = *flag.String("p", ":9090", "listen http port")
}

func onKilled(c net.Conn) {
	log.Println(c.RemoteAddr(), "has been killed")
}

func main() {
	flag.Parse()
	httptraffic.OnKilled = onKilled
	httpSrv := http.Server{
		Addr:           port,
		Handler:        &MyHandler{},
		ReadTimeout:    time.Duration(time.Second * time.Duration(30)),
		MaxHeaderBytes: 1,
		ConnState:      httptraffic.OnConnState}

	go func() {
		log.Printf("Start Server Listening %v", httpSrv.Addr)
		err := httpSrv.ListenAndServe()
		if err != nil {
			log.Fatalf("ListenAndServe: %v", err)
		}
	}()
	select {}
}
