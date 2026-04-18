package main

// ServerConfig maps to the server's ServerConfiguration protobuf message
// (serialized as JSON and passed via -configjson to the dedicated server).
type ServerConfig struct {
	ServerTCPListenerPort int      `json:"server_tcp_listener_port"`
	ServerUDPListenerPort int      `json:"server_udp_listener_port"`
	ServerTCPInternalPort int      `json:"server_tcp_internal_port"`
	ServerUDPInternalPort int      `json:"server_udp_internal_port"`
	ServerHTTPPort        int      `json:"server_http_port"`
	ServerName            string   `json:"server_name"`
	MaxPlayers            int      `json:"max_players"`
	Cycle                 bool     `json:"cycle"`
	AllowedCarsListFull   []AllowedCar `json:"allowed_cars_list_full"`
	DriverPassword        string   `json:"driver_password"`
	SpectatorPassword     string   `json:"spectator_password"`
	AdminPassword         string   `json:"admin_password"`
	Type                  string   `json:"type"`
	EntryListPath         string   `json:"entry_list_path"`
	ResultsPath           string   `json:"results_path"`

	NetcodeUpdateInterval int      `json:"netcode_update_interval,omitempty"`
	EntryListServerURL    string   `json:"entry_list_server_url,omitempty"`
	ResultsPostURL        string   `json:"results_post_url,omitempty"`
	Token                 string   `json:"token,omitempty"`
	TuningAllowed         bool     `json:"tuning_allowed,omitempty"`
	PIMin                 float32  `json:"pi_min,omitempty"`
	PIMax                 float32  `json:"pi_max,omitempty"`
	Property1             []bool   `json:"property_1,omitempty"`
	Property2             []bool   `json:"property_2,omitempty"`
	Property3             []bool   `json:"property_3,omitempty"`
}

// AllowedCar is an entry in ServerConfiguration.allowed_cars_list_full.
// An empty list means "no cars allowed" — the server browser seems to
// hide such servers, so ship at least one entry.
type AllowedCar struct {
	CarName    string  `json:"car_name"`
	Ballast    int     `json:"ballast"`
	Restrictor float32 `json:"restrictor"`
}

// EventItem picks the track/layout/event for a session. Matches the
// entries in events_practice.json / events_race_weekend.json.
type EventItem struct {
	Track       string `json:"track"`
	Layout      string `json:"layout"`
	EventName   string `json:"event_name"`
	TrackLength int    `json:"track_length"`
	MaxPitSlot  int    `json:"max_pit_slot,omitempty"`
}

// TimeOfDay matches GameModeSelectionTimeOfDay.
type TimeOfDay struct {
	Year           int     `json:"year"`
	Month          int     `json:"month"`
	Day            int     `json:"day"`
	Hour           int     `json:"hour"`
	Minute         int     `json:"minute"`
	Second         int     `json:"second"`
	TimeMultiplier float32 `json:"time_multiplier"`
}

// GameConfig maps to SimpleGameConfig. All 19 proto fields are exposed;
// zero values are suppressed from the JSON so individual modes only emit
// what they actually use.
type GameConfig struct {
	PracticeDuration                   int        `json:"practice_duration,omitempty"`
	PracticeTimeOfDay                  *TimeOfDay `json:"practice_time_of_day,omitempty"`
	PracticeOvertimeWaitingNextSession int        `json:"practice_overtime_waiting_next_session,omitempty"`
	PracticeMaxWaitToBox               int        `json:"practice_max_wait_to_box,omitempty"`

	QualifyDuration                   int        `json:"qualify_duration,omitempty"`
	QualifyTimeOfDay                  *TimeOfDay `json:"qualify_time_of_day,omitempty"`
	QualifyOvertimeWaitingNextSession int        `json:"qualify_overtime_waiting_next_session,omitempty"`
	QualifyMaxWaitToBox               int        `json:"qualify_max_wait_to_box,omitempty"`

	WarmupDuration                   int        `json:"warmup_duration,omitempty"`
	WarmupTimeOfDay                  *TimeOfDay `json:"warmup_time_of_day,omitempty"`
	WarmupOvertimeWaitingNextSession int        `json:"warmup_overtime_waiting_next_session,omitempty"`
	WarmupMaxWaitToBox               int        `json:"warmup_max_wait_to_box,omitempty"`

	RaceDuration                   int        `json:"race_duration,omitempty"`
	RaceDurationType               string     `json:"race_duration_type,omitempty"`
	RaceTimeOfDay                  *TimeOfDay `json:"race_time_of_day,omitempty"`
	RaceOvertimeWaitingNextSession int        `json:"race_overtime_waiting_next_session,omitempty"`
	RaceMaxWaitToBox               int        `json:"race_max_wait_to_box,omitempty"`

	MinWaitingForPlayers int  `json:"min_waiting_for_players,omitempty"`
	MaxWaitingForPlayers int  `json:"max_waiting_for_players,omitempty"`
	// note the spelling: the proto field is "enable_custom_penalities"
	EnableCustomPenalities  bool `json:"enable_custom_penalities,omitempty"`
	CarCutTyresOut          int  `json:"car_cut_tyres_out,omitempty"`
	WarningTriggerCountdown int  `json:"warning_trigger_countdown,omitempty"`
	TimePenaltyMs           int  `json:"time_penalty_ms,omitempty"`
}

// SeasonDefinition maps to BuildSeasonDefinitionRequest.
type SeasonDefinition struct {
	GameType         string     `json:"game_type"`
	Event            EventItem  `json:"event"`
	ExportJSON       bool       `json:"export_json"`
	GameConfig       GameConfig `json:"game_config"`
	WeatherType      string     `json:"weather_type"`
	WeatherBehaviour string     `json:"weather_behaviour"`
	InitialGrip      string     `json:"initial_grip"`
}

func defaultServerConfig() ServerConfig {
	return ServerConfig{
		ServerTCPListenerPort: 9700,
		ServerUDPListenerPort: 9700,
		ServerTCPInternalPort: 9700,
		ServerUDPInternalPort: 9700,
		ServerHTTPPort:        8080,
		ServerName:            "openevo server",
		MaxPlayers:            8,
		Cycle:                 true,
		AllowedCarsListFull: []AllowedCar{
			// Abarth 695 — safe default that's always present in cars.json.
			{CarName: "preset_695b_mech_1"},
		},
		Type:                  "MultiplayerServerListSessionType_RANKED",
	}
}

// sessionTimeOfDay returns a fresh TimeOfDay matching the launcher's
// defaults. Each call returns a new pointer — the server sometimes
// trips on shared instances across phases.
func sessionTimeOfDay() *TimeOfDay {
	return &TimeOfDay{Year: 2024, Month: 8, Day: 15, Hour: 16, TimeMultiplier: 1}
}

// baseSeason is the skeleton shared by every mode — clear static weather
// on a green track.
func baseSeason(gameType string) SeasonDefinition {
	return SeasonDefinition{
		GameType:         gameType,
		WeatherType:      "GameModeSelectionWeatherType_CLEAR",
		WeatherBehaviour: "GameModeSelectionWeatherBehaviour_STATIC",
		InitialGrip:      "InitialGrip_GREEN",
	}
}

// singleSessionGameConfig returns the minimal SimpleGameConfig the
// server accepts for single-session modes (practice-like).
func singleSessionGameConfig(duration int) GameConfig {
	return GameConfig{
		PracticeDuration:                   duration,
		PracticeTimeOfDay:                  sessionTimeOfDay(),
		PracticeOvertimeWaitingNextSession: 10,
		PracticeMaxWaitToBox:               10,
	}
}

// raceOnlyGameConfig returns the minimal SimpleGameConfig for modes that
// only use a race phase (instant_race).
func raceOnlyGameConfig(duration int, durationType string) GameConfig {
	return GameConfig{
		RaceDuration:                   duration,
		RaceDurationType:               durationType,
		RaceTimeOfDay:                  sessionTimeOfDay(),
		RaceOvertimeWaitingNextSession: 10,
		RaceMaxWaitToBox:               10,
	}
}

func defaultPracticeSeason() SeasonDefinition {
	s := baseSeason("GameModeType_PRACTICE")
	s.GameConfig = singleSessionGameConfig(1800)
	return s
}

func defaultRaceWeekendSeason() SeasonDefinition {
	s := baseSeason("GameModeType_RACE_WEEKEND")
	s.GameConfig = GameConfig{
		PracticeDuration:                   600,
		PracticeTimeOfDay:                  sessionTimeOfDay(),
		PracticeOvertimeWaitingNextSession: 10,
		PracticeMaxWaitToBox:               10,

		QualifyDuration:                   600,
		QualifyTimeOfDay:                  sessionTimeOfDay(),
		QualifyOvertimeWaitingNextSession: 10,
		QualifyMaxWaitToBox:               10,

		WarmupDuration:                   60,
		WarmupTimeOfDay:                  sessionTimeOfDay(),
		WarmupOvertimeWaitingNextSession: 10,
		WarmupMaxWaitToBox:               10,

		RaceDuration:                   1200,
		RaceDurationType:               "GameModeSelectionDuration_TIME",
		RaceTimeOfDay:                  sessionTimeOfDay(),
		RaceOvertimeWaitingNextSession: 10,
		RaceMaxWaitToBox:               10,
	}
	return s
}

func defaultInstantRaceSeason() SeasonDefinition {
	s := baseSeason("GameModeType_INSTANT_RACE")
	s.GameConfig = raceOnlyGameConfig(1200, "GameModeSelectionDuration_TIME")
	return s
}

// defaultForGameType returns the default SeasonDefinition for a mode.
// Modes that need game-mode-specific JSON templates the open-source
// build doesn't have yet (drift, superpole, sro_race) are omitted.
func defaultForGameType(shortMode string) (SeasonDefinition, bool) {
	switch shortMode {
	case "practice":
		return defaultPracticeSeason(), true
	case "race_weekend":
		return defaultRaceWeekendSeason(), true
	case "instant_race":
		return defaultInstantRaceSeason(), true
	case "hotlap":
		s := baseSeason("GameModeType_HOTLAP")
		s.GameConfig = singleSessionGameConfig(1800)
		return s, true
	case "hotstint":
		s := baseSeason("GameModeType_HOTSTINT")
		s.GameConfig = singleSessionGameConfig(1800)
		return s, true
	case "cruise":
		s := baseSeason("GameModeType_CRUISE")
		s.GameConfig = singleSessionGameConfig(1800)
		return s, true
	case "rally":
		s := baseSeason("GameModeType_RALLY")
		s.GameConfig = singleSessionGameConfig(1800)
		return s, true
	case "test_drive":
		s := baseSeason("GameModeType_TEST_DRIVE")
		s.GameConfig = singleSessionGameConfig(1800)
		return s, true
	case "a_to_b":
		s := baseSeason("GameModeType_A_TO_B")
		s.GameConfig = singleSessionGameConfig(1800)
		return s, true
	}
	return SeasonDefinition{}, false
}
