package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
)

const apiURL = "https://aviationweather.gov/adds/dataserver_current/httpparam?datasource=metars&requestType=retrieve&format=xml&mostRecentForEachStation=constraint&hoursBeforeNow=1.25&stationString="

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
	SkyCond   []struct {
		Cover     string `xml:"sky_cover,attr"`
		CloudBase int    `xml:"cloud_base_ft_agl,attr"`
	} `xml:"sky_condition"`
}

type byStation []metar

func (s byStation) Len() int           { return len(s) }
func (s byStation) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s byStation) Less(i, j int) bool { return s[i].RawText < s[j].RawText }

func main() {

	airports, err := getSettingsCmdLine()
	if err != nil {
		log.Fatal("usage: metar ICAO1 ICAO2 ...")
	}
	queryString := apiURL + airports

	resp, err := http.Get(queryString)
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
	fmt.Println(m.RawText)
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

func getSettingsDotFile() (string, error) {
	file, err := os.Open(path.Join(os.Getenv("HOME"), ".metar.json"))
	if err != nil {
		return "", fmt.Errorf("could not open: %v", err)
	}
	defer file.Close()

	contents, err := ioutil.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("can't read: %v", err)
	}

	type config struct {
		Stations []string `xml:"stations"`
	}
	conf := config{}
	err = json.Unmarshal(contents, &conf)
	if err != nil {
		return "", fmt.Errorf("can't parse JSON: %v", err)
	}
	return strings.Join(conf.Stations, "%20"), nil
}

func getSettingsCmdLine() (string, error) {
	if !(len(os.Args) >= 2) {
		return "", fmt.Errorf("no command line args provided")
	}
	return strings.Join(os.Args[1:], "%20"), nil
}
