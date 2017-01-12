package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
)

const apiURL = "https://aviationweather.gov/adds/dataserver_current/httpparam?datasource=metars&requestType=retrieve&format=xml&mostRecentForEachStation=constraint&hoursBeforeNow=1.25&stationString=LRSV%20LRBS%20LRAR%20LRCV%20LRCK%20EBOS"

type metar struct {
	RawText   string `xml:"raw_text"`
	FlightCat string `xml:"flight_category"`
}

type byStation []metar

func (s byStation) Len() int           { return len(s) }
func (s byStation) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s byStation) Less(i, j int) bool { return s[i].RawText < s[j].RawText }

func main() {
	resp, err := http.Get(apiURL)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	responseBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	metars, err := generateMetars(responseBytes)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(len(metars))
	sort.Sort(byStation(metars))
	for _, met := range metars {
		fmt.Printf("%-4s %s\n", met.FlightCat, met.RawText)
	}
}

func generateMetars(xmlText []byte) ([]metar, error) {
	response := struct {
		TimeTaken int     `xml:"time_taken_ms"`
		Data      []metar `xml:"data>METAR"`
	}{}
	err := xml.Unmarshal(xmlText, &response)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse XML response: %v", err)
	}
	return response.Data, nil
}
