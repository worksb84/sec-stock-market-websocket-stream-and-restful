package main

import (
	"log"
	"pbm"
	"sync"
	"time"

	"github.com/go-co-op/gocron/v2"
	polygonws "github.com/polygon-io/client-go/websocket"
)

type Requester struct {
	mu              sync.RWMutex
	Valkey          *ValkeyClient
	Uri             string
	ApiKey          string
	Snapshot        map[string]*pbm.Snapshot
	DelayedSnapshot map[string]*pbm.Snapshot
}

func NewRequester(
	valkey *ValkeyClient,
	uri string,
	apiKey string) *Requester {
	return &Requester{
		Valkey:          valkey,
		Uri:             uri,
		ApiKey:          apiKey,
		Snapshot:        make(map[string]*pbm.Snapshot, 0),
		DelayedSnapshot: make(map[string]*pbm.Snapshot, 0),
	}
}

func (r *Requester) Run() {
	location, _ := time.LoadLocation("America/New_York")
	time.Local = location

	r.Initialize()

	s, _ := gocron.NewScheduler()
	_, err := s.NewJob(gocron.DurationJob(30000*time.Millisecond), gocron.NewTask(r.SetRepair))
	if err != nil {
		log.Println(err)
	}

	_, err = s.NewJob(gocron.DailyJob(1, gocron.NewAtTimes(gocron.NewAtTime(16, 15, 0))), gocron.NewTask(r.SetLog))
	if err != nil {
		log.Println(err)
	}

	s.Start()

	var wg sync.WaitGroup
	go r.RunSubscribe(polygonws.DelayedBusinessFeed, &wg)
	go r.RunSubscribe(polygonws.BusinessFeed, &wg)

	go func() {
		wg.Wait()
	}()
}

func (r *Requester) Initialize() {
	r.GetTicker()
	r.GetPrevPrice()
}
