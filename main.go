package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/fatih/color"
)

const apiURL = "https://aviationweather.gov/adds/dataserver_current/httpparam?datasource=metars&requestType=retrieve&format=xml&mostRecentForEachStation=constraint&hoursBeforeNow=1.25&stationString=LRBS%20LRCT%20LRCK%20LRSB"

type metar struct {
	RawText   string    `xml:"raw_text"`
	StationID string    `xml:"station_id"`
	Time      time.Time `xml:"observation_time"`
	Temp      float64   `xml:"temp_c"`
	Dewpoint  float64   `xml:"dewpoint_c"`
	WDir      int       `xml:"wind_dir_degrees"`
	WSpd      int       `xml:"wind_speed_kt"`
	VisSM     float64   `xml:"visibility_statute_mi"`
	QNHHg     float64   `xml:"altim_in_hg"`
	WX        string    `xml:"wx_string"`
	FlightCat string    `xml:"flight_category"`
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

	fmt.Println()
	sort.Sort(byStation(metars))
	for _, met := range metars {
		printMetarSimple(met)
	}
	fmt.Println()
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

func printMetarSimple(m metar) {
	printer := getColorPrinter(m.FlightCat)
	printer.Printf("%-4s ", m.FlightCat)
	fmt.Println(m.RawText, m.Time)
}

func printMetarFull(m metar) {

}

func getColorPrinter(flightCat string) *color.Color {
	switch flightCat {
	case "VFR", "MVFR":
		return color.New(color.FgGreen, color.Bold)
	case "IFR":
		return color.New(color.FgBlue, color.Bold)
	case "LIFR":
		return color.New(color.FgRed, color.Bold)
	}
	return color.New(color.BgBlack)
}
