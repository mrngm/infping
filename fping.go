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
	"time"

	"github.com/spf13/viper"
)

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
	return fmt.Sprintf("[%v] (%s) %s: %.2f/%.2f/%.2f (%d%%)\n", fp.Time.Format("2006-01-02 15:04:05"), fp.RxHost, fp.TxHost, fp.Min, fp.Avg, fp.Max, fp.LossPercent)
}

func SetupFPing() FPingConfig {
	cfg := FPingConfig{
		"-B": viper.GetString("fping.backoff"),
		"-r": viper.GetString("fping.retries"),
		"-O": viper.GetString("fping.tos"),
		"-Q": viper.GetString("fping.summary"),
		"-p": viper.GetString("fping.period"),
		"-l": "",
		"-D": "",
	}

	if viper.GetBool("fping.dualstack") {
		cfg["-m"] = "" // send to all addresses
		cfg["-n"] = "" // show DNS names
		cfg["-A"] = "" // display address
	}

	fpingCustom := viper.GetStringMapString("fping.custom")
	for k, v := range fpingCustom {
		cfg[k] = v
	}
	return cfg
}
