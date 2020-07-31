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
	"github.com/spf13/viper"
)

func InitConfiguration() error {
	viper.SetDefault("influx.enabled", false)
	viper.SetDefault("influx.host", "localhost")
	viper.SetDefault("influx.port", "8086")
	viper.SetDefault("influx.user", "")
	viper.SetDefault("influx.pass", "")
	viper.SetDefault("influx.secure", false)
	viper.SetDefault("influx.db", "infping")

	viper.SetDefault("fping.backoff", "1")
	viper.SetDefault("fping.retries", "0")
	viper.SetDefault("fping.tos", "0")
	viper.SetDefault("fping.summary", "10")
	viper.SetDefault("fping.period", "1000")
	viper.SetDefault("fping.dualstack", false)
	viper.SetDefault("fping.custom", map[string]string{})

	viper.SetDefault("hosts.hosts", []string{"localhost"})

	viper.SetConfigName("infping")
	viper.AddConfigPath("/etc/")
	viper.AddConfigPath("/usr/local/etc/")
	viper.AddConfigPath("/config/")
	viper.AddConfigPath(".")

	return viper.ReadInConfig()
}
