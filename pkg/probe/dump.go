package probe

import (
	"encoding/json"
	"fmt"
)

func ToMap(proberes interface{}) (map[string]interface{}, error) {

	var probeFinalRes map[string]interface{}
	inrec, errMarsh := json.Marshal(proberes)
	if errMarsh != nil {
		return nil, fmt.Errorf("Error while ProbeRes convertion to map: %s", errMarsh)
	}
	errUnmarsh := json.Unmarshal(inrec, &probeFinalRes)
	if errUnmarsh != nil {
		return nil, fmt.Errorf("Error while ProbeRes convertion to map: %s", errUnmarsh)
	}

	return probeFinalRes, nil

}
