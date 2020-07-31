/*
The MIT License (MIT)

Copyright (c) 2017 Nicholas Van Wiggeren  https://github.com/nickvanw/infping
Copyright (c) 2018 Michael Newton         https://github.com/miken32/infping
Copyright (c) 2020 Gerdriaan Mulder       https://github.com/mrngm/infping

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package main

import (
	"fmt"
	"log"
	"net/url"
	"time"

	influxdbClient "github.com/influxdata/influxdb1-client/v2"
	"github.com/spf13/viper"
)

func SetupInfluxDBClient() InfPingClient {
	influxScheme := "https"
	if !viper.GetBool("influx.secure") {
		influxScheme = "http"
	}
	influxHost := viper.GetString("influx.host")
	influxPort := viper.GetString("influx.port")
	influxUser := viper.GetString("influx.user")
	influxPass := viper.GetString("influx.pass")
	influxDB := viper.GetString("influx.db")
	influxRetPolicy := viper.GetString("influx.policy")

	u, err := url.Parse(fmt.Sprintf("%s://%s:%s", influxScheme, influxHost, influxPort))
	if err != nil {
		log.Fatalf("Unable to build valid Influx URL: %v", err)
	}

	conf := influxdbClient.HTTPConfig{
		Addr:     u.String(),
		Username: influxUser,
		Password: influxPass,
	}

	rawClient, err := influxdbClient.NewHTTPClient(conf)
	if err != nil {
		log.Fatal("Failed to create Influx client", err)
	}

	influxClient := NewInfluxClient(rawClient, influxDB, influxRetPolicy)

	dur, version, err := influxClient.Ping()
	if err != nil {
		log.Fatal("Unable to ping InfluxDB", err)
	}
	log.Printf("Pinged InfluxDB (version %s) in %v", version, dur)

	q := influxdbClient.Query{
		Command: "SHOW DATABASES",
	}
	databases, err := influxClient.Query(q)
	if err != nil {
		log.Fatal("Unable to list databases", err)
	}
	if len(databases.Results) != 1 {
		log.Fatalf("Expected 1 result in response, got %d", len(databases.Results))
	}
	if len(databases.Results[0].Series) != 1 {
		log.Fatalf("Expected 1 series in result, got %d", len(databases.Results[0].Series))
	}
	found := false
	for i := 0; i < len(databases.Results[0].Series[0].Values); i++ {
		if databases.Results[0].Series[0].Values[i][0] == influxDB {
			found = true
		}
	}
	if !found {
		q = influxdbClient.Query{
			Command: fmt.Sprintf("CREATE DATABASE %s", influxDB),
		}
		_, err := influxClient.Query(q)
		if err != nil {
			log.Fatalf("Failed to create database %s %v", influxDB, err)
		}
		log.Printf("Created new database %s", influxDB)
	}

	return influxClient
}

// NewInfluxClient creates a concrete InfluxDB Writer
func NewInfluxClient(client influxdbClient.Client, db, retPolicy string) *InfluxClient {
	return &InfluxClient{
		influx:    client,
		db:        db,
		retPolicy: retPolicy,
	}
}

// InfluxClient implements the Client interface to provide a metrics client
// backed by InfluxDB
type InfluxClient struct {
	influx    influxdbClient.Client
	db        string
	retPolicy string
}

// Ping calls Ping on the underlying influx client
func (i *InfluxClient) Ping() (time.Duration, string, error) {
	return i.influx.Ping(time.Second)
}

// Query calls Query on the underlying influx client
func (i *InfluxClient) Query(q influxdbClient.Query) (*influxdbClient.Response, error) {
	return i.influx.Query(q)
}

// Write a single FPingPoint to influx
func (i *InfluxClient) Write(point FPingPoint) error {
	var fields map[string]interface{}
	if point.Min != 0 && point.Avg != 0 && point.Max != 0 {
		fields = map[string]interface{}{
			"loss": point.LossPercent,
			"min":  point.Min,
			"avg":  point.Avg,
			"max":  point.Max,
		}
	} else {
		fields = map[string]interface{}{
			"loss": point.LossPercent,
		}
	}
	pt, err := influxdbClient.NewPoint(
		"ping",
		map[string]string{
			"rx_host": point.RxHost,
			"tx_host": point.TxHost,
		},
		fields,
		point.Time)

	if err != nil {
		return err
	}

	batchConfig := influxdbClient.BatchPointsConfig{Database: i.db, Precision: "s"}
	bp, err := influxdbClient.NewBatchPoints(batchConfig)
	if err != nil {
		return err
	}
	bp.AddPoint(pt)
	if i.retPolicy != "" {
		bp.SetRetentionPolicy(i.retPolicy)
	}

	return i.influx.Write(bp)
}
