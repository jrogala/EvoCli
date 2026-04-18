package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

const usage = `evocli — launch an Assetto Corsa EVO dedicated server from the command line.

Usage:
  evocli -folder <server-path> [flags]

Folder:
  -folder                path to the dedicated server install (required)

Pre-made config files (pass-through, bypass the flag builders below):
  -serverconfigjson      path to a ready ServerConfiguration JSON
  -seasonjson            path to a ready BuildSeasonDefinitionRequest JSON

Server (ServerConfiguration) flags:
  -server_name                 display name
  -max_players                 max players (default 8)
  -server_tcp_listener_port    TCP listener port (default 9700)
  -server_udp_listener_port    UDP listener port (default 9700)
  -server_tcp_internal_port    TCP internal port (default = listener)
  -server_udp_internal_port    UDP internal port (default = listener)
  -server_http_port            HTTP port (default 8080)
  -admin_password              admin password
  -driver_password             driver password
  -spectator_password          spectator password
  -cycle                       cycle sessions (default true)
  -type                        session type: ranked | unranked | both (default ranked)
  -netcode_update_interval     netcode tick interval (default 55)
  -cars                        comma-separated car_name list (e.g. preset_695b_mech_1,preset_m2_mech_1)
                               or 'all' to allow every entry in cars.json (default: Abarth 695 only)
  -entry_list_path             path to entry list JSON
  -results_path                path to results JSON
  -entry_list_server_url       URL to fetch entry list from
  -results_post_url            URL to POST results to
  -token                       backend token
  -tuning_allowed              allow setup tuning
  -pi_min / -pi_max            performance index filter

Session (BuildSeasonDefinitionRequest) flags:
  -game_type                   practice | race_weekend | instant_race | hotlap |
                               hotstint | cruise | rally | test_drive | a_to_b (default practice)
  -track                       track name (default Monza)
  -layout                      layout name (default GP)
  -weather_type                clear | rain | overcast | drizzle | damp | heavy_rain |
                               scattered_clouds | broken_clouds | custom (default clear)
  -weather_behaviour           static | dynamic (default static)
  -initial_grip                green | fast | optimum (default green)
  -practice_duration           practice phase duration in seconds
  -qualify_duration            qualifying phase duration in seconds
  -warmup_duration             warmup phase duration in seconds
  -race_duration               race duration — seconds if race_duration_type=time, laps if =laps
  -race_duration_type          none | time | laps
  -min_waiting_for_players     minimum players to start
  -max_waiting_for_players     maximum wait-lobby size
  -car_cut_tyres_out           tyres-outside-track threshold for a cut
  -time_penalty_ms             added time penalty (ms)
  -warning_trigger_countdown   warning countdown before penalty
  -enable_custom_penalities    enable custom penalty rules (note: spelling matches proto)

Note: drift, superpole, sro_race are not supported yet — they need
game-mode-specific JSON templates not shipped with the server binary.

Examples:
  evocli -folder ./server -track Monza -layout GP
  evocli -folder ./server -game_type race_weekend -race_duration 1800 -race_duration_type time
  evocli -folder ./server -serverconfigjson my_sc.json -seasonjson my_season.json
`

type cliFlags struct {
	folder           string
	serverConfigJSON string
	seasonJSON       string
	configSuffix     string
	logName          string

	// server
	serverName   string
	maxPlayers   int
	tcpPort      int
	udpPort      int
	tcpInternal  int
	udpInternal  int
	httpPort     int
	adminPwd     string
	driverPwd    string
	specPwd      string
	cycle        bool
	sessType     string
	netcode      int
	entryPath    string
	resultsPath  string
	entryURL     string
	resultsURL   string
	token        string
	tuningAllow  bool
	piMin        float64
	piMax        float64
	carsCSV      string

	// season
	gameType      string
	track         string
	layout        string
	weather       string
	weatherBehave string
	grip          string
	practiceDur   int
	qualifyDur    int
	warmupDur     int
	raceDur       int
	raceDurType   string
	minWait       int
	maxWait       int
	carCut        int
	timePenalty   int
	warnCountdown int
	custPenalties bool

	// explicit setters (so we only overwrite fields the user actually touched)
	setPracticeDur   bool
	setQualifyDur    bool
	setWarmupDur     bool
	setRaceDur       bool
	setRaceDurType   bool
}

func parseFlags() *cliFlags {
	f := &cliFlags{}
	flag.StringVar(&f.folder, "folder", "", "")
	flag.StringVar(&f.serverConfigJSON, "serverconfigjson", "", "")
	flag.StringVar(&f.seasonJSON, "seasonjson", "", "")
	flag.StringVar(&f.configSuffix, "config_suffix", "", "")
	flag.StringVar(&f.logName, "log_name", "", "")

	flag.StringVar(&f.serverName, "server_name", "openevo server", "")
	flag.IntVar(&f.maxPlayers, "max_players", 8, "")
	flag.IntVar(&f.tcpPort, "server_tcp_listener_port", 9700, "")
	flag.IntVar(&f.udpPort, "server_udp_listener_port", 9700, "")
	flag.IntVar(&f.tcpInternal, "server_tcp_internal_port", 0, "")
	flag.IntVar(&f.udpInternal, "server_udp_internal_port", 0, "")
	flag.IntVar(&f.httpPort, "server_http_port", 8080, "")
	flag.StringVar(&f.adminPwd, "admin_password", "", "")
	flag.StringVar(&f.driverPwd, "driver_password", "", "")
	flag.StringVar(&f.specPwd, "spectator_password", "", "")
	flag.BoolVar(&f.cycle, "cycle", true, "")
	flag.StringVar(&f.sessType, "type", "ranked", "")
	flag.IntVar(&f.netcode, "netcode_update_interval", 0, "")
	flag.StringVar(&f.entryPath, "entry_list_path", "", "")
	flag.StringVar(&f.resultsPath, "results_path", "", "")
	flag.StringVar(&f.entryURL, "entry_list_server_url", "", "")
	flag.StringVar(&f.resultsURL, "results_post_url", "", "")
	flag.StringVar(&f.token, "token", "", "")
	flag.BoolVar(&f.tuningAllow, "tuning_allowed", false, "")
	flag.Float64Var(&f.piMin, "pi_min", 0, "")
	flag.Float64Var(&f.piMax, "pi_max", 0, "")
	flag.StringVar(&f.carsCSV, "cars", "", "comma-separated car_name list OR 'all' to allow every car in cars.json; default: Abarth 695 only")

	flag.StringVar(&f.gameType, "game_type", "practice", "")
	flag.StringVar(&f.track, "track", "Monza", "")
	flag.StringVar(&f.layout, "layout", "GP", "")
	flag.StringVar(&f.weather, "weather_type", "clear", "")
	flag.StringVar(&f.weatherBehave, "weather_behaviour", "static", "")
	flag.StringVar(&f.grip, "initial_grip", "green", "")
	flag.IntVar(&f.practiceDur, "practice_duration", 0, "")
	flag.IntVar(&f.qualifyDur, "qualify_duration", 0, "")
	flag.IntVar(&f.warmupDur, "warmup_duration", 0, "")
	flag.IntVar(&f.raceDur, "race_duration", 0, "")
	flag.StringVar(&f.raceDurType, "race_duration_type", "", "")
	flag.IntVar(&f.minWait, "min_waiting_for_players", 0, "")
	flag.IntVar(&f.maxWait, "max_waiting_for_players", 0, "")
	flag.IntVar(&f.carCut, "car_cut_tyres_out", 0, "")
	flag.IntVar(&f.timePenalty, "time_penalty_ms", 0, "")
	flag.IntVar(&f.warnCountdown, "warning_trigger_countdown", 0, "")
	flag.BoolVar(&f.custPenalties, "enable_custom_penalities", false, "")

	flag.Usage = func() { fmt.Fprint(os.Stderr, usage) }
	flag.Parse()

	// Track which duration flags the user actually provided so mode
	// defaults aren't blown away with zeroes.
	flag.Visit(func(fl *flag.Flag) {
		switch fl.Name {
		case "practice_duration":
			f.setPracticeDur = true
		case "qualify_duration":
			f.setQualifyDur = true
		case "warmup_duration":
			f.setWarmupDur = true
		case "race_duration":
			f.setRaceDur = true
		case "race_duration_type":
			f.setRaceDurType = true
		}
	})
	return f
}

func main() {
	f := parseFlags()

	if f.folder == "" {
		fmt.Fprintln(os.Stderr, "evocli: -folder is required")
		flag.Usage()
		os.Exit(2)
	}

	if f.serverConfigJSON != "" && f.seasonJSON != "" {
		if err := runServer(f.folder, nil, nil, f.serverConfigJSON, f.seasonJSON, f.configSuffix, f.logName); err != nil {
			fail(err)
		}
		return
	}

	cfg, season, err := buildConfigs(f)
	if err != nil {
		fail(err)
	}
	if err := runServer(f.folder, cfg, season, f.serverConfigJSON, f.seasonJSON, f.configSuffix, f.logName); err != nil {
		fail(err)
	}
}

// buildConfigs translates parsed flags into a ServerConfig and
// SeasonDefinition. If either JSON pass-through flag was set the
// corresponding return value is nil (the file is used as-is).
func buildConfigs(f *cliFlags) (*ServerConfig, *SeasonDefinition, error) {
	var cfg *ServerConfig
	if f.serverConfigJSON == "" {
		sessionType, err := resolveEnum("type", f.sessType, sessionTypeMap)
		if err != nil {
			return nil, nil, err
		}
		c := defaultServerConfig()
		c.ServerName = f.serverName
		c.MaxPlayers = f.maxPlayers
		c.ServerTCPListenerPort = f.tcpPort
		c.ServerUDPListenerPort = f.udpPort
		c.ServerTCPInternalPort = f.tcpInternal
		if c.ServerTCPInternalPort == 0 {
			c.ServerTCPInternalPort = f.tcpPort
		}
		c.ServerUDPInternalPort = f.udpInternal
		if c.ServerUDPInternalPort == 0 {
			c.ServerUDPInternalPort = f.udpPort
		}
		c.ServerHTTPPort = f.httpPort
		c.AdminPassword = f.adminPwd
		c.DriverPassword = f.driverPwd
		c.SpectatorPassword = f.specPwd
		c.Cycle = f.cycle
		c.Type = sessionType
		c.NetcodeUpdateInterval = f.netcode
		c.EntryListPath = f.entryPath
		c.ResultsPath = f.resultsPath
		c.EntryListServerURL = f.entryURL
		c.ResultsPostURL = f.resultsURL
		c.Token = f.token
		c.TuningAllowed = f.tuningAllow
		c.PIMin = float32(f.piMin)
		c.PIMax = float32(f.piMax)
		if f.carsCSV != "" {
			if strings.EqualFold(f.carsCSV, "all") {
				cars, err := loadAllCars(f.folder)
				if err != nil {
					return nil, nil, fmt.Errorf("resolve -cars all: %w", err)
				}
				c.AllowedCarsListFull = cars
			} else {
				var cars []AllowedCar
				for _, name := range strings.Split(f.carsCSV, ",") {
					name = strings.TrimSpace(name)
					if name == "" {
						continue
					}
					cars = append(cars, AllowedCar{CarName: name})
				}
				if len(cars) > 0 {
					c.AllowedCarsListFull = cars
				}
			}
		}
		cfg = &c
	}

	var season *SeasonDefinition
	if f.seasonJSON == "" {
		gt, err := resolveEnum("game_type", f.gameType, gameTypeMap)
		if err != nil {
			return nil, nil, err
		}
		wt, err := resolveEnum("weather_type", f.weather, weatherTypeMap)
		if err != nil {
			return nil, nil, err
		}
		wb, err := resolveEnum("weather_behaviour", f.weatherBehave, weatherBehaviourMap)
		if err != nil {
			return nil, nil, err
		}
		ig, err := resolveEnum("initial_grip", f.grip, initialGripMap)
		if err != nil {
			return nil, nil, err
		}

		s, ok := defaultForGameType(f.gameType)
		if !ok {
			return nil, nil, fmt.Errorf("game_type %q is not supported (drift, superpole, sro_race need game-mode-specific templates). Supported: practice, race_weekend, instant_race, hotlap, hotstint, cruise, rally, test_drive, a_to_b", f.gameType)
		}
		s.GameType = gt
		s.WeatherType = wt
		s.WeatherBehaviour = wb
		s.InitialGrip = ig

		catalogue := catalogueForMode(f.gameType)
		event, err := resolveEvent(f.folder, catalogue, f.track, f.layout)
		if err != nil {
			return nil, nil, err
		}
		// The official launcher omits max_pit_slot from the event it
		// sends — it's catalogue metadata, not runtime config. Drop it
		// so our output byte-matches theirs.
		event.MaxPitSlot = 0
		s.Event = event

		if f.setPracticeDur {
			s.GameConfig.PracticeDuration = f.practiceDur
		}
		if f.setQualifyDur {
			s.GameConfig.QualifyDuration = f.qualifyDur
		}
		if f.setWarmupDur {
			s.GameConfig.WarmupDuration = f.warmupDur
		}
		if f.setRaceDur {
			s.GameConfig.RaceDuration = f.raceDur
		}
		if f.setRaceDurType {
			rdt, err := resolveEnum("race_duration_type", f.raceDurType, raceDurationTypeMap)
			if err != nil {
				return nil, nil, err
			}
			s.GameConfig.RaceDurationType = rdt
		}
		if f.minWait != 0 {
			s.GameConfig.MinWaitingForPlayers = f.minWait
		}
		if f.maxWait != 0 {
			s.GameConfig.MaxWaitingForPlayers = f.maxWait
		}
		if f.carCut != 0 {
			s.GameConfig.CarCutTyresOut = f.carCut
		}
		if f.timePenalty != 0 {
			s.GameConfig.TimePenaltyMs = f.timePenalty
		}
		if f.warnCountdown != 0 {
			s.GameConfig.WarningTriggerCountdown = f.warnCountdown
		}
		if f.custPenalties {
			s.GameConfig.EnableCustomPenalities = true
		}
		season = &s
	}
	return cfg, season, nil
}

// catalogueForMode picks the events_*.json the given mode looks up in.
// race-style modes read from events_race_weekend.json; everything else
// shares events_practice.json.
func catalogueForMode(mode string) string {
	switch mode {
	case "race_weekend", "instant_race":
		return "race_weekend"
	}
	return "practice"
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "evocli: %v\n", err)
	os.Exit(1)
}

// used by probes / tests elsewhere — keeps the import list honest.
var _ = strings.Split
