package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "list":
		cmdList()
	case "delete":
		cmdDelete()
	case "keep":
		cmdKeep()
	case "auto":
		cmdAuto()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: mac-wifi-cleaner <command> [flags]

Commands:
  list               List all preferred Wi-Fi networks (one per line)
    -o <file>        Write to file instead of stdout

  delete -f <file>   Remove every SSID listed in the file
    --dry-run        Show what would be removed without removing

  keep -f <file>     Remove every SSID NOT listed in the file
    --dry-run        Show what would be removed without removing

  auto               Interactive: list → edit → confirm → clean
`)
}

// cmdList prints or saves the list of preferred SSIDs.
func cmdList() {
	iface, ssids := mustGetSSIDs()

	outFile := flagValue(os.Args[2:], "-o")
	if outFile != "" {
		if err := writeSSIDFile(outFile, ssids); err != nil {
			fatal("writing file: %v", err)
		}
		fmt.Printf("Wrote %d SSIDs to %s (interface %s)\n", len(ssids), outFile, iface)
		return
	}

	for _, s := range ssids {
		fmt.Println(s)
	}
}

// cmdDelete removes every SSID present in the given file.
func cmdDelete() {
	args := os.Args[2:]
	file := flagValue(args, "-f")
	if file == "" {
		fatal("delete requires -f <file>")
	}
	dryRun := hasFlag(args, "--dry-run")

	iface, currentSSIDs := mustGetSSIDs()
	toDelete, err := readSSIDFile(file)
	if err != nil {
		fatal("reading file: %v", err)
	}

	deleteSet := toSet(toDelete)
	var removing []string
	for _, s := range currentSSIDs {
		if deleteSet[s] {
			removing = append(removing, s)
		}
	}

	executeRemovals(iface, removing, dryRun)
}

// cmdKeep removes every SSID NOT present in the given file.
func cmdKeep() {
	args := os.Args[2:]
	file := flagValue(args, "-f")
	if file == "" {
		fatal("keep requires -f <file>")
	}
	dryRun := hasFlag(args, "--dry-run")

	iface, currentSSIDs := mustGetSSIDs()
	toKeep, err := readSSIDFile(file)
	if err != nil {
		fatal("reading file: %v", err)
	}

	keepSet := toSet(toKeep)
	var removing []string
	for _, s := range currentSSIDs {
		if !keepSet[s] {
			removing = append(removing, s)
		}
	}

	executeRemovals(iface, removing, dryRun)
}

// cmdAuto runs the interactive flow: list → edit → confirm → clean.
func cmdAuto() {
	dryRun := hasFlag(os.Args[2:], "--dry-run")

	iface, ssids := mustGetSSIDs()
	if len(ssids) == 0 {
		fmt.Println("No preferred Wi-Fi networks found.")
		return
	}

	tmpFile, err := os.CreateTemp("", "wifi-ssids-*.txt")
	if err != nil {
		fatal("creating temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if err := writeSSIDFile(tmpPath, ssids); err != nil {
		fatal("writing temp file: %v", err)
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	fmt.Printf("Opening %d SSIDs in %s. Remove the networks you want to DELETE, save and quit.\n", len(ssids), editor)
	cmd := exec.Command(editor, tmpPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fatal("editor exited with error: %v", err)
	}

	kept, err := readSSIDFile(tmpPath)
	if err != nil {
		fatal("reading edited file: %v", err)
	}
	keepSet := toSet(kept)

	var removing []string
	for _, s := range ssids {
		if !keepSet[s] {
			removing = append(removing, s)
		}
	}

	if len(removing) == 0 {
		fmt.Println("No networks to remove.")
		return
	}

	fmt.Printf("\nKeeping %d networks, removing %d:\n", len(kept), len(removing))
	for _, s := range removing {
		fmt.Printf("  - %s\n", s)
	}
	fmt.Print("\nContinue? [y/N] ")

	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer != "y" && answer != "yes" {
		fmt.Println("Aborted.")
		return
	}

	executeRemovals(iface, removing, dryRun)
}

// executeRemovals removes the given SSIDs, or prints what would be removed in dry-run mode.
func executeRemovals(iface string, removing []string, dryRun bool) {
	if len(removing) == 0 {
		fmt.Println("Nothing to remove.")
		return
	}

	if dryRun {
		fmt.Printf("Dry run — would remove %d network(s):\n", len(removing))
		for _, s := range removing {
			fmt.Printf("  - %s\n", s)
		}
		return
	}

	var failed []string
	for _, s := range removing {
		if err := removeSSID(iface, s); err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ %s: %v\n", s, err)
			failed = append(failed, s)
		} else {
			fmt.Printf("  ✓ Removed %s\n", s)
		}
	}

	if len(failed) > 0 {
		fmt.Fprintf(os.Stderr, "\n%d network(s) failed to remove.\n", len(failed))
		os.Exit(1)
	}
	fmt.Printf("\nDone. Removed %d network(s).\n", len(removing))
}

// --- helpers ---

func mustGetSSIDs() (string, []string) {
	iface, err := detectWiFiInterface()
	if err != nil {
		fatal("detecting Wi-Fi interface: %v", err)
	}
	ssids, err := listSSIDs(iface)
	if err != nil {
		fatal("listing SSIDs: %v", err)
	}
	return iface, ssids
}

func readSSIDFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var ssids []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			ssids = append(ssids, line)
		}
	}
	return ssids, scanner.Err()
}

func writeSSIDFile(path string, ssids []string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for _, s := range ssids {
		fmt.Fprintln(w, s)
	}
	return w.Flush()
}

func toSet(items []string) map[string]bool {
	m := make(map[string]bool, len(items))
	for _, s := range items {
		m[s] = true
	}
	return m
}

func flagValue(args []string, name string) string {
	for i, a := range args {
		if a == name && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

func hasFlag(args []string, name string) bool {
	for _, a := range args {
		if a == name {
			return true
		}
	}
	return false
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}
