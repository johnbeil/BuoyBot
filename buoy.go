// Copyright (c) 2015 John Beil.
// Use of this source code is governed by the MIT License.
// The MIT license that can be found in the LICENSE file.

// BuoyBot 0.1
// Tweets updates from the NBDC Station 46026 to @SFBuoy

package main

import (
	"fmt"

	"github.com/ChimeraCoder/anaconda"
)

var api *anaconda.TwitterApi

func main() {
	fmt.Println(".....starting buoybot")

	api = anaconda.NewTwitterApi(TOKEN, TOKEN_SECRET)
	anaconda.SetConsumerKey(CONSUMER_KEY)
	anaconda.SetConsumerSecret(CONSUMER_SECRET)
}
