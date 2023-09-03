/*
Copyright 2023 Avi Zimmerman <avi.zimmerman@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package campfire

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/pion/webrtc/v3"
)

const (
	defaultStunHost = "stun.l.google.com"
	defaultStunPort = "19302"
	defaultTurnHost = "a.relay.metered.ca"
	defaultTurnPort = "443"
	defaultTurnUser = "9d4e8faba9a93ef397554dc4"
	defaultTurnCred = "hLxK4U49l6fcZLH0"
)

// CampfireURI represents the components camp from a camp URL.
type CampfireURI struct {
	PublicKeyFingerprint string
	FullPath             string
	Arguments            string
	TURNServers          []string
	STUNServers          []string
	WebsocketServers     []string
	HTTPServers          []string
	PSK                  string
}

// ParseCampfireURI parses the given rawURL and returns a CampfireURI struct.
func ParseCampfireURI(rawURL string) (*CampfireURI, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	campURL := &CampfireURI{}
	campURL.PublicKeyFingerprint = u.Host
	campURL.FullPath = u.Path
	campURL.PSK = u.Fragment

	queryParams, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return nil, err
	}

	for i := 0; ; i++ {
		serverURL := queryParams.Get(strconv.Itoa(i))
		if serverURL == "" {
			break
		}
		// Delete the query param so we can keep track of the `Arguments`
		queryParams.Del(strconv.Itoa(i))

		decodedServerURL, err := url.QueryUnescape(serverURL)
		if err != nil {
			return nil, err
		}
		lowerServerURL := strings.ToLower(decodedServerURL)
		switch {
		case strings.HasPrefix(lowerServerURL, "turn:"):
			// Fix a common typo  turn:// isn't a valid connection string.
			decodedServerURL = strings.Replace(decodedServerURL, "turn://", "turn:", -1)
			campURL.TURNServers = append(campURL.TURNServers, decodedServerURL)
		case strings.HasPrefix(lowerServerURL, "stun:"):
			decodedServerURL = strings.Replace(decodedServerURL, "stun://", "stun:", -1)
			campURL.STUNServers = append(campURL.STUNServers, decodedServerURL)
		case strings.HasPrefix(lowerServerURL, "wss://") || strings.HasPrefix(lowerServerURL, "ws://"):
			campURL.WebsocketServers = append(campURL.WebsocketServers, decodedServerURL)
		case strings.HasPrefix(lowerServerURL, "http://") || strings.HasPrefix(lowerServerURL, "https://"):
			campURL.HTTPServers = append(campURL.HTTPServers, decodedServerURL)
		}
	}

	// Any query params that are not numeric are stored:
	campURL.Arguments = queryParams.Encode()

	if len(u.Fragment) > 0 {
		campURL.PSK = u.Fragment
	}

	return campURL, nil
}

// EncodeURI encodes the CampfireURI into a string.
func (camp *CampfireURI) EncodeURI() string {
	queryParams := make(url.Values)
	i := 0

	for _, server := range camp.TURNServers {
		queryParams.Add(strconv.Itoa(i), server)
		i++
	}
	for _, server := range camp.STUNServers {
		queryParams.Add(strconv.Itoa(i), server)
		i++
	}
	for _, server := range camp.WebsocketServers {
		queryParams.Add(strconv.Itoa(i), server)
		i++
	}
	for _, server := range camp.HTTPServers {
		queryParams.Add(strconv.Itoa(i), server)
		i++
	}

	// We need atleast one connection canidate.
	if i == 0 && defaultTurnHost != "" {
		queryParams.Add(strconv.Itoa(i), "turn:"+defaultTurnUser+":"+defaultTurnCred+"@"+defaultTurnHost)
	}

	query := camp.Arguments

	// If it isn't empty we need to allow for more params:
	if query != "" {
		query += "&"
	}
	// Add the query Arguments:
	query += queryParams.Encode()

	u := url.URL{
		Scheme:   "camp",
		Host:     camp.PublicKeyFingerprint,
		Path:     camp.FullPath,
		RawQuery: query,
		Fragment: string(camp.PSK),
	}

	return u.String()
}

func parseTurnURL(turnURL string) (*webrtc.ICEServer, error) {
	serverURL := turnURL
	parts := strings.SplitN(turnURL, "@", 2)
	user := "-"
	pass := "-"
	// Password is optional
	if len(parts) == 2 {
		credentials := strings.SplitN(parts[0], ":", 2)
		if len(credentials) == 2 {
			user = credentials[0]
			pass = credentials[1]
		}
		// Remove the username and password from the connection string:
		serverURL = "turn:" + parts[1]
	}

	iceServer := webrtc.ICEServer{
		URLs:       []string{serverURL},
		Username:   user,
		Credential: pass,
	}

	return &iceServer, nil
}

// EncodeURI encodes the CampfireURI into a string.
func (camp *CampfireURI) GetICEServers() ([]webrtc.ICEServer, error) {
	iceServers := []webrtc.ICEServer{}

	for _, serverURL := range camp.TURNServers {
		turnServer, err := parseTurnURL(serverURL)
		if err != nil {
			return nil, err
		}
		iceServers = append(iceServers, *turnServer)
	}

	for _, serverURL := range camp.STUNServers {
		iceServers = append(iceServers, webrtc.ICEServer{
			URLs: []string{serverURL},
		})
	}

	return iceServers, nil
}
