// openweather query openweather.org API
package openweather

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
)

const (
	currentUri = "http://api.openweathermap.org/data/2.5/weather?q=%s&units=metric"
	IMPERIAL   = 0
	METRIC     = 1
)

type Coord struct {
	Lon float64
	Lat float64
}

type Sys struct {
	Country string
	Sunrise int
	Sunset  int
}

type Weather struct {
	Id          int
	Main        string
	Description string
	Icon        string
}

type Main struct {
	Temp       float64
	UTemp      float64
	TempUnits  string
	Temp_min   float64
	Temp_max   float64
	Pressure   float64
	UPressure  float64
	Sea_level  float64
	Grnd_level float64
	Humidity   float64
}

type Wind struct {
	Speed  float64
	USpeed float64
	Deg    float64
	Units  string
}

func (w Wind) Card() (dir string) {
	var dirs = []string{"N", "NNE", "NE", "ENE", "E", "ESE", "SE", "SSE",
		"S", "SSW", "SW", "WSW", "W", "WNW", "NW", "NNW"}
	var i = float64(w.Deg+11.25) / 22.5
	dir = dirs[int(math.Mod(i, 16))]
	return
}

func (w Wind) String() (out string) {
	out = fmt.Sprintf("%v @ %.2f %v", w.Card(), w.USpeed, w.Units)
	return
}

type Rain struct {
	H3 float64 `json:"3h"`
}

type Clouds struct {
	All int
}

//CurrentData holds data for current conditions requests
type CurrentData struct {
	Coord   *Coord
	Sys     *Sys
	Weather []*Weather
	Base    string
	Main    *Main
	Wind    *Wind
	Rain    *Rain
	Clouds  *Clouds
	Dt      int
	Id      int
	Name    string
	Message string
}

func (c CurrentData) String() string {
	return fmt.Sprintf("Current conditions for %s: %s %.2f%v Humidity: %v%% Wind: %v Clouds: %v%%",
		c.Name, c.Weather[0].Description, c.Main.UTemp, c.Main.TempUnits,
		c.Main.Humidity, c.Wind, c.Clouds.All)
}

//Fill in unit values
func (c CurrentData) genUnits(units int) {
	if &c.Wind == nil || &c.Wind.Speed == nil {
		return
	}
	if &c.Main == nil || &c.Main.Temp == nil {
		return
	}
	if units == IMPERIAL {
		c.genImperial()
		return
	}
	c.genMetric()
	return
}

func (c CurrentData) genImperial() {
	//Wind speed meters per second to mph
	c.Wind.USpeed = c.Wind.Speed * 2.236936
	c.Wind.Units = "mph"
	//Temperature C to F
	c.Main.UTemp = ((c.Main.Temp) * 1.8) + 32
	c.Main.TempUnits = "°F"
}

func (c CurrentData) genMetric() {
	//Wind speed meters per second to kph
	c.Wind.USpeed = c.Wind.Speed * 3.6
	c.Wind.Units = "kph"
	//Temperature is in C
	c.Main.UTemp = c.Main.Temp
	c.Main.TempUnits = "°C"
}

// CurrentCond returns the current conditions at location
func CurrentCond(location string, units int) (jsonObj CurrentData, err error) {
	var resp *http.Response
	resp, err = http.Get(fmt.Sprintf(currentUri, url.QueryEscape(location)))
	if err != nil {
		return
	}

	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&jsonObj)
	if err != nil {
		return
	}

	if jsonObj.Message != "" {
		err = errors.New(jsonObj.Message)
		return
	}

	jsonObj.genUnits(units)

	if jsonObj.Name == "" {
		jsonObj.Name = fmt.Sprintf("Long: %v Lat: %v", jsonObj.Coord.Lon, jsonObj.Coord.Lat)
	}

	return
}
