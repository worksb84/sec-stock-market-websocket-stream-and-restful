package models

import (
	"encoding/json"
	"strconv"
	"time"
)

type Ticker struct {
	Ticker          string `json:"ticker"`
	Name            string `json:"name"`
	Market          string `json:"market"`
	Locale          string `json:"locale"`
	PrimaryExchange string `json:"primary_exchange"`
	Type            string `json:"type"`
	Active          string `json:"active"`
	CurrencyName    string `json:"currency_name"`
	Cik             string `json:"cik"`
	CompositeFIGI   string `json:"composite_figi"`
	Share_classFIGI string `json:"share_class_figi"`
	LastUpdatedUTC  string `json:"last_updated_utc"`
}

type Tickers struct {
	Results []Ticker `json:"results"`
	NextUrl string   `json:"next_url"`
}

type Millis time.Time

func (m *Millis) UnmarshalJSON(data []byte) error {
	d, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return err
	}
	*m = Millis(time.UnixMilli(d))
	return nil
}

func (m Millis) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(m).UnixMilli())
}

type Nanos time.Time

func (n *Nanos) UnmarshalJSON(data []byte) error {
	d, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return err
	}
	timeNano := time.Unix(d/1_000_000_000, d%1_000_000_000)
	*n = Nanos(timeNano)
	return nil
}

func (n Nanos) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(n).UnixNano())
}

type Date time.Time

func (d *Date) UnmarshalJSON(data []byte) error {
	unquoteData, err := strconv.Unquote(string(data))
	if err != nil {
		return err
	}
	t, err := time.Parse("2006-01-02", unquoteData)
	if err != nil {
		return err
	}
	*d = Date(t)
	return nil
}

func (d *Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(*d).Format("2006-01-02"))
}

type Time time.Time

func (t *Time) UnmarshalJSON(data []byte) error {
	unquoteData, err := strconv.Unquote(string(data))
	if err != nil {
		return err
	}

	if parsedTime, err := time.Parse("2006-01-02T15:04:05.000-0700", unquoteData); err == nil {
		*t = Time(parsedTime)
		return nil
	}

	if parsedTime, err := time.Parse("2006-01-02T15:04:05-07:00", unquoteData); err == nil {
		*t = Time(parsedTime)
		return nil
	}

	if parsedTime, err := time.Parse("2006-01-02T15:04:05.000Z", unquoteData); err == nil {
		*t = Time(parsedTime)
		return nil
	}

	if parsedTime, err := time.Parse("2006-01-02T15:04:05Z", unquoteData); err != nil {
		return err
	} else {
		*t = Time(parsedTime)
	}

	return nil
}

func (t *Time) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(*t).Format("2006-01-02T15:04:05.000Z"))
}

type DaySnapshot struct {
	Close                 float64 `json:"c,omitempty"`
	High                  float64 `json:"h,omitempty"`
	Low                   float64 `json:"l,omitempty"`
	Open                  float64 `json:"o,omitempty"`
	Volume                float64 `json:"v,omitempty"`
	VolumeWeightedAverage float64 `json:"vw,omitempty"`
	OTC                   bool    `json:"otc,omitempty"`
}

type LastQuoteSnapshot struct {
	AskPrice  float64 `json:"P,omitempty"`
	BidPrice  float64 `json:"p,omitempty"`
	AskSize   float64 `json:"S,omitempty"`
	BidSize   float64 `json:"s,omitempty"`
	Timestamp Nanos   `json:"t,omitempty"`
}

type LastTradeSnapshot struct {
	Conditions []int   `json:"c,omitempty"`
	TradeID    string  `json:"i,omitempty"`
	Price      float64 `json:"p,omitempty"`
	Size       float64 `json:"s,omitempty"`
	Timestamp  Nanos   `json:"t,omitempty"`
	ExchangeID int     `json:"x,omitempty"`
}

type MinuteSnapshot struct {
	AccumulatedVolume     float64 `json:"av,omitempty"`
	Close                 float64 `json:"c,omitempty"`
	High                  float64 `json:"h,omitempty"`
	Low                   float64 `json:"l,omitempty"`
	Open                  float64 `json:"o,omitempty"`
	Volume                float64 `json:"v,omitempty"`
	VolumeWeightedAverage float64 `json:"vw,omitempty"`
	NumberOfTransactions  float64 `json:"n,omitempty"`
	Timestamp             Millis  `json:"t,omitempty"`
	OTC                   bool    `json:"otc,omitempty"`
}

type TickerSnapshot struct {
	Day              DaySnapshot       `json:"day,omitempty"`
	LastQuote        LastQuoteSnapshot `json:"lastQuote,omitempty"`
	LastTrade        LastTradeSnapshot `json:"lastTrade,omitempty"`
	Minute           MinuteSnapshot    `json:"min,omitempty"`
	PrevDay          DaySnapshot       `json:"prevDay,omitempty"`
	Ticker           string            `json:"ticker,omitempty"`
	TodaysChange     float64           `json:"todaysChange,omitempty"`
	TodaysChangePerc float64           `json:"todaysChangePerc,omitempty"`
	Updated          Nanos             `json:"updated,omitempty"`
	FairMarketValue  float64           `json:"fmv,omitempty"`
}

type TickersSnapshot struct {
	Tickers []TickerSnapshot `json:"tickers,omitempty"`
}

type MarketHoliday struct {
	Exchange string `json:"exchange,omitempty"`
	Name     string `json:"name,omitempty"`
	Date     Date   `json:"date,omitempty"`
	Status   string `json:"status,omitempty"`
	Open     Time   `json:"open,omitempty"`
	Close    Time   `json:"close,omitempty"`
}
