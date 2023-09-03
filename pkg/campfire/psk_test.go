// SPDX-License-Identifier: GPL-2.0
/* Campfire Protocol
 *
 * Copyright (C) 2023 Michael Brooks <mike@flake.art>. All Rights Reserved.
 * Written by Michael Brooks (mike@flake.art)
 */

package campfire

import (
	"testing"
	"time"
)

func FuzzGeneratePSK(f *testing.F) {
	Now = func() time.Time {
		return time.Unix(0, 0)
	}
	testcases := []string{
		// Some randomly generated 32 byte PSKs.
		string(MustGeneratePSK()),
		string(MustGeneratePSK()),
		string(MustGeneratePSK()),
		string(MustGeneratePSK()),
		string(MustGeneratePSK()),
		string(MustGeneratePSK()),
		string(MustGeneratePSK()),
		string(MustGeneratePSK()),
		string(MustGeneratePSK()),
		string(MustGeneratePSK()),
	}
	seenPSKs := make(map[string]struct{})
	for _, tc := range testcases {
		seenPSKs[tc] = struct{}{}
		f.Add(tc)
	}
	f.Fuzz(func(t *testing.T, psk string) {
		newPSK, err := GeneratePSK()
		if err != nil {
			t.Fatal(err)
		}
		if string(newPSK) == psk {
			t.Fatal("duplicate PSK generated")
		}
		if _, ok := seenPSKs[string(newPSK)]; ok {
			t.Fatal("duplicate PSK generated")
		}
		seenPSKs[string(newPSK)] = struct{}{}
	})
}
