package main

import (
	"io"
	"log"
	"net/http"
	"os"
)

const apiURL = "https://aviationweather.gov/adds/dataserver_current/httpparam?datasource=metars&requestType=retrieve&format=xml&mostRecentForEachStation=constraint&hoursBeforeNow=1.25&stationString=LRBS"

type metar struct {
	rawText string
}

func main() {
	resp, err := http.Get(apiURL)
	if err != nil {
		log.Fatal("couldn't connect to NOAA service")
	}
	defer resp.Body.Close()
	io.Copy(os.Stdout, resp.Body)
}
