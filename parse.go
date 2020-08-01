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
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

var (
	hostname  = mustHostname()
	last_time = time.Now()

	fpingTimeRegexp = regexp.MustCompile(`^\[(?P<time>\d\d:\d\d:\d\d)\]$`)
	// "localhost : xmt/rcv/%loss = 10/10/0%, min/avg/max = 0.02/0.06/0.08"
	fpingSingleRegexp = regexp.MustCompile(`^(?P<dns>[^ ]+)\s+:`)
	// "localhost (::1)       : xmt/rcv/%loss = 10/10/0%, min/avg/max = 0.04/0.06/0.07"
	fpingDualstackRegexp  = regexp.MustCompile(`^(?P<dns>[^ ]+)\s+\((?P<ip>[^)]+)\)\s+:`)
	fpingXmtRcvLossRegexp = regexp.MustCompile(`xmt/rcv/%loss\s+=\s+(?P<xmt>\d+)/(?P<rcv>\d+)/(?P<loss>\d+)%,`)
	fpingMinAvgMaxRegexp  = regexp.MustCompile(`min/avg/max\s+=\s+(?P<min>\d+\.\d+)/(?P<avg>\d+\.\d+)/(?P<max>\d+\.\d+)$`)
)

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

		min, max, avg := 0.0, 0.0, 0.0
		lossp := 0
		rxHost := ""

		// Decide what type of line this is
		switch {
		case fpingTimeRegexp.MatchString(text):
			t, err := time.Parse("15:04:05", fpingTimeRegexp.FindStringSubmatch(text)[1])
			if err != nil {
				return fmt.Errorf("cannot parse time in line: %q, %v", text, err)
			}
			now := time.Now()
			last_time = time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), t.Second(), 0, time.Local)
			log.Printf("parsed time: %v", last_time)
			continue
		case fpingDualstackRegexp.MatchString(text):
			matches := fpingDualstackRegexp.FindStringSubmatch(text)
			rxHost = fmt.Sprintf("%s (%s)", matches[1], matches[2])
			log.Printf("parsed rxHost: %q", rxHost)
		case fpingSingleRegexp.MatchString(text):
			rxHost = fpingSingleRegexp.FindStringSubmatch(text)[1]
			log.Printf("parsed rxHost: %q", rxHost)
		default:
			return fmt.Errorf("unrecognized format: %q", text)
		}

		lossp = mustInt(fpingXmtRcvLossRegexp.ReplaceAllString(text, "${loss}"))
		minAvgMax := fpingMinAvgMaxRegexp.FindStringSubmatch(text)
		min = mustFloat(minAvgMax[1])
		avg = mustFloat(minAvgMax[2])
		max = mustFloat(minAvgMax[3])

		pt := FPingPoint{
			TxHost:      hostname,
			RxHost:      rxHost,
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
