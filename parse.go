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
	"bufio"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var hostname = mustHostname()
var last_time = time.Now()

// runAndRead executes fping, parses the output into an FPingPoint, and then writes it to InfPingClient
func runAndRead(hosts []string, con InfPingClient, cfg FPingConfig) error {
	args := make([]string, 0, len(cfg)+len(hosts))
	for k, v := range cfg {
		args = append(args, k, v)
	}
	for _, v := range hosts {
		args = append(args, v)
	}
	log.Printf("args: %q", args)
	cmd, err := exec.LookPath("fping")
	if err != nil {
		return err
	}
	runner := exec.Command(cmd, args...)
	stderr, err := runner.StderrPipe()
	if err != nil {
		return err
	}
	runner.Start()

	buff := bufio.NewScanner(stderr)
	for buff.Scan() {
		text := buff.Text()
		log.Printf("raw: %q", text)
		fields := strings.Fields(text)

		if len(fields) == 1 {
			tm := strings.TrimLeft(fields[0], "[")
			tm = strings.TrimRight(tm, "]")
			parsed, err := time.Parse("15:04:05", tm)
			if err != nil {
				log.Printf("Failed to parse time %s: %s", tm, err)
			} else {
				now := time.Now()
				last_time = time.Date(now.Year(), now.Month(), now.Day(), parsed.Hour(), parsed.Minute(), parsed.Second(), 0, time.Local)
			}
		} else {
			host := fields[0]
			data := fields[4]
			dataSplitted := strings.Split(data, "/")
			// Remove ,
			dataSplitted[2] = strings.TrimRight(dataSplitted[2], "%,")
			lossp := mustInt(dataSplitted[2])
			min, max, avg := 0.0, 0.0, 0.0
			// Ping times
			if len(fields) > 5 {
				times := fields[7]
				td := strings.Split(times, "/")
				min, avg, max = mustFloat(td[0]), mustFloat(td[1]), mustFloat(td[2])
			}

			pt := FPingPoint{
				TxHost:      hostname,
				RxHost:      host,
				Min:         min,
				Max:         max,
				Avg:         avg,
				LossPercent: lossp,
				Time:        last_time,
			}

			if err := con.Write(pt); err != nil {
				log.Printf("Error writing data point: %s", err)
			}
		}
	}
	return nil
}

// mustInt ensures the string contains an integer, returning 0 if not
func mustInt(data string) int {
	in, err := strconv.Atoi(data)
	if err != nil {
		return 0
	}
	return in
}

// mustFloat ensures the string contains a float, returning 0.0 if not
func mustFloat(data string) float64 {
	flt, err := strconv.ParseFloat(data, 64)
	if err != nil {
		return 0.0
	}
	return flt
}

// mustHostname returns the local hostname or throws an error
func mustHostname() string {
	name, err := os.Hostname()
	if err != nil {
		panic("unable to find hostname " + err.Error())
	}
	return name
}
