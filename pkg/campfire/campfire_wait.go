// SPDX-License-Identifier: GPL-2.0
/* Campfire Protocol
 *
 * Copyright (C) 2023 Michael Brooks <mike@flake.art>. All Rights Reserved.
 * Written by Michael Brooks (mike@flake.art)
 */

package campfire

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/pion/webrtc/v3"
)

// Wait will wait for peers to join at the given location.
func (camp *CampfireURI) Wait(ctx context.Context, cert *webrtc.Certificate) (CampfireChannel, error) {

	iceList, err := camp.GetICEServers()
	if err != nil {
		return nil, fmt.Errorf("No locations: %w", err)
	}
	//log.Debug("Found campfire location", "location", iceList)
	if err != nil {
		return nil, fmt.Errorf("new campfire client: %w", err)
	}

	s := webrtc.SettingEngine{}
	s.SetICECredentials("fKVhbscsMWDGAnBg", "xGjQkAvKIVkBeVTGWcvCQtnVAeapczwa")
	var peerConnection *webrtc.PeerConnection
	if cert != nil {
		peerConnection, err = webrtc.NewPeerConnection(webrtc.Configuration{
			ICEServers:   iceList,
			Certificates: []webrtc.Certificate{*cert},
		})
	} else {
		peerConnection, err = webrtc.NewPeerConnection(webrtc.Configuration{
			ICEServers: iceList,
		})
	}
	if err != nil {
		return nil, fmt.Errorf("bad peer connection: %w", err)
	}
	if len(camp.TURNServers) > 0 {
		turnHost, err := extractHostname(camp.TURNServers[0])
		if err != nil {
			return nil, fmt.Errorf("bad turn host: %w", err)
		}
		addr, err := net.LookupIP(turnHost)
		if err != nil {
			return nil, fmt.Errorf("lookup failed: %w", err)
		}

		// Create a TURN relay ICE candidate with no remote IP
		iceCandidate := webrtc.ICECandidateInit{
			Candidate: "candidate:0 1 UDP 2130706431 " + addr[0].String() + " 3478 typ relay",
		}
		// Add the TURN relay ICE candidate to the peer connection
		err = peerConnection.AddICECandidate(iceCandidate)
	}
	if len(camp.STUNServers) > 0 {
		// todo add STUN canidates if they exist
	}

	if err != nil {
		fmt.Println("Error adding ICE candidate:", err)
	}
	/*
		peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
			fmt.Printf("Connection State has changed %s \n", connectionState.String())
		})
	*/
	/*
		localOffer, err := camp.CampfireOffer(true)
		if err != nil {
			return nil, fmt.Errorf("bad offers: %w", err)
		}

		err = peerConnection.SetLocalDescription(*localOffer)

	*/

	// Create a data channel (optional)
	/*
		dataChannel, err := peerConnection.CreateDataChannel("data", nil)
		if err != nil {
			panic(err)
		}
	*/
	// Create an offer
	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		panic(err)
	}

	localOffer, err := camp.CampfireOffer(true)
	if err != nil {
		panic(err)
	}
	// Set local description
	err = peerConnection.SetLocalDescription(*localOffer)
	if err != nil {
		fmt.Println(offer.SDP)
		//log.Debug(offer.SDP)
		panic(err)
	}

	/*
		if err != nil {
			return nil, fmt.Errorf("local offer: %w\n%s", err, localOffer.SDP)
		}
	*/
	remoteOffer, err := camp.CampfireOffer(false)
	if err != nil {
		return nil, fmt.Errorf("remote description error: %w,\n%s", err, remoteOffer)
	}
	err = peerConnection.SetRemoteDescription(*remoteOffer)

	if err != nil {
		return nil, fmt.Errorf("remote description error: %w,\n%s", err, remoteOffer.SDP)
	}
	// todo - I belive we need to diable this repaly protection:
	s.DisableSRTCPReplayProtection(false)
	s.DetachDataChannels()
	s.SetIncludeLoopbackCandidate(true)

	peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
		d.OnOpen(func() {
			fmt.Printf("DataChannel %s has opened \n", d.Label())
		})

		d.OnMessage(func(m webrtc.DataChannelMessage) {
			fmt.Printf("%s \n", m.Data)
		})
	})

	select {}
	return nil, nil
}

func extractHostname(connectionURL string) (string, error) {
	// Find the position of the first ":" to identify the prefix
	prefixEnd := strings.Index(connectionURL, ":")

	// Extract the hostname part based on the prefix position
	if prefixEnd == -1 {
		return "", fmt.Errorf("Invalid connection URL")
	}

	hostnamePart := connectionURL[prefixEnd+1:]

	// Remove any leading "//" after the prefix
	hostnamePart = strings.TrimPrefix(hostnamePart, "//")

	// Split by "@" to separate username:password and hostname
	parts := strings.SplitN(hostnamePart, "@", 2)
	hostnamePart = parts[len(parts)-1]

	// Split by ":" to handle hostname and optional port
	hostnameAndPort := strings.SplitN(hostnamePart, ":", 2)
	hostname := hostnameAndPort[0]

	return hostname, nil
}

type turnWait struct {
	api      *webrtc.API
	camp     *CampfireURI
	location *Location
	//fireconn     *turn.CampfireClient
	acceptc      chan io.ReadWriteCloser
	closec       chan struct{}
	errc         chan error
	inProgress   map[string]*webrtc.PeerConnection
	log          *slog.Logger
	mu           sync.Mutex
	certificates []webrtc.Certificate
}

// Accept returns a connection to a peer.
func (t *turnWait) Accept() (io.ReadWriteCloser, error) {
	select {
	case <-t.closec:
		return nil, ErrClosed
	case conn := <-t.acceptc:
		return conn, nil
	}
}

/*
// Close closes the camp fire.
func (t *turnWait) Close() error {
	close(t.closec)
	return t.Close()
}
*/
// Opened returns true if the camp fire is opened.
func (t *turnWait) Opened() bool {
	select {
	case <-t.closec:
		return false
	default:
		return true
	}
}

// Errors returns a channel of errors.
func (t *turnWait) Errors() <-chan error { return t.errc }

// Expired returns a channel that is closed when the camp fire expires.
func (t *turnWait) Expired() <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		defer close(ch)
		select {
		case <-t.closec:
		case <-time.After(time.Until(t.location.ExpiresAt)):
		}
	}()
	return ch
}

/*
	func (t *turnWait) handleIncomingOffers() {
		offers := t.fireconn.Offers()
		for {
			select {
			case <-t.closec:
				return
			case err := <-t.fireconn.Errors():
				t.errc <- fmt.Errorf("campfire client: %w", err)
			case offer := <-offers:
				if offer.Ufrag != t.location.RemoteUfrag() || offer.Pwd != t.location.RemotePwd() {
					t.log.Warn("received offer with unexpected ufrag/pwd", "ufrag", offer.Ufrag, "pwd", offer.Pwd)
					continue
				}
				go t.handleNewPeerConnection(&offer)
			}
		}
	}
*/
func (t *turnWait) SetCertificatefromX509(cert webrtc.Certificate) {
	// Lock to ensure thread-safe modification of certificates slice
	t.mu.Lock()
	t.certificates = []webrtc.Certificate{cert}
	defer t.mu.Unlock()
}

/*
	func (t *turnWait) handleIncomingCandidates() {
		candidates := t.fireconn.Candidates()
		for {
			select {
			case <-t.closec:
				return
			case cand := <-candidates:
				t.mu.Lock()
				conn, ok := t.inProgress[cand.ID]
				if !ok {
					t.log.Warn("Received candidate for unknown connection", "id", cand.ID)
					t.mu.Unlock()
					continue
				}
				t.log.Debug("Received remote ice candidate", "candidate", cand)
				err := conn.AddICECandidate(cand.Cand)
				if err != nil {
					t.log.Error("Error adding ice candidate", "error", err)
				}
				t.mu.Unlock()
			}
		}
	}
*/
/*
func (t *turnWait) handleNewPeerConnection(offer *CampfireOffer) {
	t.mu.Lock()
	t.log.Debug("Creating new peer connection", "offer", offer)

	iceList, err := t.camp.GetICEServers()
	if err != nil {
		t.log.Warn("failed to generate ice list", "err", err)
	}
	pc, err := t.api.NewPeerConnection(webrtc.Configuration{
		ICEServers:   iceList,
		Certificates: t.certificates,
	})
	if err != nil {
		t.mu.Unlock()
		t.errc <- fmt.Errorf("new peer connection: %w", err)
		return
	}
	t.inProgress[offer.ID] = pc
	t.mu.Unlock()
	pc.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}
		t.log.Debug("Sending local ice candidate", "candidate", c)
		err := t.fireconn.SendCandidate(offer.ID, t.location.RemoteUfrag(), t.location.RemotePwd(), c)
		if err != nil {
			t.log.Warn("failed to send ice candidate", "err", err)
		}
	})
	pc.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		t.log.Debug("ICE connection state changed", "state", state)
		if state == webrtc.ICEConnectionStateConnected || state == webrtc.ICEConnectionStateCompleted {
			t.mu.Lock()
			delete(t.inProgress, offer.ID)
			t.mu.Unlock()
		}
	})
	pc.OnDataChannel(func(dc *webrtc.DataChannel) {
		t.log.Debug("Received data channel", "label", dc.Label())
		if dc.Label() != string(t.location.PSK) {
			t.log.Warn("received data channel with unexpected label", "label", dc.Label())
			return
		}
		dc.OnOpen(func() {
			rw, err := dc.Detach()
			if err != nil {
				t.errc <- fmt.Errorf("detach data channel: %w", err)
				return
			}
			t.acceptc <- rw
		})
	})
	t.log.Debug("remote SDP:", "sdp", offer.SDP.SDP)
	err = pc.SetRemoteDescription(offer.SDP)
	if err != nil {
		t.errc <- fmt.Errorf("set remote description: %w", err)
		return
	}
	answer, err := pc.CreateAnswer(nil)
	if err != nil {
		t.errc <- fmt.Errorf("create answer: %w", err)
		return
	}
	t.log.Debug("local SDP:", "sdp", answer.SDP)
	err = pc.SetLocalDescription(answer)
	if err != nil {
		t.errc <- fmt.Errorf("set local description: %w", err)
		return
	}
	t.log.Debug("Sending answer", "answer", answer)
	err = t.fireconn.SendAnswer(offer.ID, t.location.RemoteUfrag(), t.location.RemotePwd(), answer)
	if err != nil {
		t.errc <- fmt.Errorf("send answer: %w", err)
		return
	}
}
*/
