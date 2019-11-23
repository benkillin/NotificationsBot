package main
// Copyright (c) 2019 Benkillin. 
// This program is distributed under the terms of the GNU Affero General Public License..
// See LICENSE for the full license.

import (
    "testing"
)

func TestHello(t *testing.T) {
    want := "Hello, world!"
    if got := hello(); got != want {
        t.Errorf("Hello() = %q, want %q", got, want)
    }
}

// TODO: actual unit tests??? somehow with the discord API???
