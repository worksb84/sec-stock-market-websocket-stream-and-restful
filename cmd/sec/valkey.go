package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"sync"

	"github.com/valkey-io/valkey-go"
)

type Config struct {
	Address    string
	Port       string
	DB         int
	ClientName string
}

type ValkeyClient struct {
	val valkey.Client
}

func NewValkeyClient(config Config) *ValkeyClient {
	val, _ := valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{fmt.Sprintf("%s:%s", config.Address, config.Port)},
		TLSConfig:   &tls.Config{},
		SelectDB:    config.DB,
		ClientName:  config.ClientName,
	})

	valkeyClient := &ValkeyClient{
		val: val,
	}
	return valkeyClient
}

func (p *ValkeyClient) Save(key string, msg []byte, wg *sync.WaitGroup) {
	ctx1 := context.Background()
	if err := p.val.Do(ctx1, p.val.B().Publish().Channel(key).Message(string(msg)).Build()).Error(); err != nil {
		log.Println("Save", err)
	}
	ctx2 := context.Background()
	if err := p.val.Do(ctx2, p.val.B().Set().Key(key+"_").Value(string(msg)).Build()).Error(); err != nil {
		log.Println("Save", err)
	}
	wg.Done()
}

func (p *ValkeyClient) CallBack(msg valkey.PubSubMessage) {
	log.Println(msg.Channel)
}
