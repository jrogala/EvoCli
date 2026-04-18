//go:build probe

package main

import (
	"bufio"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestProbeAllGameModes tries every GameModeType enum value with a
// minimal BuildSeasonDefinitionRequest and records what the server
// does with it. Run with: go test -tags=probe -v -run ProbeAllGameModes.
//
// This is separate from the normal test suite because it intentionally
// sends configs that are expected to fail — the goal is to harvest the
// server's error messages so we know which fields each mode needs.
func TestProbeAllGameModes(t *testing.T) {
	if _, err := os.Stat(filepath.Join(*serverFolder, "AssettoCorsaEVOServer.exe")); err != nil {
		t.Skipf("server binary not found at %s", *serverFolder)
	}

	for _, mode := range []string{
		"practice", "race_weekend", "instant_race",
		"hotlap", "hotstint", "cruise", "drift", "rally",
		"test_drive", "a_to_b", "superpole", "sro_race", "none",
	} {
		t.Run(mode, func(t *testing.T) {
			cfg, _ := baseConfig(t)

			// Minimal season — just the mode + an event. Every mode needs
			// an event to resolve a track, so we always attach one.
			gt, err := resolveEnum("game_type", mode, gameTypeMap)
			if err != nil {
				t.Fatalf("resolveEnum: %v", err)
			}
			event, err := resolveEvent(*serverFolder, "practice", "Monza", "GP")
			if err != nil {
				t.Fatalf("resolveEvent: %v", err)
			}
			season := defaultPracticeSeason()
			season.GameType = gt
			season.Event = event

			result := probeServer(t, cfg, &season)
			t.Logf("MODE %s → %s", mode, result)
		})
	}
}

// probeServer launches the server with the given config, captures the
// first ~2 seconds of log output, and returns a one-line verdict.
func probeServer(t *testing.T, cfg *ServerConfig, season *SeasonDefinition) string {
	t.Helper()
	if err := writeJSON(filepath.Join(*serverFolder, "server_config.json"), cfg); err != nil {
		return "write server config: " + err.Error()
	}
	if err := writeJSON(filepath.Join(*serverFolder, "season.json"), season); err != nil {
		return "write season: " + err.Error()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, filepath.Join(*serverFolder, "AssettoCorsaEVOServer.exe"),
		"-configjson", "server_config.json",
		"-seasonjson", "season.json",
	)
	cmd.Dir = *serverFolder

	stdout, _ := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout
	if err := cmd.Start(); err != nil {
		return "start: " + err.Error()
	}
	defer func() { _ = cmd.Process.Kill(); _ = cmd.Wait() }()

	var verdict string
	done := make(chan struct{})
	go func() {
		defer close(done)
		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
		for scanner.Scan() {
			line := scanner.Text()
			switch {
			case strings.Contains(line, "Start Server "+cfg.ServerName):
				verdict = "OK (started)"
				io.Copy(io.Discard, stdout)
				return
			case strings.Contains(line, "[error]") && strings.Contains(line, "required data is missing"):
				verdict = "required data missing"
				io.Copy(io.Discard, stdout)
				return
			case strings.Contains(line, "[error]") && strings.Contains(line, "Invalid season"):
				if verdict == "" {
					verdict = "Invalid season definition"
				}
			case strings.Contains(line, "Cannot find field"):
				verdict = "Cannot find field: " + line
				io.Copy(io.Discard, stdout)
				return
			case strings.Contains(line, "[error]") && verdict == "":
				verdict = line
			}
		}
	}()

	select {
	case <-done:
	case <-ctx.Done():
	}
	if verdict == "" {
		return "no verdict (timeout)"
	}
	return verdict
}
