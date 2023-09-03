// SPDX-License-Identifier: GPL-2.0
/* Campfire Protocol
 *
 * Copyright (C) 2023 Michael Brooks <mike@flake.art>. All Rights Reserved.
 * Written by Michael Brooks (mike@flake.art)
 */
package campfire

import "crypto/rand"

var validPSKChars = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// GeneratePSK generates a random PSK of length PSKSize.
func GeneratePSK() ([]byte, error) {
	out := make([]byte, PSKSize)
	if _, err := rand.Read(out); err != nil {
		return nil, err
	}
	for i, b := range out {
		out[i] = validPSKChars[b%byte(len(validPSKChars))]
	}
	return out, nil
}

// MustGeneratePSK generates a random PSK of length PSKSize.
// It panics if an error occurs.
func MustGeneratePSK() []byte {
	psk, err := GeneratePSK()
	if err != nil {
		panic(err)
	}
	return psk
}
