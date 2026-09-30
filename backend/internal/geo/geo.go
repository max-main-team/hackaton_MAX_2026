package geo

import "strings"

type Coords struct {
	Lat float64
	Lng float64
}

var cities = map[string]Coords{
	"санкт-петербург":  {Lat: 59.9386, Lng: 30.3141},
	"спб":              {Lat: 59.9386, Lng: 30.3141},
	"питер":            {Lat: 59.9386, Lng: 30.3141},
	"москва":           {Lat: 55.7558, Lng: 37.6173},
	"казань":           {Lat: 55.7963, Lng: 49.1088},
	"екатеринбург":     {Lat: 56.8389, Lng: 60.6057},
	"новосибирск":      {Lat: 55.0084, Lng: 82.9357},
	"нижний новгород":  {Lat: 56.3269, Lng: 44.0059},
	"великий новгород": {Lat: 58.5215, Lng: 31.2755},
	"краснодар":        {Lat: 45.0355, Lng: 38.9753},
}

func Lookup(city string) (Coords, bool) {
	c, ok := cities[strings.ToLower(strings.TrimSpace(city))]
	return c, ok
}
