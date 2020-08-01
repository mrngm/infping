/*
The MIT License (MIT)

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
	"regexp"
	"strconv"
	"time"

	"github.com/spf13/viper"
)

var (
	FPingTimeRegexp         = regexp.MustCompile(`^\[(?P<time>\d\d:\d\d:\d\d)\]$`)
	FPingHostnameOnlyRegexp = regexp.MustCompile(`^(?P<dns>[^ ]+)\s+:`)                     // "localhost : xmt/rcv/%loss = 10/10/0%, min/avg/max = 0.02/0.06/0.08"
	FPingHostnameIPRegexp   = regexp.MustCompile(`^(?P<dns>[^ ]+)\s+\((?P<ip>[^)]+)\)\s+:`) // "localhost (::1)       : xmt/rcv/%loss = 10/10/0%, min/avg/max = 0.04/0.06/0.07"
	FPingXmtRcvLossRegexp   = regexp.MustCompile(`xmt/rcv/%loss\s+=\s+(?P<xmt>\d+)/(?P<rcv>\d+)/(?P<loss>\d+)%,`)
	FPingMinAvgMaxRegexp    = regexp.MustCompile(`min/avg/max\s+=\s+(?P<min>\d+\.\d+)/(?P<avg>\d+\.\d+)/(?P<max>\d+\.\d+)$`)
)

type FPingOutputLine int

const (
	FPingOutputUnknown FPingOutputLine = iota
	FPingOutputTime
	FPingOutputHostnameOnly
	FPingOutputHostnameIP
)

func FPingLineType(s string) FPingOutputLine {
	switch {
	case FPingTimeRegexp.MatchString(s):
		return FPingOutputTime
	case FPingHostnameIPRegexp.MatchString(s):
		return FPingOutputHostnameIP
	case FPingHostnameOnlyRegexp.MatchString(s):
		return FPingOutputHostnameOnly
	}
	return FPingOutputUnknown
}

func FPingExtractTimestamp(s string) (time.Time, error) {
	timestamp := FPingTimeRegexp.FindStringSubmatch(s)
	if len(timestamp) != 2 {
		return time.Time{}, fmt.Errorf("cannot find time in line: %q", s)
	}

	ts, err := time.Parse("15:04:05", timestamp[1])
	if err != nil {
		return time.Time{}, fmt.Errorf("cannot parse time in line: %q, %v", s, err)
	}

	now := time.Now()
	t := time.Date(now.Year(), now.Month(), now.Day(), ts.Hour(), ts.Minute(), ts.Second(), 0, time.Local)
	log.Printf("parsed time: %v", t)
	return t, nil
}

func FPingExtractHostnameIP(s string) string {
	matches := FPingHostnameIPRegexp.FindStringSubmatch(s)
	if len(matches) != 3 {
		return ""
	}

	host := fmt.Sprintf("%s (%s)", matches[1], matches[2])
	log.Printf("parsed hostname+ip: %q", host)
	return host
}

func FPingExtractHostnameOnly(s string) string {
	host := FPingHostnameOnlyRegexp.FindStringSubmatch(s)
	if len(host) != 2 {
		return ""
	}

	log.Printf("parsed hostname: %q", host[1])
	return host[1]
}

func FPingExtractMinAvgMax(s string) (min, avg, max float64) {
	matches := FPingMinAvgMaxRegexp.FindStringSubmatch(s)
	if len(matches) != 4 {
		return 0.0, 0.0, 0.0
	}

	// The regexp asserts that the float has the correct format, so we can ignore the error returned by strconv.ParseFloat
	min, _ = strconv.ParseFloat(matches[1], 64)
	avg, _ = strconv.ParseFloat(matches[2], 64)
	max, _ = strconv.ParseFloat(matches[3], 64)
	log.Printf("parsed min/avg/max: %.2f/%.2f/%.2f", min, avg, max)
	return min, avg, max
}

func FPingExtractLossPercentage(s string) int {
	matches := FPingXmtRcvLossRegexp.FindStringSubmatch(s)
	if len(matches) != 4 {
		return 0
	}

	// The regexp asserts we matched a decimal number, so we can ignore the error returned by strconv.Atoi
	lossp, _ := strconv.Atoi(matches[3])
	log.Printf("parsed loss percentage: %d", lossp)
	return lossp
}

type FPingConfig map[string]string

// FPingPoint represents the fping results for a single host
type FPingPoint struct {
	Time        time.Time
	RxHost      string
	TxHost      string
	LossPercent int
	Min         float64
	Avg         float64
	Max         float64
}

func (fp FPingPoint) String() string {
	return fmt.Sprintf("[%v] (%s) %s: %.2f/%.2f/%.2f (%d%%)\n", fp.Time.Format("2006-01-02 15:04:05"), fp.TxHost, fp.RxHost, fp.Min, fp.Avg, fp.Max, fp.LossPercent)
}

func SetupFPing() FPingConfig {
	cfg := FPingConfig{
		"-B": viper.GetString("fping.backoff"),
		"-r": viper.GetString("fping.retries"),
		"-O": viper.GetString("fping.tos"),
		"-Q": viper.GetString("fping.summary"),
		"-p": viper.GetString("fping.period"),
		"-l": "", // loop
		"-D": "", // timestamp
		"-n": "", // show DNS names
		"-A": "", // display address
	}

	if viper.GetBool("fping.dualstack") {
		cfg["-m"] = "" // send to all addresses
	}

	fpingCustom := viper.GetStringMapString("fping.custom")
	for k, v := range fpingCustom {
		cfg[k] = v
	}
	return cfg
}
