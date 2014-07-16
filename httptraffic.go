package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

var (
	period_second_num  = 1
	max_per_second_msg = 3 // max period second message number
)

type httpTraffic struct {
	conn        net.Conn
	connectTime time.Time
	packetCount int
	lastNTime   time.Time
	lastNCount  int
}

var ConnMan = &ConnManager{make(map[string]*httpTraffic, 10240), sync.Mutex{}}

func onConnState(c net.Conn, s http.ConnState) {
	remoteAddr := c.RemoteAddr().String()
	switch s {
	case http.StateNew:
		ConnMan.NewConn(remoteAddr, c)
	case http.StateActive:
		traffic, err := ConnMan.GetConn(remoteAddr)
		if err != nil {
			log.Println("Error Get", remoteAddr)
		} else {
			timeNow := time.Now()
			traffic.packetCount++
			traffic.lastNCount++
			if traffic.lastNCount >= max_per_second_msg {
				elapseSecond := timeNow.Second() - traffic.lastNTime.Second()
				if elapseSecond >= period_second_num || elapseSecond == 0 {
					log.Println("Killed User", remoteAddr)
					c.Close()
					return
				}
				traffic.lastNTime = timeNow
				traffic.lastNCount = 0
			}
		}
	case http.StateClosed:
		ConnMan.RemoveConn(remoteAddr)
	}
}

type ConnManager struct {
	connMap map[string]*httpTraffic
	mu      sync.Mutex
}

func (this *ConnManager) NewConn(remoteAddr string, c net.Conn) (err error) {
	this.mu.Lock()
	if _, ok := this.connMap[remoteAddr]; ok {
		err = fmt.Errorf("RemoteAddr %v existed", remoteAddr)
	} else {
		now := time.Now()
		this.connMap[remoteAddr] = &httpTraffic{
			conn:        c,
			connectTime: now,
			packetCount: 0,
			lastNTime:   now,
			lastNCount:  0}
		log.Println(remoteAddr, "connected")
	}
	this.mu.Unlock()
	return
}

func (this *ConnManager) GetConn(remoteAddr string) (c *httpTraffic, err error) {
	this.mu.Lock()
	var ok bool
	if c, ok = this.connMap[remoteAddr]; !ok {
		err = fmt.Errorf("RemoteAddr %v not existed", remoteAddr)
	}
	this.mu.Unlock()
	return
}

func (this *ConnManager) RemoveConn(remoteAddr string) (c *httpTraffic, err error) {
	this.mu.Lock()
	var ok bool
	if c, ok = this.connMap[remoteAddr]; !ok {
		err = fmt.Errorf("RemoteAddr %v not existed", remoteAddr)
	}
	log.Printf("%s disconnected, packet count: %d online second: %v\n", remoteAddr, c.packetCount, time.Now().Sub(c.connectTime).Seconds())

	this.mu.Unlock()
	return
}

func (this *ConnManager) Count() (num int) {
	this.mu.Lock()
	num = len(this.connMap)
	this.mu.Unlock()
	return
}
