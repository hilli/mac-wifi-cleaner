# mac-wifi-cleaner

Clean up the long list of preferred Wi-Fi networks on your Mac.

macOS remembers every Wi-Fi network you've ever joined, which is a security risk — your Mac broadcasts these names when scanning. This tool makes it easy to prune that list.

## Install

With Homebrew:

```sh
brew install hilli/tap/mac-wifi-cleaner
```

With Go:

```sh
go install github.com/hilli/mac-wifi-cleaner@latest
```

Or build locally:

```sh
git clone https://github.com/hilli/mac-wifi-cleaner.git
cd mac-wifi-cleaner
go build -o mac-wifi-cleaner .
```

## Usage

### List all preferred networks

```sh
mac-wifi-cleaner list              # print to stdout
mac-wifi-cleaner list -o wifi.txt  # save to file
```

### Remove specific networks

```sh
# Remove every SSID in the file
mac-wifi-cleaner delete -f unwanted.txt

# Preview first with --dry-run
mac-wifi-cleaner delete -f unwanted.txt --dry-run
```

### Keep only specific networks

```sh
# Remove every SSID NOT in the file
mac-wifi-cleaner keep -f keepers.txt

# Preview first
mac-wifi-cleaner keep -f keepers.txt --dry-run
```

### Interactive mode (recommended)

```sh
mac-wifi-cleaner auto
```

This will:

1. Fetch your full SSID list
2. Open it in `$EDITOR` (defaults to `vim`)
3. You delete the lines for networks you **don't** want to keep
4. Save and quit — it shows what will be removed and asks for confirmation
5. Removes the unwanted networks

Add `--dry-run` to preview without making changes.

Example:

```shell
$ mac-wifi-cleaner auto
Opening 123 SSIDs in zed. Remove the networks you want to DELETE, save and quit.

Keeping 117 networks, removing 6:
  - BreezeGardenHill
  - RNet_2
  - Radisson_Guest
  - Skodsborg Guest
  - dlink-2FD4
  - HomeBox-C5B0_2.4G

Continue? [y/N] y
  ✓ Removed BreezeGardenHill
  ✓ Removed RNet_2
  ✓ Removed Radisson_Guest
  ✓ Removed Skodsborg Guest
  ✓ Removed dlink-2FD4
  ✓ Removed HomeBox-C5B0_2.4G

Done. Removed 6 network(s).
```

Verify:

```shell
$ mac-wifi-cleaner list | wc -l
     117
```

## Notes

- Automatically detects your Wi-Fi interface (no need to know if it's `en0`, `en1`, etc.)
- Lines starting with `#` are ignored in SSID files (useful for comments)
- No external dependencies — pure Go stdlib
