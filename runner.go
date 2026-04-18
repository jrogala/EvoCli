package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
)

func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// copyFile copies src to dst. Used when the user passes -serverconfigjson
// or -seasonjson so the file ends up in the server's working directory
// under the name the server expects.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

// runServer writes (or copies) the two config files into the server
// folder and execs AssettoCorsaEVOServer.exe. stdout/stderr are
// inherited so logs appear live.
//
// configSuffix, when non-empty, is inserted into the written JSON
// filenames (server_config_<suffix>.json / season_<suffix>.json) so
// multiple evocli instances can run against the same server folder
// without overwriting each other's files.
func runServer(folder string, cfg *ServerConfig, season *SeasonDefinition, serverConfigJSON, seasonJSON, configSuffix, logName string) error {
	configFile := "server_config.json"
	seasonFile := "season.json"
	if configSuffix != "" {
		configFile = "server_config_" + configSuffix + ".json"
		seasonFile = "season_" + configSuffix + ".json"
	}
	configPath := filepath.Join(folder, configFile)
	seasonPath := filepath.Join(folder, seasonFile)

	if serverConfigJSON != "" {
		if err := copyFile(serverConfigJSON, configPath); err != nil {
			return fmt.Errorf("copy server config: %w", err)
		}
	} else {
		if err := writeJSON(configPath, cfg); err != nil {
			return fmt.Errorf("write server config: %w", err)
		}
	}
	if seasonJSON != "" {
		if err := copyFile(seasonJSON, seasonPath); err != nil {
			return fmt.Errorf("copy season: %w", err)
		}
	} else {
		if err := writeJSON(seasonPath, season); err != nil {
			return fmt.Errorf("write season: %w", err)
		}
	}

	binary := filepath.Join(folder, "AssettoCorsaEVOServer.exe")
	if _, err := os.Stat(binary); err != nil {
		return fmt.Errorf("server binary not found at %s", binary)
	}
	// exec.Cmd needs an absolute path when Dir is set to a different
	// directory — otherwise Windows looks up the relative path after
	// chdir and fails with "path not found".
	absBinary, err := filepath.Abs(binary)
	if err != nil {
		return fmt.Errorf("resolve binary path: %w", err)
	}
	binary = absBinary

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	args := []string{
		"-configjson", configFile,
		"-seasonjson", seasonFile,
	}
	if logName != "" {
		// -name controls the server's own log filename under
		// serverConfig/<name>.txt. Useful for fleet deployments so
		// every instance writes to its own file.
		args = append(args, "-name", logName)
	}
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Dir = folder
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Fprintf(os.Stderr, "[evocli] launching %s\n", binary)
	if cfg != nil {
		fmt.Fprintf(os.Stderr, "[evocli] server_name=%q, tcp=%d, http=%d\n",
			cfg.ServerName, cfg.ServerTCPListenerPort, cfg.ServerHTTPPort)
	}
	if season != nil {
		fmt.Fprintf(os.Stderr, "[evocli] %s @ %s, mode=%s\n",
			season.Event.Track, season.Event.Layout, season.GameType)
	}
	return cmd.Run()
}
