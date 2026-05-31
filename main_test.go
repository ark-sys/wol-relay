package main

import (
	"bytes"
	"testing"
)

func TestNormalizeMAC(t *testing.T) {
	cases := []struct {
		in      string
		wantErr bool
	}{
		{"AA:BB:CC:DD:EE:FF", false},
		{"aa-bb-cc-dd-ee-ff", false},
		{"AABBCCDDEEFF", true},
		{"AA:BB:CC:DD:EE", true},
		{"ZZ:BB:CC:DD:EE:FF", true},
	}
	for _, c := range cases {
		_, err := normalizeMAC(c.in)
		if (err != nil) != c.wantErr {
			t.Errorf("normalizeMAC(%q) error=%v wantErr=%v", c.in, err, c.wantErr)
		}
	}
}

func TestBuildMagicPacket(t *testing.T) {
	mac, err := normalizeMAC("AA:BB:CC:DD:EE:FF")
	if err != nil {
		t.Fatal(err)
	}
	packet := buildMagicPacket(mac)
	if len(packet) != 102 {
		t.Fatalf("expected 102 bytes, got %d", len(packet))
	}
	header := bytes.Repeat([]byte{0xFF}, 6)
	if !bytes.Equal(packet[:6], header) {
		t.Errorf("header mismatch: %x", packet[:6])
	}
	for i := 0; i < 16; i++ {
		offset := 6 + i*6
		if !bytes.Equal(packet[offset:offset+6], mac) {
			t.Errorf("MAC repetition %d mismatch", i)
		}
	}
}