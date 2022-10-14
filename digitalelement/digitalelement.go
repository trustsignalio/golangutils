// Package digitalelement is written to de data written in a MMDB file
// the MMDB file contain the country,region and city code for particular
// ip ranges and then we get the corresponding name of country with the
// help of data map initialized when connecting to the database
package digitalelement

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/oschwald/maxminddb-golang"
)

var (
	ipv4dbReader *maxminddb.Reader
	ipv6dbReader *maxminddb.Reader
	countryCodes = make(map[string]string)
	regionCodes  = make(map[string]string)
	cityCodes    = make(map[string]string)
)

// Record struct contains all the information regarding IP which
// digitalelement database can provide
type Record struct {
	Country, Region, City string
	Lat, Long             float64
}

type ipinfo struct {
	Country string  `maxminddb:"country"`
	City    string  `maxminddb:"city"`
	Region  string  `maxminddb:"region"`
	Lat     float64 `maxminddb:"lat"`
	Long    float64 `maxminddb:"long"`
}

// DataFiles struct contains the path to all the files which are required to connect
// to the digitalelement database and get the data in the desired format
type DataFiles struct {
	V4Database  string
	V6Database  string
	CountryCode string
	RegionCode  string
	CityCode    string

	CountryCorrection string
	CityCorrection    string
}

func openFile(filepath string) *os.File {
	f, err := os.Open(filepath)
	if err != nil {
		panic(err)
	}
	return f
}

func createMapFromCsv(filepath string, mapObj map[string]string, k, v int) {
	f := openFile(filepath)
	defer f.Close()
	csvReader := csv.NewReader(bufio.NewReader(f))

	for {
		line, err := csvReader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			break
		}
		code := line[k]
		val := line[v]

		mapObj[code] = val
	}
}

// Connect method will open a connection to the database which will only be closed
func Connect(df *DataFiles) {
	var err error
	ipv4dbReader, err = maxminddb.Open(df.V4Database)
	if err != nil {
		panic(err)
	}
	ipv6dbReader, err = maxminddb.Open(df.V6Database)
	if err != nil {
		panic(err)
	}

	// create a map by parsing the CSV files
	createMapFromCsv(df.CountryCode, countryCodes, 6, 1)
	createMapFromCsv(df.RegionCode, regionCodes, 3, 2)
	createMapFromCsv(df.CityCode, cityCodes, 4, 2)

	// fix the map codes
	createMapFromCsv(df.CountryCorrection, countryCodes, 0, 1)
	createMapFromCsv(df.CityCorrection, cityCodes, 0, 1)

	for k, v := range countryCodes {
		countryCodes[k] = strings.ToUpper(v)
	}
}

// Lookup method will return the information about any given ip address
func Lookup(ip string) Record {
	netIP := net.ParseIP(ip)
	ipv4 := netIP.To4()
	ipv6 := netIP.To16()

	ipData := Record{}
	record := &ipinfo{}

	var err error
	if ipv4 != nil {
		err = ipv4dbReader.Lookup(netIP, &record)
	} else if ipv6 != nil {
		err = ipv6dbReader.Lookup(netIP, &record)
	}
	if err != nil || len(record.Country) == 0 {
		ipData.Country = "XX"
		return ipData
	}
	ipData.Country = countryCodes[record.Country]
	ipData.Region = regionCodes[record.Region]
	ipData.City = cityCodes[record.City]
	ipData.Lat = record.Lat
	ipData.Long = record.Long

	return ipData
}

// Close method will close the digitalelement db files
func Close() {
	fmt.Println("closing digitalelement db files!!")
	if ipv4dbReader != nil {
		ipv4dbReader.Close()
	}
	if ipv6dbReader != nil {
		ipv6dbReader.Close()
	}
}
