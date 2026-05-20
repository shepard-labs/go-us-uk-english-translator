package dict

import (
	"encoding/json"
	"fmt"
	"os"

	_ "embed"
)

//go:embed uk_spellings.json
var defaultDictJSON []byte

type Direction string

const (
	DirectionAmerican Direction = "american"
	DirectionBritish  Direction = "british"
)

// LoadDictionary loads the default dictionary and merges it with an optional user dictionary.
// It returns a mapping of source words to target words based on the specified direction.
func LoadDictionary(direction Direction, userDictPath string) (map[string]string, error) {
	ukToUS := make(map[string]string)

	if err := json.Unmarshal(defaultDictJSON, &ukToUS); err != nil {
		return nil, fmt.Errorf("failed to parse built-in dictionary: %w", err)
	}

	if userDictPath != "" {
		userData, err := os.ReadFile(userDictPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read user dictionary at %s: %w", userDictPath, err)
		}

		userDict := make(map[string]string)
		if err := json.Unmarshal(userData, &userDict); err != nil {
			return nil, fmt.Errorf("failed to parse user dictionary: %w", err)
		}

		for k, v := range userDict {
			ukToUS[k] = v
		}
	}

	result := make(map[string]string)

	if direction == DirectionAmerican {
		// Target is American, meaning UK -> US
		// The dictionary is already UK -> US, so just copy it
		for uk, us := range ukToUS {
			result[uk] = us
		}
	} else if direction == DirectionBritish {
		// Target is British, meaning US -> UK
		for uk, us := range ukToUS {
			result[us] = uk
		}
	} else {
		return nil, fmt.Errorf("unknown direction: %s", direction)
	}

	return result, nil
}
