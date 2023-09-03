// SPDX-License-Identifier: GPL-2.0
/* Campfire Protocol
 *
 * Copyright (C) 2023 Michael Brooks <mike@flake.art>. All Rights Reserved.
 * Written by Michael Brooks (mike@flake.art)
 */
package campfire

import (
	"context"
	"io"
)

// Join will attempt to join the peer waiting at the given location.
func Join(ctx context.Context, camp *CampfireURI) (io.ReadWriteCloser, error) {
	/*log := context.LoggerFrom(ctx).With("protocol", "campfire")
	location, err := Find(camp.PSK, camp.TURNServers)
	if err != nil {
		return nil, fmt.Errorf("find campfire: %w", err)
	}
	if !strings.HasPrefix(location.TURNServer, "turn:") {
		location.TURNServer = "turn:" + location.TURNServer
	}
	// Parse the selected turn server
	fireconn, err := camp.NewCampfireClient()
	if err != nil {
		return nil, fmt.Errorf("new campfire client: %w", err)
	}
	defer fireconn.Close()
	s := webrtc.SettingEngine{}
	s.DetachDataChannels()
	s.SetIncludeLoopbackCandidate(true)
	api := webrtc.NewAPI(webrtc.WithSettingEngine(s))

	iceList, err := camp.GetICEServers()
	if err != nil {
		return nil, err
	}
	pc, err := api.NewPeerConnection(webrtc.Configuration{
		ICEServers: iceList,
	})

	if err != nil {
		return nil, fmt.Errorf("create peer connection: %w", err)
	}
	errs := make(chan error, 1)
	acceptc := make(chan io.ReadWriteCloser)
	dc, err := pc.CreateDataChannel(string(location.PSK), nil)
	if err != nil {
		return nil, fmt.Errorf("create data channel: %w", err)
	}
	dc.OnOpen(func() {
		log.Debug("Data channel opened")
		rw, err := dc.Detach()
		if err != nil {
			errs <- fmt.Errorf("detach data channel: %w", err)
			return
		}
		acceptc <- rw
	})
	offer, err := pc.CreateOffer(nil)
	if err != nil {
		return nil, fmt.Errorf("create offer: %w", err)
	}
	err = pc.SetLocalDescription(offer)
	if err != nil {
		return nil, fmt.Errorf("set local description: %w", err)
	}
	log.Debug("Sending offer", "offer", offer.SDP)
	err = fireconn.SendOffer(location.LocalUfrag(), location.LocalPwd(), offer)
	if err != nil {
		return nil, fmt.Errorf("send offer: %w", err)
	}
	log.Debug("Waiting for answer")
	*/
	/*var answer turn.CampfireAnswer
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errs:
		return nil, err
	case answer = <-fireconn.Answers():
	}*/
	/*
		log.Debug("Received answer", "answer", answer.SDP)
		err = pc.SetRemoteDescription(answer.SDP)
		if err != nil {
			return nil, fmt.Errorf("set remote description: %w", err)
		}
		connectedc := make(chan struct{})
		pc.OnICECandidate(func(c *webrtc.ICECandidate) {
			if c == nil {
				return
			}
			select {
			case <-connectedc:
				return
			default:
			}
			log.Debug("Sending local ICE candidate", "candidate", c.String())
			err = fireconn.SendCandidate("", location.LocalUfrag(), location.LocalPwd(), c)
			if err != nil {
				errs <- fmt.Errorf("send ice candidate: %w", err)
				return
			}
		})
		pc.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
			log.Debug("Peer connection state change", "state", state.String())
			if state == webrtc.PeerConnectionStateConnected {
				close(connectedc)
			}
			if state == webrtc.PeerConnectionStateDisconnected || state == webrtc.PeerConnectionStateFailed {
				errs <- fmt.Errorf("peer connection state: %s", state.String())
			}
		})
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case <-connectedc:
					return
				case cand := <-fireconn.Candidates():
					log.Debug("Received remote ICE candidate", "candidate", cand.Cand)
					err = pc.AddICECandidate(cand.Cand)
					if err != nil {
						errs <- fmt.Errorf("add ice candidate: %w", err)
						return
					}
				}
			}
		}()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case err := <-fireconn.Errors():
			return nil, err
		case err := <-errs:
			return nil, err
		case rw := <-acceptc:
			return rw, nil
		}
	*/
	return nil, nil
}
