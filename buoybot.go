// Copyright (c) 2015 John Beil.
// Use of this source code is governed by the MIT License.
// The MIT license that can be found in the LICENSE file.

// BuoyBot 1.1
// Obtains latest observation for NBDC Station 46026
// Tweets observation from @SFBuoy
// Each line contains 19 data points
// Headers are in the first two lines
// Latest data is in the third line
// Other lines are not needed

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ChimeraCoder/anaconda"
)

// First two rows of text file, fixed width delimited, used for debugging
const header = "#YY  MM DD hh mm WDIR WSPD GST  WVHT   DPD   APD MWD   PRES  ATMP  WTMP  DEWP  VIS PTDY  TIDE\n#yr  mo dy hr mn degT m/s  m/s     m   sec   sec degT   hPa  degC  degC  degC  nmi  hPa    ft"

// struct to store observation data
type BuoyData struct {
	Date                  string
	WindDirection         string
	WindSpeed             float64
	SignificantWaveHeight float64
	DominantWavePeriod    int
	AveragePeriod         float64
	MeanWaveDirection     string
	AirTemperature        float64
	WaterTemperature      float64
}

// struct to store credentials
type Config struct {
	UserName       string `json:UserName`
	ConsumerKey    string `json:ConsumerKey`
	ConsumerSecret string `json:ConsumerSecret`
	Token          string `json:Token`
	TokenSecret    string `json:TokenSecret`
}

func main() {
	fmt.Println("Starting BuoyBot...")

	config := Config{}
	loadConfig(&config)

	// var data string
	var noaaUrl string = "http://www.ndbc.noaa.gov/data/realtime2/46026.txt"

	var api *anaconda.TwitterApi
	api = anaconda.NewTwitterApi(config.Token, config.TokenSecret)
	anaconda.SetConsumerKey(config.ConsumerKey)
	anaconda.SetConsumerSecret(config.ConsumerSecret)

	c := time.Tick(10800 * time.Second)
	for _ = range c {

		body := getDataFromURL(noaaUrl)
		output := parseData(body)

		//  create go routine and post update every 6 hours
		//  60 sec * 60 min * 6 hours = 21600 seconds
		// go updateAtInterval(21600, url, config)
		// select {} // this will cause the program to run forever

		// print raw data for most recent observation
		// fmt.Println(header)
		// fmt.Println(data)

		// post tweet
		tweet, err := api.PostTweet(output, nil)
		if err != nil {
			fmt.Println("update error:", err)
		}
		fmt.Println("Tweet contents:")
		fmt.Println(tweet.Text)

		fmt.Println("\n...hibernating")
	}

}

func getDataFromURL(url string) (body []byte) {
	fmt.Println("...Fetching data...")
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error fetching data:", err)
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("ioutil error:", err)
	}
	fmt.Println("Status:", resp.Status)
	return
}

// load config
func loadConfig(config *Config) {
	// Get config
	file, _ := os.Open("config.json")
	decoder := json.NewDecoder(file)
	err := decoder.Decode(&config)
	if err != nil {
		fmt.Println("error loading config.json:", err)
	}
}

// process latest observation
func parseData(d []byte) string {
	var data string = string(d[188:281])
	// convert most recent observation into array of strings
	datafield := strings.Fields(data)

	// convert wave height from meters to feet
	waveheightmeters, _ := strconv.ParseFloat(datafield[8], 64)
	waveheightfeet := waveheightmeters * 3.28084

	// convert wave direction from degrees to cardinal
	wavedegrees, _ := strconv.ParseInt(datafield[11], 0, 64)
	wavecardinal := direction(wavedegrees)

	// convert wind speed from m/s to mph
	windspeedms, _ := strconv.ParseFloat((datafield[6]), 64)
	windspeedmph := windspeedms / 0.44704

	// convert wind direction from degrees to cardinal
	winddegrees, _ := strconv.ParseInt(datafield[5], 0, 64)
	windcardinal := direction(winddegrees)

	// convert air temp from C to F
	airtempC, _ := strconv.ParseFloat(datafield[13], 64)
	airtempF := airtempC*9/5 + 32
	airtempF = RoundPlus(airtempF, 1)

	// convert water temp from C to F
	watertempC, _ := strconv.ParseFloat(datafield[14], 64)
	watertempF := watertempC*9/5 + 32
	watertempF = RoundPlus(watertempF, 1)

	// process date/time and convert to PST
	rawtime := strings.Join(datafield[0:5], " ")
	t, err := time.Parse("2006 01 02 15 04", rawtime)
	if err != nil {
		fmt.Println(err)
	}
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		fmt.Println(err)
	}
	t = t.In(loc)
	// fmt.Println(t.Format(time.RFC822))

	// var buoydata BuoyData

	// concatenate data to output and print to console
	output := fmt.Sprint("\nSF Buoy at ", t.Format(time.RFC822), "\nSwell: ", strconv.FormatFloat(float64(waveheightfeet), 'f', 1, 64), "ft at ", datafield[9], " sec from ", wavecardinal, "\nWind: ", strconv.FormatFloat(float64(windspeedmph), 'f', 0, 64), "mph from ", windcardinal, "\nTemp: Air ", airtempF, "F / Water: ", watertempF, "F")
	// fmt.Println(output)
	return output
}

// given degrees returns cardinal direction
func direction(deg int64) string {
	switch {
	case deg < 0:
		return "ERROR - DEGREE LESS THAN ZERO"
	case deg <= 11:
		return "N"
	case deg <= 34:
		return "NNE"
	case deg <= 56:
		return "NE"
	case deg <= 79:
		return "ENE"
	case deg <= 101:
		return "E"
	case deg <= 124:
		return "ESE"
	case deg <= 146:
		return "SE"
	case deg <= 169:
		return "SSE"
	case deg <= 191:
		return "S"
	case deg <= 214:
		return "SSW"
	case deg <= 236:
		return "SW"
	case deg <= 259:
		return "WSW"
	case deg <= 281:
		return "W"
	case deg <= 304:
		return "WNW"
	case deg <= 326:
		return "NW"
	case deg <= 349:
		return "NNW"
	case deg <= 360:
		return "N"
	default:
		return "ERROR - DEGREE GREATER THAN 360"
	}
}

// round input to nearest integer
func Round(f float64) float64 {
	return math.Floor(f + .5)
}

// round input to defined number of decimals
func RoundPlus(f float64, places int) float64 {
	shift := math.Pow(10, float64(places))
	return Round(f*shift) / shift
}
