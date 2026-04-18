# EvoCli

[![ci](https://github.com/jrogala/EvoCli/actions/workflows/ci.yml/badge.svg)](https://github.com/jrogala/EvoCli/actions/workflows/ci.yml)
[![Latest release](https://img.shields.io/github/v/release/jrogala/EvoCli?label=latest&color=blue)](https://github.com/jrogala/EvoCli/releases/latest)
[![VirusTotal](https://img.shields.io/badge/VirusTotal-latest%20scan-5c5c5c?logo=virustotal)](https://github.com/jrogala/EvoCli/releases/latest)

A command-line launcher for the Assetto Corsa EVO dedicated server.

Does the same job as the official `ServerLauncher.exe` GUI, but as a scriptable CLI so you can run the server from a service, CI, or a shell loop.

## Download

Grab the latest `evocli.exe` from the **[Releases page](https://github.com/jrogala/EvoCli/releases/latest)**. Each release's notes contain the VirusTotal scan link for the binary attached to that release.

Direct link to the most recent build:

    https://github.com/jrogala/EvoCli/releases/latest/download/evocli.exe

Or from PowerShell:

```powershell
Invoke-WebRequest -Uri https://github.com/jrogala/EvoCli/releases/latest/download/evocli.exe -OutFile evocli.exe
```

## Quick start

1. **Install the dedicated server** from Steam: "Assetto Corsa EVO Dedicated Server". Note the install folder — typically `C:\Program Files (x86)\Steam\steamapps\common\Assetto Corsa EVO Dedicated Server`.
2. **Download `evocli.exe`** (Windows amd64) — see the section above.
3. **Launch a practice server on Monza GP**:
   ```
   evocli -folder "C:\Program Files (x86)\Steam\steamapps\common\Assetto Corsa EVO Dedicated Server" -track "Monza" -layout "GP" -server_name "My Server"
   ```
4. **Stop the server** with `Ctrl+C`.

That's it. The CLI writes `server_config.json` and `season.json` into the server folder and runs `AssettoCorsaEVOServer.exe -configjson ... -seasonjson ...` for you.

## Common recipes

### Race weekend on Spa, 30-minute race

```
evocli -folder <server> -game_type race_weekend -track "Circuit de Spa Francorchamps" -layout "GP" \
  -practice_duration 600 -qualify_duration 900 -warmup_duration 60 \
  -race_duration 1800 -race_duration_type time
```

### Rain practice at Nurburgring Nordschleife, 90 minutes

```
evocli -folder <server> -track "Nurburgring" -layout "Nordschleife" \
  -practice_duration 5400 -weather_type rain -weather_behaviour dynamic
```

### Allow every car in the catalogue

```
evocli -folder <server> -track "Monza" -layout "GP" -cars all
```

### Allow a specific set of cars

```
evocli -folder <server> -track "Monza" -layout "GP" \
  -cars preset_695b_mech_1,preset_m2_mech_1,preset_rs6_mech_1
```

See `cars.json` in the server folder for the full list of `car_name` values.

### Use your own pre-made JSON configs

If you already have hand-crafted `ServerConfiguration` and `BuildSeasonDefinitionRequest` JSON files, skip the builders:

```
evocli -folder <server> -serverconfigjson my_sc.json -seasonjson my_season.json
```

## Supported game modes

The in-game server browser filters servers into four categories. Only the modes that map to one of those categories actually show up to real players:

| `-game_type` | Browser category | Status |
|---|---|---|
| `practice` | **Practice** | ✅ fully working |
| `hotlap` | **Trackday** | ✅ fully working |
| `race_weekend` | **Race** | ✅ fully working |
| `instant_race` | **Race** | ✅ fully working |
| `rally` | **A→B Challenge** | ✅ fully working (the session becomes "A to B Challenge" internally) |

Accepted by the server but **not listed in the in-game browser** (so only joinable by direct IP):

| `-game_type` | Notes |
|---|---|
| `a_to_b` | Falls back to TimeAttack specialization on closed circuits — prefer `rally` instead. |
| `test_drive` | Test-drive session. |

Not usable yet (server accepts the game_type but crashes during session bootstrap because it's missing a mode-specific template or spawn metadata):

- `hotstint` — `libprotobuf FATAL key not found: -1` on every circuit layout we tried.
- `cruise` — throws `No starting position!` on pit-box layouts.
- `drift`, `superpole`, `sro_race` — fail early with `Couldn't load json proto file` (the server expects a game-mode-specific JSON template we don't have a copy of).

## All flags

Run `evocli -h` for the full list. Highlights:

**Server (ServerConfiguration):**
- `-folder` — path to dedicated server install (required)
- `-server_name`, `-max_players`, `-cycle`, `-type` (ranked/unranked/both)
- `-server_tcp_listener_port`, `-server_udp_listener_port`, `-server_http_port` (defaults 9700/9700/8080)
- `-admin_password`, `-driver_password`, `-spectator_password`
- `-cars` — comma-separated `car_name` list or `all`
- `-pi_min`, `-pi_max`, `-tuning_allowed`
- `-entry_list_path`, `-results_path`, `-entry_list_server_url`, `-results_post_url`, `-token`

**Session (BuildSeasonDefinitionRequest):**
- `-game_type`, `-track`, `-layout`
- `-weather_type` (clear / scattered_clouds / broken_clouds / overcast / damp / drizzle / rain / heavy_rain / custom)
- `-weather_behaviour` (static / dynamic)
- `-initial_grip` (green / fast / optimum)
- `-practice_duration`, `-qualify_duration`, `-warmup_duration`, `-race_duration`
- `-race_duration_type` (none / time / laps)
- `-min_waiting_for_players`, `-max_waiting_for_players`, `-car_cut_tyres_out`, `-time_penalty_ms`, `-warning_trigger_countdown`, `-enable_custom_penalities`

## Troubleshooting

**"Could not bind TCP listener socket"** — another process is already on the port. Either kill it or pick a different port with `-server_tcp_listener_port` / `-server_udp_listener_port` / `-server_http_port`.

**Server registers but doesn't appear in the in-game browser** — check `allowed_cars_list_full` isn't empty. The CLI ships one car by default for this reason. Also make sure the TCP listener binds cleanly; the lobby hides unreachable servers.

**"Wrong car selected for the server"** — the player's car isn't in your allowed list. Add more via `-cars` or switch car in-game.

**"track not found in catalogue" / "layout not found"** — use exact track/layout names from `events_practice.json` (or `events_race_weekend.json` when `-game_type` is `race_weekend` / `instant_race`) in the server folder.

## Building from source

Requires Go 1.25+.

```
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o evocli.exe .
```

The `-trimpath -ldflags="-s -w"` combination strips symbols and local paths from the binary — same flags the release workflow uses — which also reduces the rate at which Defender's ML heuristics mis-classify Go binaries.

### Layout

| File | Purpose |
|---|---|
| `main.go` | CLI entry, flag parsing, dispatch |
| `config.go` | `ServerConfig` / `SeasonDefinition` / `GameConfig` structs + defaults (mirrors the server's protobuf schema) |
| `enums.go` | Short-form → full-name mappings for every enum the server accepts |
| `events.go` | Track/layout lookup from `events_*.json`; `-cars all` reader |
| `runner.go` | Writes the two JSON configs, execs `AssettoCorsaEVOServer.exe` |
| `integration_test.go` | Per-enum / per-mode end-to-end tests (requires a real server install) |
| `probe_test.go` | Harness for mapping server behaviour per game-mode (`-tags=probe`) |

### Testing

The integration tests actually launch the dedicated server per case and verify startup. Point them at your server install with `-server`:

```
go test -v -p 1 -server "C:\Program Files (x86)\Steam\steamapps\common\Assetto Corsa EVO Dedicated Server" ./...
```

Unit tests (no server required) run with a narrower filter:

```
go test -v -run "TestInvalidEnumRejected|TestUnsupportedGameTypeRejected" ./...
```

The mode-probe test is gated behind a build tag because it deliberately sends invalid configs to harvest server error messages:

```
go test -v -tags=probe -run ProbeAllGameModes ./...
```

### Releasing

Tag and push:

```
git tag v0.1.2
git push origin v0.1.2
```

`.github/workflows/release.yml` then:

1. Builds `evocli.exe` on a clean Ubuntu runner with the strip flags.
2. Computes the SHA-256 of the binary.
3. Uploads the binary to VirusTotal (requires `VIRUSTOTAL_API_KEY` secret).
4. Publishes a GitHub release on this repo with the binary, checksum, and VT scan link.

