# hostfilemanager

A small Go CLI tool that reads a `hosts.json` file and creates/updates a
**managed section** in the system hosts file (`/etc/hosts` on Linux/macOS).

## What it does

1. Loads `hosts.json` from the current directory (or a path you pass as an argument).
2. Reads `/etc/hosts`.
3. Replaces (or appends) the section between its marker comments with the
   entries from `hosts.json`.
4. Writes the file back, leaving everything outside the markers untouched.

## Managed section

The tool writes a block like this into `/etc/hosts`:

```
# BEGIN hostfilemanager managed entries
# Managed by hostfilemanager - entries between these markers are managed automatically
127.0.0.1   example.local www.example.local  # local development
192.168.1.50  nas.local storage.local        # home NAS
# END hostfilemanager managed entries
```

The marker comments serve two purposes:

* **For a human:** it's obvious which records are managed by an external app,
  so you don't accidentally edit or delete them by hand.
* **For the app:** it knows exactly where to insert or update the records it
  owns, and it never touches anything outside the markers.

## hosts.json format

```json
[
  {
    "ip": "127.0.0.1",
    "hostnames": ["example.local", "www.example.local"],
    "comment": "local development"
  },
  {
    "ip": "192.168.1.50",
    "hostnames": ["nas.local", "storage.local"],
    "comment": "home NAS"
  }
]
```

| Field | Required | Description |
|-------|----------|-------------|
| `ip` | yes | The IP address to map the hostnames to. |
| `hostnames` | yes | One or more hostnames to map to that IP. |
| `comment` | no | A human-readable note appended to the line. |

An example file is included as `hosts.json.example`. Copy it to `hosts.json`
and edit it for your environment.

## Build & run

```bash
# Build
go build -o hostfilemanager .

# Copy and edit the example config
cp hosts.json.example hosts.json
# edit hosts.json ...

# Run (requires root because /etc/hosts is root-owned)
sudo ./hostfilemanager

# Or with a custom config path
sudo ./hostfilemanager /path/to/hosts.json
```

The tool prints a summary on success:

```
hosts file updated: 2 managed entry(ies) written to /etc/hosts
```

## Requirements

* Go 1.27+
* Linux or macOS (`/etc/hosts`)
* Root / `sudo` privileges to write `/etc/hosts`

## Safety

Only the section between the `# BEGIN hostfilemanager managed entries` and
`# END hostfilemanager managed entries` markers is modified. Any manual entries
you've added to `/etc/hosts` outside those markers are preserved.
