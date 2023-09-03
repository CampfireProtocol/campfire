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
	"testing"

	"github.com/pion/webrtc/v3"
)

func TestCampfire(t *testing.T) {
	var nilCert *webrtc.Certificate
	t.Parallel()

	ctx := context.Background()
	turnAddr := setupTest(t)
	campURI := fmt.Sprintf("camp://fingerprint/?0=9d4e8faba9a93ef397554dc4:hLxK4U49l6fcZLH0@%s#abcdefghijklmnopqrstuvwx12345678", turnAddr)
	ourcamp, err := ParseCampfireURI(campURI)
	if err != nil {
		t.Fatal(err)
	}

	cf, err := ourcamp.Wait(ctx, nilCert)
	if err != nil {
		t.Fatal(err)
	}

	waitErrs := make(chan error)
	go func() {
		defer close(waitErrs)
		conn, err := cf.Accept()
		if err != nil {
			waitErrs <- err
			return
		}
		defer conn.Close()
		_, err = conn.Write([]byte("hello"))
		if err != nil {
			waitErrs <- err
			return
		}
		b := make([]byte, 5)
		n, err := conn.Read(b)
		if err != nil {
			waitErrs <- err
			return
		}
		if string(b[:n]) != "world" {
			waitErrs <- fmt.Errorf("expected 'world' got %s", string(b[:n]))
			return
		}
	}()

	conn, err := Join(ctx, ourcamp)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	b := make([]byte, 5)
	n, err := conn.Read(b)
	if err != nil {
		t.Fatal(err)
	}
	if string(b[:n]) != "hello" {
		t.Fatalf("expected 'hello' got %s", string(b[:n]))
	}
	_, err = conn.Write([]byte("world"))
	if err != nil {
		t.Fatal(err)
	}
	for err := range waitErrs {
		t.Fatal(err)
	}
}

func setupTest(t *testing.T) (turnServer string) {
	t.Helper()
	/*server, err := turn.NewServer(&turn.Options{
		PublicIP:        "127.0.0.1",
		RelayAddressUDP: "0.0.0.0",
		ListenUDP:       ":0",
		EnableCampfire:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		server.Close()
	})*/
	return "" //fmt.Sprintf("127.0.0.1:%d", server.ListenPort())
}
