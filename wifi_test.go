package main

import (
	"testing"
)

func TestParseWiFiInterface(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		want    string
		wantErr bool
	}{
		{
			name: "typical output with Wi-Fi on en0",
			output: `Hardware Port: Ethernet
Device: en6
Ethernet Address: ab:cd:ef:12:34:56

Hardware Port: Wi-Fi
Device: en0
Ethernet Address: ab:cd:ef:78:90:12

Hardware Port: Thunderbolt Bridge
Device: bridge0
Ethernet Address: ab:cd:ef:34:56:78
`,
			want: "en0",
		},
		{
			name: "Wi-Fi on en1",
			output: `Hardware Port: Ethernet
Device: en0
Ethernet Address: aa:bb:cc:dd:ee:ff

Hardware Port: Wi-Fi
Device: en1
Ethernet Address: 11:22:33:44:55:66
`,
			want: "en1",
		},
		{
			name:    "no Wi-Fi interface",
			output:  "Hardware Port: Ethernet\nDevice: en0\nEthernet Address: aa:bb:cc:dd:ee:ff\n",
			wantErr: true,
		},
		{
			name:    "empty output",
			output:  "",
			wantErr: true,
		},
		{
			name:    "Wi-Fi header but no device line",
			output:  "Hardware Port: Wi-Fi\nEthernet Address: aa:bb:cc:dd:ee:ff\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseWiFiInterface(tt.output)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseSSIDs(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   []string
	}{
		{
			name:   "typical list",
			output: "Preferred networks on en0:\n\tHomeWifi\n\tCoffeeShop\n\tOffice5G\n",
			want:   []string{"HomeWifi", "CoffeeShop", "Office5G"},
		},
		{
			name:   "single SSID",
			output: "Preferred networks on en0:\n\tMyNetwork\n",
			want:   []string{"MyNetwork"},
		},
		{
			name:   "empty list (header only)",
			output: "Preferred networks on en0:\n",
			want:   nil,
		},
		{
			name:   "blank lines ignored",
			output: "Preferred networks on en0:\n\tAlpha\n\n\tBravo\n\t\n",
			want:   []string{"Alpha", "Bravo"},
		},
		{
			name:   "SSID with spaces",
			output: "Preferred networks on en0:\n\tMy Home Network\n\tGuest WiFi 5G\n",
			want:   []string{"My Home Network", "Guest WiFi 5G"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseSSIDs(tt.output)
			if len(got) != len(tt.want) {
				t.Fatalf("got %d SSIDs %v, want %d %v", len(got), got, len(tt.want), tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("SSID[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}
