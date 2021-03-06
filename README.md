# BuoyBot
BuoyBot is a twitter bot that periodically tweets updates from NBDC Station 46026 (the San Francisco Buoy).
Tide data is from NOAA Station 9414275 (Ocean Beach, San Francisco, California).

BuoyBot is live on Twitter: https://twitter.com/SFBuoy

Feature requests and code contributions are welcome.

## Usage
All testing has been done on Ubuntu 18.04 LTS.

BuoyBot runs at 10 minutes past the hour since NBDC observations are taken at 50 minutes past the hour and updates are available approximately 15 minutes thereafter.

BuoyBot is designed to be run at pre-defined intervals via Cron. Crontab.txt contains the cron entry required to run BuoyBot. Twitter and database credentials need to be saved in a config.json file. The configexample.json file contains the template that should be used. Path to config.js is stored in a CONFIGPATH environment variable that needs to be configured by the user.

BuoyBot saves its hourly observations in a Postgres database. This needs to be configured by the user or the database code needs to be removed. Observations.sql contains the necessary Postgres table schema for BuoyBot.

## Tide Data Note
Buoybot presumes that the database contains relevant tide predictions. This data can be obtained from github.com/johnbeil/tidecrawler


## Development Roadmap:
- Use environment variables rather than config.json for credentials
- Tweet at high and low tides
