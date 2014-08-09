package httptraffic

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

var (
	PeriodSecondNum = 1
	MaxPerSecondMsg = 3 // max period second message number
)

type httpTraffic struct {
	conn        net.Conn
	connectTime time.Time
	packetCount int
	lastNTime   time.Time
	lastNCount  int
}

var connMan = &connManager{connMap: make(map[string]*httpTraffic, 16<<10)}

var OnKilled func(c net.Conn)

func OnConnState(c net.Conn, s http.ConnState) {
	remoteAddr := c.RemoteAddr().String()
	switch s {
	case http.StateNew:
		connMan.NewConn(remoteAddr, c)
	case http.StateActive:
		traffic, err := connMan.GetConn(remoteAddr)
		if err != nil {
			// log.Println("Error Get", remoteAddr)
			panic(err)
		} else {
			timeNow := time.Now()
			traffic.packetCount++
			traffic.lastNCount++
			if traffic.lastNCount >= MaxPerSecondMsg {
				elapseSecond := timeNow.Second() - traffic.lastNTime.Second()
				if elapseSecond >= PeriodSecondNum || elapseSecond == 0 {
					// log.Println("Killed User", remoteAddr)
					c.Close()
					if OnKilled != nil {
						OnKilled(c)
					}
					return
				}
				traffic.lastNTime = timeNow
				traffic.lastNCount = 0
			}
		}
	case http.StateClosed:
		connMan.RemoveConn(remoteAddr)
	}
}

type connManager struct {
	connMap map[string]*httpTraffic
	sync.Mutex
}

func (this *connManager) NewConn(remoteAddr string, c net.Conn) (err error) {
	this.Lock()
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
		// log.Println(remoteAddr, "connected")
	}
	this.Unlock()
	return
}

func (this *connManager) GetConn(remoteAddr string) (c *httpTraffic, err error) {
	this.Lock()
	var ok bool
	if c, ok = this.connMap[remoteAddr]; !ok {
		err = fmt.Errorf("RemoteAddr %v not existed", remoteAddr)
	}
	this.Unlock()
	return
}

func (this *connManager) RemoveConn(remoteAddr string) (c *httpTraffic, err error) {
	this.Lock()
	var ok bool
	if c, ok = this.connMap[remoteAddr]; !ok {
		err = fmt.Errorf("RemoteAddr %v not existed", remoteAddr)
	}
	// log.Printf("%s disconnected, packet count: %d online second: %v\n", remoteAddr, c.packetCount, time.Now().Sub(c.connectTime).Seconds())
	this.Unlock()
	return
}

func (this *connManager) Count() (num int) {
	this.Lock()
	num = len(this.connMap)
	this.Unlock()
	return
}
