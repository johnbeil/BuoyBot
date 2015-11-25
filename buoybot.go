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
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ChimeraCoder/anaconda"
	_ "github.com/lib/pq"
)

// First two rows of text file, fixed width delimited, used for debugging
const header = "#YY  MM DD hh mm WDIR WSPD GST  WVHT   DPD   APD MWD   PRES  ATMP  WTMP  DEWP  VIS PTDY  TIDE\n#yr  mo dy hr mn degT m/s  m/s     m   sec   sec degT   hPa  degC  degC  degC  nmi  hPa    ft"

const noaaUrl = "http://www.ndbc.noaa.gov/data/realtime2/46026.txt"

// struct to store observation data
type Observation struct {
	Date                  time.Time
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
	UserName         string `json:"UserName"`
	ConsumerKey      string `json:"ConsumerKey"`
	ConsumerSecret   string `json:"ConsumerSecret"`
	Token            string `json:"Token"`
	TokenSecret      string `json:"TokenSecret"`
	DatabaseUrl      string `json:"DatabaseUrl"`
	DatabaseUser     string `json:"DatabaseUser"`
	DatabasePassword string `json:"DatabasePassword"`
	DatabaseName     string `json:"DatabaseName"`
}

var db *sql.DB

func main() {
	fmt.Println("Starting BuoyBot...")

	// load configuration
	config := Config{}
	loadConfig(&config)

	// Print runtime to console
	t := time.Now()
	fmt.Println(t.Format(time.UnixDate))

	// Load database
	dbinfo := fmt.Sprintf("user=%s password=%s host=%s dbname=%s sslmode=disable",
		config.DatabaseUser, config.DatabasePassword, config.DatabaseUrl, config.DatabaseName)
	var err error
	db, err = sql.Open("postgres", dbinfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal("Error: Could not establish connection with the database.", err)
	}

	var observation Observation
	observation = getObservation()
	saveObservation(observation)

	// tweet current conditions
	// tweetCurrent(config)

	fmt.Println("Exiting BuoyBot...")
}

func searchAtInterval(n time.Duration, query string, config Config) {
	var api *anaconda.TwitterApi
	api = anaconda.NewTwitterApi(config.Token, config.TokenSecret)
	anaconda.SetConsumerKey(config.ConsumerKey)
	anaconda.SetConsumerSecret(config.ConsumerSecret)
	for _ = range time.Tick(n * time.Second) {
		// t := time.Now()

		searchResult, _ := api.GetSearch(query, nil)
		for _, tweet := range searchResult.Statuses {
			fmt.Println(tweet.Text)
		}
	}
}

func getObservation() Observation {
	observationRaw := getDataFromURL(noaaUrl)
	observationData := parseData(observationRaw)
	return observationData
}

func saveObservation(o Observation) {
	_, err := db.Exec("INSERT INTO observations(observationtime, windspeed, winddirection, significantwaveheight, dominantwaveperiod, averageperiod, meanwavedirection, airtemperature, watertemperature) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9)", o.Date, o.WindSpeed, o.WindDirection, o.SignificantWaveHeight, o.DominantWavePeriod, o.AveragePeriod, o.MeanWaveDirection, o.AirTemperature, o.WaterTemperature)
	if err != nil {
		log.Fatal("Error saving observation:", err)
	}
}

func tweetCurrent(config Config) {
	var api *anaconda.TwitterApi
	api = anaconda.NewTwitterApi(config.Token, config.TokenSecret)
	anaconda.SetConsumerKey(config.ConsumerKey)
	anaconda.SetConsumerSecret(config.ConsumerSecret)
	observationRaw := getDataFromURL(noaaUrl)
	observationData := parseData(observationRaw)
	observationOutput := formatObservation(observationData)
	tweet, err := api.PostTweet(observationOutput, nil)
	if err != nil {
		fmt.Println("update error:", err)
	} else {
		fmt.Println(tweet.Text)
	}
}

func tweetAtInterval(n time.Duration, config Config) {
	var api *anaconda.TwitterApi
	api = anaconda.NewTwitterApi(config.Token, config.TokenSecret)
	anaconda.SetConsumerKey(config.ConsumerKey)
	anaconda.SetConsumerSecret(config.ConsumerSecret)
	for _ = range time.Tick(n * time.Second) {
		t := time.Now()
		fmt.Println("\n", t.Format(time.UnixDate))
		observationRaw := getDataFromURL(noaaUrl)
		observationData := parseData(observationRaw)
		observationOutput := formatObservation(observationData)

		tweet, err := api.PostTweet(observationOutput, nil)
		if err != nil {
			fmt.Println("update error:", err)
		} else {
			fmt.Println(tweet.Text)
		}
	}
}

func getDataFromURL(url string) (body []byte) {
	fmt.Println("Fetching data...")
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal("Error fetching data:", err)
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("ioutil error reading resp.Body:", err)
	}
	fmt.Println("Status:", resp.Status)
	return
}

// load config
func loadConfig(config *Config) {
	// load path to config from CONFIGPATH environment variable
	configpath := os.Getenv("CONFIGPATH")
	file, _ := os.Open(configpath)
	decoder := json.NewDecoder(file)
	err := decoder.Decode(&config)
	if err != nil {
		log.Fatal("Error loading config.json:", err)
	}
}

// process latest observation
func parseData(d []byte) Observation {
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

	// population observation
	var o Observation
	o.Date = t
	o.WindDirection = windcardinal
	o.WindSpeed = windspeedmph
	o.SignificantWaveHeight = waveheightfeet
	o.DominantWavePeriod, err = strconv.Atoi(datafield[9])
	if err != nil {
		fmt.Println("o.AveragePeriod:", err)
	}
	o.AveragePeriod, err = strconv.ParseFloat(datafield[10], 64)
	if err != nil {
		fmt.Println("o.AveragePeriod:", err)
	}
	o.MeanWaveDirection = wavecardinal
	o.AirTemperature = airtempF
	o.WaterTemperature = watertempF

	return o
}

// given Observation returns formatted text
func formatObservation(o Observation) string {
	output := fmt.Sprint(o.Date.Format(time.RFC822), "\nSwell: ", strconv.FormatFloat(float64(o.SignificantWaveHeight), 'f', 1, 64), "ft at ", o.DominantWavePeriod, " sec from ", o.MeanWaveDirection, "\nWind: ", strconv.FormatFloat(float64(o.WindSpeed), 'f', 0, 64), "mph from ", o.WindDirection, "\nTemp: Air ", o.AirTemperature, "F / Water: ", o.WaterTemperature, "F")
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
