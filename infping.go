/*
The MIT License (MIT)

Copyright (c) 2016 Tor Hveem              https://github.com/torhve/infping
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
	"log"
	"strings"

	"github.com/spf13/viper"
)

// InfPingClient defines how results can be obtained from this program
type InfPingClient interface {
	Write(point FPingPoint) error
}

func main() {
	if err := InitConfiguration(); err != nil {
		log.Fatalf("Unable to read config file: %v", err)
	}

	var client InfPingClient
	if viper.GetBool("influx.enabled") {
		log.Print("InfluxDB enabled, setting up client")
		client = SetupInfluxDBClient()
	} else {
		log.Print("Setting up mock client")
		client = SetupMockClient()
	}

	log.Print("Setting up fping")
	fpingCfg := SetupFPing()

	hosts := viper.GetStringSlice("hosts.hosts")

	log.Printf("Launching fping with hosts: %s", strings.Join(hosts, ", "))
	log.Fatalf("%v", runAndRead(client, fpingCfg, hosts))
}
