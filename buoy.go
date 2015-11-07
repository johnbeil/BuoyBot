// Copyright (c) 2015 John Beil.
// Use of this source code is governed by the MIT License.
// The MIT license that can be found in the LICENSE file.

// BuoyBot 0.1
// Tweets updates from the NBDC Station 46026 to @SFBuoy

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
)

type Configuration struct {
	APIKey    string `json:"APIKey"`
	APISecret string `json:APISecret"`
}

func main() {

	// import Twitter OAuth credentials from config.json
	var configuration Configuration
	file, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(file, &configuration)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(configuration)
}
