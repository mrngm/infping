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
	"time"
)

// runAndRead executes fping, parses the output into an FPingPoint, and then writes it to InfPingClient
func runAndRead(con InfPingClient, cfg FPingConfig, hosts []string) error {
	hostname := mustHostname()

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

	last_time := time.Now()
	buff := bufio.NewScanner(stderr)
	for buff.Scan() {
		text := buff.Text()
		log.Printf("raw: %q", text)

		rxHost := ""

		// Decide what type of line this is
		switch FPingLineType(text) {
		case FPingOutputTime:
			last_time, err = FPingExtractTimestamp(text)
			if err != nil {
				return err
			}
			continue
		case FPingOutputHostnameOnly:
			rxHost = FPingExtractHostnameOnly(text)
		case FPingOutputHostnameIP:
			rxHost = FPingExtractHostnameIP(text)
		default:
			return fmt.Errorf("unrecognized format: %q", text)
		}

		min, avg, max := FPingExtractMinAvgMax(text)
		lossp := FPingExtractLossPercentage(text)

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

// mustHostname returns the local hostname or throws an error
func mustHostname() string {
	name, err := os.Hostname()
	if err != nil {
		panic("unable to find hostname " + err.Error())
	}
	return name
}
