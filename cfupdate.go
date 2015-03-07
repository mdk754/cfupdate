/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

// CloudFlare Updater
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/codegangsta/cli"
)

var Config struct {
	LogFile string
	Email   string
	Token   string
	Zone    string
	Records []struct {
		Hostname string
		Id       string
	}
	History struct {
		LastIP     string
		LastSet    int64
		NextVerify int64
	}
}

// getState will try to load the config file and return any error. It takes the
// file location as a string and a reference to the object to populate.
func getState(file string, o interface{}) error {
	config, err := ioutil.ReadFile(file)
	if err == nil {
		err = json.Unmarshal(config, &o)
	}
	return err
}

// setState tries to save the current config object to the file. It takes the
// file location as a string and a reference to the object it is saving.
func setState(file string, o interface{}) error {
	export, err := json.MarshalIndent(&o, "", "\t")
	if err == nil {
		err = ioutil.WriteFile(file, export, 0600)
	}
	return err
}

func main() {
	// Create command line app and populate meta.
	app := cli.NewApp()
	app.Name = "cfupdate"
	app.Version = "0.1.0"
	app.Usage = "Small utility to update CloudFlare when the public address " +
		"of the machine changes."
	app.Author = "Mdk754"
	app.HideVersion = true

	// Register slice of global flags with package cli.
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Usage: "specify the location of the config file",
		},
	}

	app.Action = run

	// This is the program entry point. Feeds arguemnts into cli package.
	app.Run(os.Args)
}

func run(c *cli.Context) {
	config := c.String("config")

	if config == "" {
		fmt.Println("error: No config file specified. Use --config or -c.")
		os.Exit(1)
	}

	if err := getState(config, &Config); err != nil {
		fmt.Println("error: ", err.Error())
		os.Exit(1)
	}

	publicIP, err := getPublicIP()
	if err != nil {
		fmt.Println("error: ", err.Error())
		os.Exit(1)
	}

	if publicIP != Config.History.LastIP {
		// Needs update, set the IP.
		if err := setPublicIP(publicIP); err != nil {
			fmt.Println("error: ", err.Error())
			os.Exit(1)
		}
		Config.History.LastIP = publicIP
		Config.History.LastSet = time.Now().Unix()
		Config.History.NextVerify = time.Now().Add(14 * time.Minute).Unix()
		if err := setState(config, &Config); err != nil {
			fmt.Println("error: ", err.Error())
			os.Exit(1)
		}
	}
}

// getPublicIP will lookup the current public IPv4 address of the machine. This
// means that if the machine is behind NAT, the internet routable IP will be
// returned.
func getPublicIP() (string, error) {
	resp, err := http.Get("http://myip.dnsomatic.com/")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	ip, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(ip), nil
}

// setPublicIP
func setPublicIP(ip string) error {
	api := "https://www.cloudflare.com"
	resource := "/api_json.html"

	for _, record := range Config.Records {
		data := url.Values{}
		data.Set("a", "rec_edit")
		data.Set("tkn", Config.Token)
		data.Set("email", Config.Email)
		data.Set("z", Config.Zone)
		data.Set("id", record.Id)
		data.Set("name", record.Hostname)
		data.Set("type", "A")
		data.Set("ttl", "1")
		data.Set("content", ip)

		u, _ := url.ParseRequestURI(api)
		u.Path = resource
		urlStr := fmt.Sprintf("%v", u)

		resp, err := http.PostForm(urlStr, data)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		fmt.Println("Body: ", string(body))
		fmt.Println("Status Code: ", resp.Status, "\n")
	}

	return nil
}
