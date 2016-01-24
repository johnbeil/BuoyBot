// Copyright (c) 2016 John Beil.
// Use of this source code is governed by the MIT License.
// The MIT license can be found in the LICENSE file.

// BuoyBot 1.1
// Obtains latest observation for NBDC Station 46026
// Saves observation to database
// Tweets observation from @SFBuoy
// See README.md for setup information

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

// URL for SF Buoy Observations
const noaaURL = "http://www.ndbc.noaa.gov/data/realtime2/46026.txt"

// Observation struct stores buoy observation data
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

// Config struct stores Twitter and Database credentials
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

// Tide stores a tide prediction from the database
type Tide struct {
	Date         string
	Day          string
	Time         string
	PredictionFt float64
	PredictionCm int64
	HighLow      string
}

// Variable for database
var db *sql.DB

func main() {
	fmt.Println("Starting BuoyBot...")

	// Load configuration
	config := Config{}
	loadConfig(&config)

	// Load database
	dbinfo := fmt.Sprintf("user=%s password=%s host=%s dbname=%s sslmode=disable",
		config.DatabaseUser, config.DatabasePassword, config.DatabaseUrl, config.DatabaseName)
	var err error
	db, err = sql.Open("postgres", dbinfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Check database connection
	err = db.Ping()
	if err != nil {
		log.Fatal("Error: Could not establish connection with the database.", err)
	}

	// Get latest observation and store in struct
	var observation Observation
	observation = getObservation()

	// Save latest observation in database
	saveObservation(observation)

	tide := getTide()
	processTide(tide)

	// Tweet observation at 0000, 0600, 0900, 1200, 1500, 1800 PST
	t := time.Now()
	if t.Hour() == 0 || t.Hour() == 6 || t.Hour() == 9 || t.Hour() == 12 || t.Hour() == 15 || t.Hour() == 18 {
		tweetCurrent(config, observation)
	} else {
		fmt.Println("Not at update interval - not tweeting.")
		fmt.Println("Observation is:\n", observation)
	}

	// Shutdown BuoyBot
	fmt.Println("Exiting BuoyBot...")
}

// Fetches and parses latest NBDC observation and returns data in Observation struct
func getObservation() Observation {
	observationRaw := getDataFromURL(noaaURL)
	observationData := parseData(observationRaw)
	return observationData
}

// Given Observation struct, saves most recent observation in database
func saveObservation(o Observation) {
	_, err := db.Exec("INSERT INTO observations(observationtime, windspeed, winddirection, significantwaveheight, dominantwaveperiod, averageperiod, meanwavedirection, airtemperature, watertemperature) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9)", o.Date, o.WindSpeed, o.WindDirection, o.SignificantWaveHeight, o.DominantWavePeriod, o.AveragePeriod, o.MeanWaveDirection, o.AirTemperature, o.WaterTemperature)
	if err != nil {
		log.Fatal("Error saving observation:", err)
	}
}

// Given config and observation, tweets latest update
func tweetCurrent(config Config, o Observation) {
	var api *anaconda.TwitterApi
	api = anaconda.NewTwitterApi(config.Token, config.TokenSecret)
	anaconda.SetConsumerKey(config.ConsumerKey)
	anaconda.SetConsumerSecret(config.ConsumerSecret)
	observationOutput := formatObservation(o)
	tweet, err := api.PostTweet(observationOutput, nil)
	if err != nil {
		fmt.Println("update error:", err)
	} else {
		fmt.Println(tweet.Text)
	}
}

// Given URL, returns raw data with recent observations from NBDC
func getDataFromURL(url string) (body []byte) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal("Error fetching data:", err)
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("ioutil error reading resp.Body:", err)
	}
	// fmt.Println("Status:", resp.Status)
	return
}

// Given path to config.js file, loads credentials
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

// Given raw data, parses latest observation and returns Observation struct
func parseData(d []byte) Observation {
	// Each line contains 19 data points
	// Headers are in the first two lines
	// Latest observation data is in the third line
	// Other lines are not needed

	// Extracts relevant data into variable for processing
	var data = string(d[188:281])
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
		log.Fatal("error processing rawtime:", err)
	}
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		log.Fatal("error processing location", err)
	}
	t = t.In(loc)

	// Create Observation struct and populate with parsed data
	var o Observation
	o.Date = t
	o.WindDirection = windcardinal
	o.WindSpeed = windspeedmph
	o.SignificantWaveHeight = waveheightfeet
	o.DominantWavePeriod, err = strconv.Atoi(datafield[9])
	if err != nil {
		log.Fatal("o.AveragePeriod:", err)
	}
	o.AveragePeriod, err = strconv.ParseFloat(datafield[10], 64)
	if err != nil {
		log.Fatal("o.AveragePeriod:", err)
	}
	o.MeanWaveDirection = wavecardinal
	o.AirTemperature = airtempF
	o.WaterTemperature = watertempF

	return o
}

// Given Observation returns formatted text for tweet
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

// Round input to nearest integer given Float64 and returns Float64
func Round(f float64) float64 {
	return math.Floor(f + .5)
}

// RoundPlus truncates a Float64 to a defined number of decimals given an Int and Float64, and returns a Float64
func RoundPlus(f float64, places int) float64 {
	shift := math.Pow(10, float64(places))
	return Round(f*shift) / shift
}

// getTide selects the next tide prediction from the database and return a Tide struct
func getTide() Tide {
	var tide Tide
	err := db.QueryRow("select date, day, time, predictionft, highlow from tidedata where datetime >= current_timestamp order by datetime limit 1").Scan(&tide.Date, &tide.Day, &tide.Time, &tide.PredictionFt, &tide.HighLow)
	if err != nil {
		log.Fatal("getTide function error:", err)
	}
	return tide
}

// processTide returns a formatted string give a Tide struct
func processTide(t Tide) string {
	if t.HighLow == "H" {
		t.HighLow = "High"
	} else {
		t.HighLow = "Low"
	}
	s := "Tide: " + t.HighLow + " " + strconv.FormatFloat(float64(t.PredictionFt), 'f', 1, 64) + " at " + t.Time
	fmt.Println(s)
	return s
}
