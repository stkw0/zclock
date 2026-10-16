package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/adrg/xdg"
	"github.com/glendc/go-external-ip"
	"github.com/ip2location/ip2location-go"
	"github.com/sixdouglas/suncalc"
	flag "github.com/spf13/pflag"
	"net"
	"os"
	"strings"
	"time"
)

const geolocCacheFile = "/tmp/zclock_geoloc_cache"

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}

func getAnimalIdx(hour int) int {
	if hour == 23 || hour == 0 {
		return 0
	}

	for i := 23; i != 1; i -= 2 {
		if hour >= i {
			return (i + 1) / 2
		}

	}

	return 1
}

func floatToByte(f float32) []byte {
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.BigEndian, f)
	if err != nil {
		fmt.Println("binary.Write failed:", err)
	}
	return buf.Bytes()
}

func getPublicIpAddr() string {
	ipCacheFile := "/tmp/zclock_ip_cache"

	if _, err := os.Stat(ipCacheFile); err == nil {
		ip, err := os.ReadFile(ipCacheFile)
		handleError(err)
		return string(ip)
	} else if !errors.Is(err, os.ErrNotExist) {
		panic(err)
	}

	ifaces, err := net.Interfaces()
	handleError(err)

	if !availableIface(ifaces) {
		fmt.Fprintln(os.Stderr, "Network not available")
		os.Exit(1)
	}

	cfg := externalip.DefaultConsensusConfig().WithTimeout(time.Second * 2)
	consensus := externalip.NewConsensus(cfg, nil)
	consensus.UseIPProtocol(4)
	consensus.AddVoter(externalip.NewHTTPSource("https://myexternalip.com/raw"), 3)
	consensus.AddVoter(externalip.NewHTTPSource("http://ifconfig.io/ip"), 1)

	tip, err := consensus.ExternalIP()
	if err != nil {
		panic(err)
	}

	ip := tip.String()

	f, err := os.Create(ipCacheFile)
	handleError(err)
	defer f.Close()

	_, err = f.WriteString(ip)
	handleError(err)

	return ip
}

func getCachedGeolocation() (float64, float64, bool) {
	if _, err := os.Stat(geolocCacheFile); err == nil {
		fd, err := os.Open(geolocCacheFile)
		handleError(err)
		defer fd.Close()

		var lat, long float32
		err = binary.Read(fd, binary.BigEndian, &lat)
		handleError(err)
		err = binary.Read(fd, binary.BigEndian, &long)
		handleError(err)

		return float64(lat), float64(long), true

	} else if !errors.Is(err, os.ErrNotExist) {
		panic(err)
	}

	return 0, 0, false
}

func getGeolocation(ip string) (float64, float64) {
	dbPath, err := xdg.SearchDataFile("IP2LOCATION-LITE-DB5.BIN")
	handleError(err)

	db, err := ip2location.OpenDB(dbPath)
	handleError(err)

	results, err := db.Get_all(ip)
	handleError(err)

	return float64(results.Latitude), float64(results.Longitude)
}

func getSolarNoon(now time.Time, lat, long float64) time.Time {
	var times = suncalc.GetTimes(now, lat, long)
	//var times = suncalc.GetTimesWithObserver(now, suncalc.Observer{lat, long, 0, now.Location()})

	solarNoon := times[suncalc.SolarNoon].Value
	return solarNoon
}

func currentSolarHour(lat, long float64) (int, int) {
	return solarHourAt(time.Now(), lat, long)
}

func solarHourAt(now time.Time, lat, long float64) (int, int) {
	solarNoon := getSolarNoon(now, lat, long)
	// Anchor solar noon at 12:00 UTC so the resulting clock is independent
	// of the machine's timezone and daylight-saving offset.
	noonUTC := solarNoon.UTC()
	referenceNoon := time.Date(noonUTC.Year(), noonUTC.Month(), noonUTC.Day(), 12, 0, 0, 0, time.UTC)
	diffTime := now.UTC().Add(referenceNoon.Sub(solarNoon))

	return diffTime.Hour(), diffTime.Minute()
}

func availableIface(ifaces []net.Interface) bool {
	for _, i := range ifaces {
		f := (net.FlagUp | net.FlagRunning)
		if i.Flags&f == f {
			s, err := os.ReadFile("/sys/class/net/" + i.Name + "/operstate")
			handleError(err)
			if strings.Trim(string(s), "\n") == "up" {
				return true
			}
		}
	}

	return false
}

func main() {
	var printIp = flag.BoolP("ip", "i", false, "Get your public IP")
	var printGeolocation = flag.BoolP("geoloc", "g", false, "Get geolocation coordinates for your public IP")
	var numeric = flag.BoolP("numeric", "n", false, "Show numeric date")
	var forcedIp = flag.String("force", "", "Use an arbitrary IP")
	var forcedLat = flag.Float64("lat", 0, "Force a given latitude")
	var forcedLong = flag.Float64("long", 0, "Force a given longitude")

	flag.Parse()

	cmdMode := false

	var ip string
	if *forcedIp != "" {
		cmdMode = true
		ip = *forcedIp
	} else {
		ip = getPublicIpAddr()
	}

	if *printIp {
		cmdMode = true
		fmt.Println(ip)
	}

	var lat, long float64
	if *forcedLat == 0 && *forcedLong == 0 {
		lat, long = getGeolocation(ip)
	} else if *forcedLat != 0 && *forcedLong != 0 {
		lat = *forcedLat
		long = *forcedLong
	} else if *forcedLat != 0 {
		lat = *forcedLat
		_, long = getGeolocation(ip)
	} else if *forcedLong != 0 {
		long = *forcedLong
		lat, _ = getGeolocation(ip)
	}

	if *printGeolocation {
		cmdMode = true

		fmt.Println("Latitude: ", lat)
		fmt.Println("Longitude: ", long)
	}

	//lat, long, b := getCachedGeolocation()
	hour, minute := currentSolarHour(lat, long)

	if *numeric {
		cmdMode = true
		fmt.Printf("%d:%02d\n", hour, minute)
	}

	if cmdMode {
		return
	}

	idx := getAnimalIdx(hour)
	hours := []rune{'🐀', '🐂', '🐅', '🐇', '🐉', '🐍', '🐎', '🐐', '🐒', '🐓', '🐕', '🐖'}
	fmt.Printf("%c", hours[idx])
}
