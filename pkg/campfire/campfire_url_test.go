// SPDX-License-Identifier: GPL-2.0
/* Campfire Protocol
 *
 * Copyright (C) 2023 Michael Brooks <mike@flake.art>. All Rights Reserved.
 * Written by Michael Brooks (mike@flake.art)
 */

package campfire

import "testing"

func TestCampfireURI(t *testing.T) {
	uri := "camp://fingerprint?0=9d4e8faba9a93ef397554dc4:hLxK4U49l6fcZLH0@a.relay.metered.ca#abcdefghijklmnopqrstuvwx12345678"
	campfire, err := ParseCampfireURI(uri)
	if err != nil {
		t.Fatal(err)
	}
	encoded := campfire.EncodeURI()
	if encoded != uri {
		t.Fatalf("Expected %s, got %s", uri, encoded)
	}
}
