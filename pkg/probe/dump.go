package probe

import (
	"encoding/json"
	"fmt"
	"strings"
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

	// Should be useless now
	//lcMap := lowercaseMap(probeFinalRes)
	//return lcMap, nil
}

func DumpV2(s interface{}) (map[string]interface{}, error) {

	var inInterface map[string]interface{}
	inrec, _ := json.Marshal(s)
	json.Unmarshal(inrec, &inInterface)

	return inInterface, nil

}

//lowerCaseMap returns a Map with all key (string) even nested as lowercase
func lowercaseMap(m map[string]interface{}) map[string]interface{} {

	lowerMap := make(map[string]interface{}, len(m))

	for k, v := range m {

		lowerMap[strings.ToLower(k)] = v

		nestedMap, isMap := v.(map[string]interface{})
		if isMap {
			lowerMap[fmt.Sprintf("%v", k)] = lowercaseMap(nestedMap)

		}

	}
	return lowerMap
}
