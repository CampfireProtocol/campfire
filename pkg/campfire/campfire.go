// SPDX-License-Identifier: GPL-2.0
/* Campfire Protocol
 *
 * Copyright (C) 2023 Michael Brooks <mike@flake.art>. All Rights Reserved.
 * Written by Michael Brooks (mike@flake.art)
 */

// Package campfire implements the "camp fire" protocol.
package campfire

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/pion/sdp/v3"
	"github.com/pion/webrtc/v3"
)

// Protocol is the protocol name.
const Protocol = "/campfire/1.0.0"

// CampfireChannel is a connection to one or more peers sharing the same pre-shared
// key.
type CampfireChannel interface {
	// Accept returns a connection to a peer.
	Accept() (io.ReadWriteCloser, error)
	// Close closes the camp fire.
	Close() error
	// Errors returns a channel of errors.
	Errors() <-chan error
	// Expired returns a channel that is closed when the camp fire expires.
	Expired() <-chan struct{}
	// Opened returns true if the camp fire is open.
	Opened() bool
}

// CampfireOffer represents an offer that was received from a peer.
type CampfireOffer struct {
	// ID contains the ID of the peer that sent the offer.
	ID string
	// Ufrag contains the username fragment of the peer that sent the offer.
	Ufrag string
	// Pwd contains the password of the peer that sent the offer.
	Pwd string
	// SDP contains the SDP of the offer.
	SDP webrtc.SessionDescription
}

var (
	// ErrClosed is returned when the camp fire is closed.
	ErrClosed = net.ErrClosed
)

func LoadCertificateFromPEMFile(certPath string, keyPath string) (webrtc.Certificate, error) {
	var dtlsCert webrtc.Certificate
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return dtlsCert, err
	}

	keyPem, err := os.ReadFile(keyPath)
	if err != nil {
		return dtlsCert, err
	}

	// Decode PEM key
	block, _ := pem.Decode(certPEM)
	if block == nil || block.Type != "CERTIFICATE" {
		return dtlsCert, errors.New("invalid certificate PEM block")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return dtlsCert, err
	}

	keyBlock, _ := pem.Decode(keyPem)
	if block == nil || block.Type != "CERTIFICATE" {
		return dtlsCert, errors.New("invalid certificate PEM block")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	if err != nil {
		return dtlsCert, err
	}
	dtlsCert = webrtc.CertificateFromX509(privateKey, cert)
	return dtlsCert, nil
}

func (camp *CampfireURI) getTemporalKey(IV string) string {
	currentTime := time.Now().UTC()
	roundedTime := currentTime.Round(time.Hour)
	timeString := roundedTime.Format("2006010215") // Format: YYYYMMDDHH

	pskBytes := []byte(camp.PSK)
	key := []byte(timeString + IV)

	hmacHash := hmac.New(sha256.New, pskBytes)
	hmacHash.Write(key)
	hmacResult := hmacHash.Sum(nil)

	// Convert the HMAC result to hexadecimal
	hmacHex := hex.EncodeToString(hmacResult)
	return hmacHex
}

func (camp *CampfireURI) CampfireOffer(isLocal bool) (*webrtc.SessionDescription, error) {
	sdp_type := webrtc.SDPTypeOffer
	if isLocal {
		sdp_type = webrtc.SDPTypeAnswer
	}
	fingerprint := "5F:F6:3B:46:BE:4B:A7:22:F4:4A:29:F7:C5:4F:35:DA:A9:44:24:1C:CB:93:78:64:FD:38:E3:63:75:46:61:E1"
	sessionID := camp.getTemporalKey(sdp_type.String())
	const SDPTemplate = "v=0\r\no=- %s 1 IN IP4 0.0.0.0\r\ns=-\r\nt=0 0\r\na=fingerprint:sha-256 %s\r\na=extmap-allow-mixed\r\na=group:BUNDLE\r\n"
	mySDP := fmt.Sprintf(SDPTemplate, sessionID, fingerprint)
	/*	const SDPTemplate = `v=0
		o=- 257729325579758 2 IN IP4 127.0.0.1
		s=-
		t=0 0
		a=group:BUNDLE 0
		a=extmap-allow-mixed
		a=msid-semantic: WMS
		m=application 9 UDP/DTLS/SCTP webrtc-datachannel
		c=IN IP4 0.0.0.0
		a=ice-ufrag:PHsS
		a=ice-pwd:FL2ncYZ/dq6j4HQtzNzlTIA=
		a=ice-options:trickle
		a=fingerprint:sha-256 5F:F6:3B:46:BE:4B:A7:22:F4:4A:29:F7:C5:4F:35:DA:A9:44:24:1C:CB:93:78:64:FD:38:E3:63:75:46:61:E1
		a=setup:actpass
		a=mid:0
		a=sctp-port:5000
		a=max-message-size:262144
		a=turn:a.relay.metered.ca username=9d4e8faba9a93ef397554dc4 password=hLxK4U49l6fcZLH0
		`*/
	webrtc_sdp := webrtc.SessionDescription{
		Type: sdp_type,
		SDP:  mySDP,
	}
	return &webrtc_sdp, nil
}

func (camp *CampfireURI) CampfireOfferStruct(isLocal bool) (*webrtc.SessionDescription, error) {
	port := 5000
	//ttl := 64
	rng := 1
	// This flag allows both sides to know the other's creds.
	sdp_type := webrtc.SDPTypeOffer
	if isLocal {
		sdp_type = webrtc.SDPTypeAnswer
	}
	// Both sides need the same Session ID:
	session := camp.getTemporalKey("session")
	// But they need a unique username:
	username := camp.getTemporalKey(sdp_type.String())

	localAddress := &sdp.Address{
		Address: "127.0.0.1",
		//TTL:     &ttl,
		//Range:   &rng,
	}

	remoteAddress := &sdp.Address{
		Address: "0.0.0.0",
		//TTL:     &ttl,
		//Range:   &rng,
	}

	sdpSession := sdp.SessionDescription{
		Origin: sdp.Origin{
			Username:       username[0:16],   // Temporal Username
			SessionID:      numeric(session), // Temporal Session
			SessionVersion: 1,
			NetworkType:    "IN",
			AddressType:    "IP4",
		},
		//Name:        "SDP",
		//Description: "test",
		ConnectionInformation: &sdp.ConnectionInformation{
			NetworkType: "IN",
			AddressType: "IP4",
			Address:     localAddress, // Incorrect IP
		},
	}

	ufragAttribute := sdp.Attribute{
		Key:   "ice-ufrag",
		Value: "your-username-fragment",
	}
	pwdAttribute := sdp.Attribute{
		Key:   "ice-pwd",
		Value: "your-password",
	}
	fingerprintAttribute := sdp.Attribute{
		Key:   "fingerprint",
		Value: "your-username-fragment",
	}
	sdpSession.Attributes = append(sdpSession.Attributes, ufragAttribute)
	sdpSession.Attributes = append(sdpSession.Attributes, pwdAttribute)
	sdpSession.Attributes = append(sdpSession.Attributes, fingerprintAttribute)

	dataMedia := sdp.MediaDescription{
		MediaName: sdp.MediaName{
			Media:   "application",
			Port:    sdp.RangedPort{Value: port, Range: &rng}, // Fill in with your desired port number
			Protos:  []string{"DTLS/SCTP"},
			Formats: []string{strconv.Itoa(port)}, // Fill in with your desired format
		},
		ConnectionInformation: &sdp.ConnectionInformation{
			NetworkType: "IN",
			AddressType: "IP4",
			Address:     remoteAddress,
		},
		Attributes: []sdp.Attribute{
			{
				Key:   "sctp-port",
				Value: strconv.Itoa(port), // Fill in with your desired SCTP port
			},
			{
				Key:   "max-message-size",
				Value: "65535", // Fill in with your desired maximum message size
			},
		},
	}
	sdpSession.MediaDescriptions = append(sdpSession.MediaDescriptions, &dataMedia)
	sdp, err := sdpSession.Marshal()
	if err != nil {
		return nil, err
	}
	mysdp := string(sdp) //strings.Replace(string(sdp), "5000/1", "5000", -1)
	webrtc_sdp := webrtc.SessionDescription{
		Type: sdp_type,
		SDP:  mysdp,
	}

	// Print the generated SDP
	return &webrtc_sdp, nil
}

func numeric(inputString string) uint64 {
	maxDigit := 9
	numericString := ""
	for i := 0; i < len(inputString); i++ {
		charCode := int(inputString[i])
		digit := charCode % (maxDigit + 1) // Ensure digit is between 0 and 9
		numericString += fmt.Sprintf("%d", digit)
	}
	result, _ := strconv.Atoi(numericString)
	return uint64(result)
}

// NewCampfireClient creates a new CampfireClient.
//func (camp *CampfireURI) NewCampfireClient() (*turn.CampfireClient, error) {
//ice := camp.GetICEServers()
// Create a new PeerConnection, this listens for all incoming DataChannel messages
//api := webrtc.NewAPI(webrtc.WithSettingEngine(s))
/*	peerConnection, err := api.NewPeerConnection(webrtc.Configuration{
		Certificates: loadCertificate(),
	})
*/
/*addr := strings.TrimPrefix(opts.Addr, "turn:")
//addr = strings.TrimPrefix(addr, "stun:")
//parts := strings.Split(addr, "@")
if len(parts) == 2 {
	addr = parts[1]
}
if !strings.Contains(addr, ":") {
	// Add default port if missing.
	addr = addr + ":443"
}
if opts.ID == "" {
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("generate random ID: %w", err)
	}
	opts.ID = id.String()
}
block, err := aes.NewCipher(opts.PSK)
if err != nil {
	return nil, fmt.Errorf("create cipher: %w", err)
}
udpAddr, err := net.ResolveUDPAddr("udp", addr)
if err != nil {
	return nil, fmt.Errorf("resolve UDP address: %w", err)
}
conn, err := net.DialUDP("udp", nil, udpAddr)
if err != nil {
	return nil, fmt.Errorf("dial UDP: %w", err)
}
aesgcm, err := cipher.NewGCM(block)
if err != nil {
	return nil, fmt.Errorf("create GCM: %w", err)
}
cli := &CampfireClient{
	id:     opts.ID,
	opts:   opts,
	cipher: aesgcm,
	conn:   conn,
	//offers:     make(chan CampfireOffer, 10),
	//answers:    make(chan CampfireAnswer, 10),
	//candidates: make(chan CampfireCandidate, 10),
	errc:   make(chan error, 10),
	closec: make(chan struct{}),
	log:    slog.Default().With("component", "campfire-client"),
}*/

//go cli.handleIncoming(
//	return nil, nil
//}
