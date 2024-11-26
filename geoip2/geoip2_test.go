package geoip2

import (
	"encoding/json"
	"log"
	"net"
	"testing"

	"github.com/oschwald/geoip2-golang"
)

func TestGeoip2(t *testing.T) {
	db, err := geoip2.Open("/Users/shipengfei/Downloads/GeoLite2-City_20230317/GeoLite2-City.mmdb")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ip := net.ParseIP("223.147.86.61")
	record, err := db.City(ip)
	if err != nil {
		t.Fatal(err)
	}
	city := ConvertToLocalCity(record)
	bs, _ := json.Marshal(city)
	t.Log(string(bs))
	t.Log(city.GetCountryName(), city.GetProvinceName(), city.GetCityName())
}
