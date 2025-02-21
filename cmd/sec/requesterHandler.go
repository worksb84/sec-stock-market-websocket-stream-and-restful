package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"pbm"
	"sync"
	"time"

	"sec/models"
)

var polygonUri = "https://api.polygon.io"
var apiUri = "http://0.0.0.0:8001"

func Api(method string, uri string, bytes io.Reader) (*http.Response, error) {
	req, _ := http.NewRequest(method, uri, bytes)
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	return client.Do(req)
}

func (r *Requester) GetTicker() {
	tickers := make([]models.Ticker, 0)

	uri := polygonUri + "/v3/reference/tickers?market=stocks&active=true&limit=1000&sort=ticker&apiKey=" + r.ApiKey
	next := true
	for {
		if !next {
			break
		}
		res, _ := Api("GET", uri, nil)

		var tickersBody models.Tickers
		if res.StatusCode == http.StatusOK {
			bytes, _ := io.ReadAll(res.Body)

			_ = json.Unmarshal(bytes, &tickersBody)
			tickers = append(tickers, tickersBody.Results...)
			if tickersBody.NextUrl != "" {
				uri = tickersBody.NextUrl + "&apiKey=" + r.ApiKey
			} else {
				next = false
			}
		}
	}

	r.mu.Lock()
	for _, v := range tickers {
		if v.PrimaryExchange == "XNAS" || v.PrimaryExchange == "XNYS" {
			exchange := "NYSE"
			if v.PrimaryExchange == "XNAS" {
				exchange = "NASDAQ"
			}
			r.Snapshot[v.Ticker] = &pbm.Snapshot{
				N:  v.Name,
				Ne: v.Name,
				S:  v.Ticker,
				E:  exchange,
			}
		}
	}
	r.mu.Unlock()

	uri = polygonUri + "/v2/snapshot/locale/us/markets/stocks/tickers?apiKey=" + r.ApiKey
	res, _ := Api("GET", uri, nil)

	var tickersSnapshot models.TickersSnapshot
	if res.StatusCode == http.StatusOK {
		bytes, _ := io.ReadAll(res.Body)

		err := json.Unmarshal(bytes, &tickersSnapshot)
		if err != nil {
			log.Println(err)
		}
	}

	r.mu.Lock()
	for _, v := range tickersSnapshot.Tickers {
		if value, exist := r.Snapshot[v.Ticker]; exist {
			value.Cp = float64(v.PrevDay.Close)
			value.Op = float64(v.PrevDay.Open)
			value.Hp = float64(v.PrevDay.High)
			value.Lp = float64(v.PrevDay.Low)
		}
	}
	r.mu.Unlock()
	log.Println("GetTicker")
}

func (r *Requester) GetPrevPrice() {
	snapshotLogs := r.GetSnapshotLogs()
	for k, v := range snapshotLogs {
		if value, exist := r.Snapshot[k]; exist {
			value.C = v.C
			value.Pcp = v.Pcp
			value.Cp = v.Cp
			value.Op = v.Op
			value.Hp = v.Hp
			value.Lp = v.Lp
			value.Cr = v.Cr
		}
	}
	log.Println("GetPrevPrice")
}

func (r *Requester) SetLog() {
	if r.AvailableWeekday() {
		r.GetTicker()

		uri := polygonUri + "/v2/snapshot/locale/us/markets/stocks/tickers?apiKey=" + r.ApiKey
		res, _ := Api("GET", uri, nil)

		var tickersSnapshot models.TickersSnapshot
		if res.StatusCode == http.StatusOK {
			bytes, _ := io.ReadAll(res.Body)

			err := json.Unmarshal(bytes, &tickersSnapshot)
			if err != nil {
				log.Println(err)
			}
		}

		r.mu.Lock()
		for _, v := range tickersSnapshot.Tickers {
			if value, exist := r.Snapshot[v.Ticker]; exist {
				value.Pcp = float64(v.PrevDay.Close)
				value.Cp = float64(v.Day.Close)
				value.Op = float64(v.Day.Open)
				value.Hp = float64(v.Day.High)
				value.Lp = float64(v.Day.Low)
				value.C = float64(v.TodaysChange)
				value.Cr = v.TodaysChangePerc
			}
		}
		r.mu.Unlock()
		log.Println("SetLog")
		r.SetSnapshot()
	}
}

func (r *Requester) SetRepair() {
	if !r.AvailableTime(15) || !r.AvailableWeekday() {
		r.SetRepairSnapshot()
	}
}

func (r *Requester) AvailableTime(minute int) bool {
	location, _ := time.LoadLocation("America/New_York")
	time.Local = location

	now := time.Now()
	startTime := time.Date(now.Year(), now.Month(), now.Day(), 9, 30+minute, 0, 0, time.Local)
	endTime := time.Date(now.Year(), now.Month(), now.Day(), 16, minute, 0, 0, time.Local)
	return now.After(startTime) && now.Before(endTime)
}

func (r *Requester) AvailableWeekday() bool {
	location, _ := time.LoadLocation("America/New_York")
	time.Local = location

	now := time.Now().Local()
	d := int(now.Weekday())

	uri := polygonUri + "/v1/marketstatus/upcoming?apiKey=" + r.ApiKey
	res, _ := Api("GET", uri, nil)

	var marketHoliday []models.MarketHoliday
	if res.StatusCode == http.StatusOK {
		bytes, _ := io.ReadAll(res.Body)

		err := json.Unmarshal(bytes, &marketHoliday)
		if err != nil {
			log.Println(err)
		}
	}

	isNotHolyday := true

	for _, v := range marketHoliday {
		if time.Time(v.Date).Format("2006-01-02") == now.Format("2006-01-02") {
			isNotHolyday = false
			break
		}
	}

	return (d != 0 && d != 6 && isNotHolyday)
}

func (r *Requester) GetSnapshotLogs() map[string]*pbm.Snapshot {
	body := &pbm.ReqResSnapshotLogs{
		Region: "SEC",
	}

	pbytes, _ := json.Marshal(body)
	buff := bytes.NewBuffer(pbytes)

	uri := apiUri + "/v1/snapshotLogs/"
	res, _ := Api("POST", uri, buff)

	if res.StatusCode == http.StatusOK {
		bytes, _ := io.ReadAll(res.Body)

		_ = json.Unmarshal(bytes, body)
	}
	snapshotLog := make(map[string]*pbm.Snapshot)
	_ = json.Unmarshal([]byte(body.Snapshot), &snapshotLog)

	return snapshotLog
}

func (r *Requester) SetSnapshot() {
	r.mu.Lock()
	b, _ := json.Marshal(r.Snapshot)
	r.mu.Unlock()

	reqResSnapshotLogs := &pbm.ReqResSnapshotLogs{
		Snapshot: string(b),
		Region:   "SEC",
	}

	pbytes, _ := json.Marshal(reqResSnapshotLogs)
	buff := bytes.NewBuffer(pbytes)

	uri := apiUri + "/v1/snapshotLogs/add"

	_, _ = Api("POST", uri, buff)

	log.Println("SetSnapshot")
}

func (r *Requester) SetRepairSnapshot() {
	r.Snapshot = r.GetSnapshotLogs()

	var wg sync.WaitGroup
	r.mu.Lock()
	b, _ := json.Marshal(r.Snapshot)
	wg.Add(1)
	go func() {
		r.Valkey.Save("STOCK:SEC:LIST", b, &wg)
	}()
	wg.Add(1)
	go func() {
		r.Valkey.Save("STOCK:SEC:DELAYED:LIST", b, &wg)
	}()

	for _, value := range r.Snapshot {
		b, _ := json.Marshal(value)
		wg.Add(1)
		go func() {
			key := fmt.Sprintf("STOCK:SEC:%s", value.S)
			r.Valkey.Save(key, b, &wg)
		}()
		wg.Add(1)
		go func() {
			key := fmt.Sprintf("STOCK:SEC:DELAYED:%s", value.S)
			r.Valkey.Save(key, b, &wg)
		}()
	}
	r.mu.Unlock()

	go func() {
		wg.Wait()
	}()
	log.Println("SetRepairSnapshot")
}
