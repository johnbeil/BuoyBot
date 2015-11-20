# BuoyBot

BuoyBot is a twitter bot that tweets updates from NBDC Station 46026 (the San Francisco Buoy).

BuoyBot is live on Twitter: https://twitter.com/SFBuoy

Feature requests and code contributions are welcome.

## Usage
BuoyBot is designed to be run at pre-defined intervals via cron. All testing has been done on Ubuntu 14.04 LTS.

Crontab.txt contains the cron entry required to run BuoyBot at 0010, 0610, 0910, 1210, 1510, and 1810 daily. BuoyBot runs at 10 minutes past the hour since NBDC observations are taken at 50 minutes past the hour and updates are available approximately 15 minutes thereafter.

Twitter credentials need to be saved in a config.json file. The configexample.json file contains the template that should be used. Note config.json is referenced in an absolute path in the loadConfig function in buoybot.go. This needs to be updated for your local machine.

## Development Roadmap:
- Reply to @ mentions with latest observation
- Auto follow people who follow BuoyBot
- Store history and make observations upon novel observation (e.g. largest wave height, coldest water temp, etc)