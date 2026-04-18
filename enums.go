package main

import (
	"fmt"
	"strings"
)

// Enum value maps. Users can pass either the short form ("clear") or the
// full proto name ("GameModeSelectionWeatherType_CLEAR"). Validation
// rejects anything else so typos don't silently fall through to the
// server, where the error surface is far noisier.

var weatherTypeMap = map[string]string{
	"clear":             "GameModeSelectionWeatherType_CLEAR",
	"scattered_clouds":  "GameModeSelectionWeatherType_SCATTERED_CLOUDS",
	"broken_clouds":     "GameModeSelectionWeatherType_BROKEN_CLOUDS",
	"overcast":          "GameModeSelectionWeatherType_OVERCAST",
	"damp":              "GameModeSelectionWeatherType_DAMP",
	"drizzle":           "GameModeSelectionWeatherType_DRIZZLE",
	"rain":              "GameModeSelectionWeatherType_RAIN",
	"heavy_rain":        "GameModeSelectionWeatherType_HEAVY_RAIN",
	"custom":            "GameModeSelectionWeatherType_CUSTOM",
}

var weatherBehaviourMap = map[string]string{
	"static":  "GameModeSelectionWeatherBehaviour_STATIC",
	"dynamic": "GameModeSelectionWeatherBehaviour_DYNAMIC",
}

var initialGripMap = map[string]string{
	"green":   "InitialGrip_GREEN",
	"fast":    "InitialGrip_FAST",
	"optimum": "InitialGrip_OPTIMUM",
}

var gameTypeMap = map[string]string{
	"practice":      "GameModeType_PRACTICE",
	"race_weekend":  "GameModeType_RACE_WEEKEND",
	"instant_race":  "GameModeType_INSTANT_RACE",
	"hotlap":        "GameModeType_HOTLAP",
	"hotstint":      "GameModeType_HOTSTINT",
	"cruise":        "GameModeType_CRUISE",
	"drift":         "GameModeType_DRIFT",
	"rally":         "GameModeType_RALLY",
	"test_drive":    "GameModeType_TEST_DRIVE",
	"a_to_b":        "GameModeType_A_TO_B",
	"superpole":     "GameModeType_SUPERPOLE",
	"sro_race":      "GameModeType_SRO_RACE",
	"none":          "GameModeType_NONE",
}

var sessionTypeMap = map[string]string{
	"ranked":   "MultiplayerServerListSessionType_RANKED",
	"unranked": "MultiplayerServerListSessionType_UNRANKED",
	"both":     "MultiplayerServerListSessionType_BOTH",
}

var raceDurationTypeMap = map[string]string{
	"none": "GameModeSelectionDuration_NONE",
	"time": "GameModeSelectionDuration_TIME",
	"laps": "GameModeSelectionDuration_LAPS",
}

// resolveEnum maps a user-supplied value (short or full form) to the
// proto enum name. Passing an empty string returns an empty string so
// callers can treat "no flag provided" and "use default" identically.
func resolveEnum(field, value string, m map[string]string) (string, error) {
	if value == "" {
		return "", nil
	}
	if v, ok := m[strings.ToLower(value)]; ok {
		return v, nil
	}
	for _, full := range m {
		if value == full {
			return full, nil
		}
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return "", fmt.Errorf("invalid %s %q (valid: %s)", field, value, strings.Join(keys, ", "))
}
