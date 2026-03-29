package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
)

// detectWiFiInterface finds the macOS Wi-Fi hardware port name (e.g. "en0").
func detectWiFiInterface() (string, error) {
	out, err := exec.Command("networksetup", "-listallhardwareports").Output()
	if err != nil {
		return "", fmt.Errorf("failed to list hardware ports: %w", err)
	}
	return parseWiFiInterface(string(out))
}

// parseWiFiInterface extracts the Wi-Fi device name from networksetup output.
func parseWiFiInterface(output string) (string, error) {
	scanner := bufio.NewScanner(strings.NewReader(output))
	foundWiFi := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "Wi-Fi") {
			foundWiFi = true
			continue
		}
		if foundWiFi && strings.HasPrefix(line, "Device:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "Device:")), nil
		}
	}
	return "", fmt.Errorf("no Wi-Fi interface found")
}

// listSSIDs returns the preferred wireless network names for the given interface.
func listSSIDs(iface string) ([]string, error) {
	out, err := exec.Command("networksetup", "-listpreferredwirelessnetworks", iface).Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list preferred networks: %w", err)
	}
	return parseSSIDs(string(out)), nil
}

// parseSSIDs extracts SSID names from networksetup preferred networks output.
func parseSSIDs(output string) []string {
	var ssids []string
	scanner := bufio.NewScanner(strings.NewReader(output))
	first := true
	for scanner.Scan() {
		if first {
			first = false // skip header line
			continue
		}
		ssid := strings.TrimSpace(scanner.Text())
		if ssid != "" {
			ssids = append(ssids, ssid)
		}
	}
	return ssids
}

// removeSSID removes a single SSID from the preferred network list.
func removeSSID(iface, ssid string) error {
	out, err := exec.Command("networksetup", "-removepreferredwirelessnetwork", iface, ssid).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove %q: %s", ssid, strings.TrimSpace(string(out)))
	}
	return nil
}
