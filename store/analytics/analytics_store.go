package analytics

import (
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	influxAPI "github.com/influxdata/influxdb-client-go/v2/api"
	influxAPIWrite "github.com/influxdata/influxdb-client-go/v2/api/write"
)

const URL_CREATION_MEASUREMENT = "EventURLCreation"
const URL_REDIRECT_MEASUREMENT = "EventURLRedirect"

type tags map[string]string
type fields map[string]any

type URLCreationEvent struct {
	ServiceName string
	URL         string
	APIVer      int32
	Success     bool
	Timestamp   time.Time
}

type URLRedirectEvent struct {
	ServiceName string
	ShortURL    string
	LongURL     string
	APIVer      int32
	Success     bool
	Timestamp   time.Time
}

type AnalyticsStore interface {
	WriteURLCreationEvent(*URLCreationEvent)
	WriteURLRedirectEvent(*URLRedirectEvent)
	Errors() <-chan error
	Flush()
	Close()
}

type InfluxDBAnalyticsStore struct {
	client   influxdb2.Client
	writeAPI influxAPI.WriteAPI
}

func NewInfluxDBAnalyticsStore(client influxdb2.Client, org string, bucket string) *InfluxDBAnalyticsStore {
	api := client.WriteAPI(org, bucket)
	return &InfluxDBAnalyticsStore{
		client:   client,
		writeAPI: api,
	}
}

func (ias *InfluxDBAnalyticsStore) WriteURLRedirectEvent(event *URLRedirectEvent) {
	t := tags{
		"service": event.ServiceName,
	}
	f := fields{
		"short_url": event.ShortURL,
		"long_url":  event.LongURL,
		"api_ver":   event.APIVer,
		"success":   event.Success,
	}
	point := influxAPIWrite.NewPoint(URL_REDIRECT_MEASUREMENT, t, f, event.Timestamp)
	ias.writeAPI.WritePoint(point)
}

func (ias *InfluxDBAnalyticsStore) WriteURLCreationEvent(event *URLCreationEvent) {
	t := tags{
		"service": event.ServiceName,
	}
	f := fields{
		"url":     event.URL,
		"api_ver": event.APIVer,
		"success": event.Success,
	}
	point := influxAPIWrite.NewPoint(URL_CREATION_MEASUREMENT, t, f, event.Timestamp)
	ias.writeAPI.WritePoint(point)
}

func (ias *InfluxDBAnalyticsStore) Errors() <-chan error {
	return ias.writeAPI.Errors()
}

func (ias *InfluxDBAnalyticsStore) Flush() {
	ias.writeAPI.Flush()
}

func (ias *InfluxDBAnalyticsStore) Close() {
	ias.Flush()
	ias.client.Close()
}
