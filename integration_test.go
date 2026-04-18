package main

import (
	"bufio"
	"context"
	"flag"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// serverFolder must point at a real dedicated server install with
// AssettoCorsaEVOServer.exe + content.kspkg. Defaults to ../server.
var serverFolder = flag.String("server", "../server", "path to dedicated server install")

// portCounter hands out a unique port per subtest so back-to-back launches
// don't collide while the OS tears down the previous socket.
var portCounter atomic.Int32

func init() {
	portCounter.Store(9720)
}

// launchAndWait writes the two configs, starts the server, and scans its
// stdout for a success marker ("Start Server ...") or a failure line.
// It kills the process as soon as it has a verdict so the test loop can
// move on.
func launchAndWait(t *testing.T, cfg *ServerConfig, season *SeasonDefinition) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(*serverFolder, "AssettoCorsaEVOServer.exe")); err != nil {
		t.Skipf("server binary not found at %s (-server flag)", *serverFolder)
	}

	if err := writeJSON(filepath.Join(*serverFolder, "server_config.json"), cfg); err != nil {
		t.Fatalf("write server_config: %v", err)
	}
	if err := writeJSON(filepath.Join(*serverFolder, "season.json"), season); err != nil {
		t.Fatalf("write season: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, filepath.Join(*serverFolder, "AssettoCorsaEVOServer.exe"),
		"-configjson", "server_config.json",
		"-seasonjson", "season.json",
	)
	cmd.Dir = *serverFolder

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	done := make(chan string, 1)
	go func() {
		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
		for scanner.Scan() {
			line := scanner.Text()
			switch {
			case strings.Contains(line, "Start Server "+cfg.ServerName):
				done <- "ok"
				io.Copy(io.Discard, stdout)
				return
			case strings.Contains(line, "Invalid season definition"),
				strings.Contains(line, "Failed to read season definition"),
				strings.Contains(line, "Failed to parse buildServer"),
				strings.Contains(line, "required data is missing"),
				strings.Contains(line, "Cannot find field"),
				strings.Contains(line, "invalid value"):
				done <- line
				io.Copy(io.Discard, stdout)
				return
			}
		}
		done <- "eof"
	}()

	select {
	case result := <-done:
		if result != "ok" {
			t.Fatalf("server rejected config: %s", result)
		}
	case <-ctx.Done():
		t.Fatalf("timeout waiting for 'Start Server'")
	}
}

// mustEnum panics the test if the enum value doesn't resolve — keeps the
// table-driven loops readable.
func mustEnum(t *testing.T, field, value string, m map[string]string) string {
	t.Helper()
	v, err := resolveEnum(field, value, m)
	if err != nil {
		t.Fatalf("%v", err)
	}
	return v
}

// baseConfig returns a fresh ServerConfig + SeasonDefinition on unique
// ports, resolved against the events catalogue. All subtests share this
// starting point and tweak one field.
func baseConfig(t *testing.T) (*ServerConfig, *SeasonDefinition) {
	t.Helper()
	port := int(portCounter.Add(1))
	cfg := defaultServerConfig()
	cfg.ServerName = "evocli-it-" + t.Name()
	cfg.ServerTCPListenerPort = port
	cfg.ServerUDPListenerPort = port
	cfg.ServerTCPInternalPort = port
	cfg.ServerUDPInternalPort = port
	cfg.ServerHTTPPort = port + 100

	event, err := resolveEvent(*serverFolder, "practice", "Monza", "GP")
	if err != nil {
		t.Fatalf("resolveEvent: %v", err)
	}
	season := defaultPracticeSeason()
	season.Event = event
	return &cfg, &season
}

func TestWeatherType(t *testing.T) {
	for _, wt := range []string{
		"clear", "scattered_clouds", "broken_clouds", "overcast",
		"damp", "drizzle", "rain", "heavy_rain", "custom",
	} {
		t.Run(wt, func(t *testing.T) {
			cfg, season := baseConfig(t)
			season.WeatherType = mustEnum(t, "weather_type", wt, weatherTypeMap)
			launchAndWait(t, cfg, season)
		})
	}
}

func TestWeatherBehaviour(t *testing.T) {
	for _, b := range []string{"static", "dynamic"} {
		t.Run(b, func(t *testing.T) {
			cfg, season := baseConfig(t)
			season.WeatherBehaviour = mustEnum(t, "weather_behaviour", b, weatherBehaviourMap)
			launchAndWait(t, cfg, season)
		})
	}
}

func TestInitialGrip(t *testing.T) {
	for _, g := range []string{"green", "fast", "optimum"} {
		t.Run(g, func(t *testing.T) {
			cfg, season := baseConfig(t)
			season.InitialGrip = mustEnum(t, "initial_grip", g, initialGripMap)
			launchAndWait(t, cfg, season)
		})
	}
}

func TestSessionType(t *testing.T) {
	for _, st := range []string{"ranked", "unranked", "both"} {
		t.Run(st, func(t *testing.T) {
			cfg, season := baseConfig(t)
			cfg.Type = mustEnum(t, "type", st, sessionTypeMap)
			launchAndWait(t, cfg, season)
		})
	}
}

func TestGameTypePractice(t *testing.T) {
	cfg, season := baseConfig(t)
	season.GameType = mustEnum(t, "game_type", "practice", gameTypeMap)
	launchAndWait(t, cfg, season)
}

// baseRaceWeekendConfig mirrors baseConfig but returns the race-weekend
// defaults — useful when a subtest wants to vary only one race field.
func baseRaceWeekendConfig(t *testing.T) (*ServerConfig, *SeasonDefinition) {
	t.Helper()
	cfg, _ := baseConfig(t)
	event, err := resolveEvent(*serverFolder, "race_weekend", "Monza", "GP")
	if err != nil {
		t.Fatalf("resolveEvent: %v", err)
	}
	season := defaultRaceWeekendSeason()
	season.Event = event
	return cfg, &season
}

func TestGameTypeRaceWeekend(t *testing.T) {
	cfg, season := baseRaceWeekendConfig(t)
	launchAndWait(t, cfg, season)
}

func TestRaceDurationType(t *testing.T) {
	for _, d := range []string{"none", "time", "laps"} {
		t.Run(d, func(t *testing.T) {
			cfg, season := baseRaceWeekendConfig(t)
			season.GameConfig.RaceDurationType = mustEnum(t, "race_duration_type", d, raceDurationTypeMap)
			launchAndWait(t, cfg, season)
		})
	}
}

// TestAllSupportedGameTypes launches every wired mode with its default
// template. Modes that need content we don't have (drift, superpole,
// sro_race) are deliberately not in defaultForGameType and so are
// skipped here.
func TestAllSupportedGameTypes(t *testing.T) {
	for _, mode := range []string{
		"practice", "race_weekend", "instant_race",
		"hotlap", "hotstint", "cruise", "rally", "test_drive", "a_to_b",
	} {
		t.Run(mode, func(t *testing.T) {
			cfg, _ := baseConfig(t)
			s, ok := defaultForGameType(mode)
			if !ok {
				t.Fatalf("no default for %s", mode)
			}
			event, err := resolveEvent(*serverFolder, catalogueForMode(mode), "Monza", "GP")
			if err != nil {
				t.Fatalf("resolveEvent: %v", err)
			}
			s.Event = event
			launchAndWait(t, cfg, &s)
		})
	}
}

// TestExtraGameConfigFields exercises the small-value fields on
// SimpleGameConfig that don't have their own enum. We just set them and
// make sure the server still accepts the JSON.
func TestExtraGameConfigFields(t *testing.T) {
	cfg, season := baseConfig(t)
	season.GameConfig.MinWaitingForPlayers = 1
	season.GameConfig.MaxWaitingForPlayers = 8
	season.GameConfig.CarCutTyresOut = 3
	season.GameConfig.TimePenaltyMs = 5000
	season.GameConfig.WarningTriggerCountdown = 10
	season.GameConfig.EnableCustomPenalities = true
	launchAndWait(t, cfg, season)
}

// TestExtraServerConfigFields exercises the ServerConfig fields added
// for full coverage.
func TestExtraServerConfigFields(t *testing.T) {
	cfg, season := baseConfig(t)
	cfg.NetcodeUpdateInterval = 55
	cfg.Token = "test-token"
	cfg.TuningAllowed = true
	cfg.PIMin = 1.0
	cfg.PIMax = 99.0
	cfg.EntryListServerURL = "http://example.invalid/entry"
	cfg.ResultsPostURL = "http://example.invalid/results"
	launchAndWait(t, cfg, season)
}

func TestUnsupportedGameTypeRejected(t *testing.T) {
	for _, mode := range []string{"drift", "superpole", "sro_race"} {
		if _, ok := defaultForGameType(mode); ok {
			t.Errorf("expected %s to be unsupported", mode)
		}
	}
}

func TestInvalidEnumRejected(t *testing.T) {
	if _, err := resolveEnum("weather_type", "sunny", weatherTypeMap); err == nil {
		t.Fatal("expected error for invalid weather type")
	}
	if _, err := resolveEnum("initial_grip", "slippery", initialGripMap); err == nil {
		t.Fatal("expected error for invalid grip")
	}
}
