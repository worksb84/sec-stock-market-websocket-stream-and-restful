package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	polygonws "github.com/polygon-io/client-go/websocket"
	modelsws "github.com/polygon-io/client-go/websocket/models"
)

func (r *Requester) RunSubscribe(feed polygonws.Feed, wg *sync.WaitGroup) {
	ticker := time.NewTicker(time.Millisecond * 1200)
	c, err := polygonws.New(polygonws.Config{
		APIKey: r.ApiKey,
		Feed:   feed,
		Market: polygonws.Stocks,
	})
	if err != nil {
		log.Println(err)
	}
	defer c.Close()

	_ = c.Subscribe(polygonws.StocksTrades, "*")

	if err := c.Connect(); err != nil {
		log.Println(err)
		return
	}

	for {
		select {
		case <-ticker.C:
			if r.AvailableTime(0) && r.AvailableWeekday() {
				r.mu.RLock()
				b, _ := json.Marshal(r.Snapshot)
				wg.Add(1)
				go func() {
					r.Valkey.Save("STOCK:SEC:LIST", b, wg)
				}()

				for _, value := range r.Snapshot {
					b, _ := json.Marshal(value)
					wg.Add(1)
					go func() {
						key := fmt.Sprintf("STOCK:SEC:%s", value.S)
						r.Valkey.Save(key, b, wg)
					}()
				}
				r.mu.RUnlock()
			}
		case <-c.Error():
			return
		case out, more := <-c.Output():
			if !more {
				return
			}
			v := out.(modelsws.EquityTrade)
			r.mu.Lock()
			if value, exist := r.Snapshot[v.Symbol]; exist {
				low := value.Lp
				high := value.Hp

				if v.Price > high {
					high = v.Price
				}

				if v.Price < low {
					low = v.Price
				}

				value.Cp = float64(v.Price)
				value.Hp = high
				value.Lp = low
				value.Cr = value.C / value.Pcp * 100
				if value.Pcp == 0 {
					value.Cr = 0
				}
			}
			r.mu.Unlock()
		}
	}
}
