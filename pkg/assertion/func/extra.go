package assertion

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

func getFloat(num interface{}) (float64, error) {
	switch val := num.(type) {
	case string:
		f64, _ := strconv.ParseFloat(val, 64)
		return f64, nil
	case float64:
		return float64(val), nil
	case float32:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case int32:
		return float64(val), nil
	case int16:
		return float64(val), nil
	case int8:
		return float64(val), nil
	case uint64:
		return float64(val), nil
	case uint32:
		return float64(val), nil
	case uint16:
		return float64(val), nil
	case uint8:
		return float64(val), nil
	case int:
		return float64(val), nil
	case uint:
		return float64(val), nil
	case time.Duration:
		return float64(val), nil
	default:
		return 0, fmt.Errorf("cannot convert %q to float64", num)
	}
}

// returns the float value of any real number, or error if it is not a numerical type
func getFloat0(num interface{}) (float64, error) {
	numValue := reflect.ValueOf(num)
	numKind := numValue.Kind()

	if numKind == reflect.Int ||
		numKind == reflect.Int8 ||
		numKind == reflect.Int16 ||
		numKind == reflect.Int32 ||
		numKind == reflect.Int64 {
		return float64(numValue.Int()), nil
	} else if numKind == reflect.Uint ||
		numKind == reflect.Uint8 ||
		numKind == reflect.Uint16 ||
		numKind == reflect.Uint32 ||
		numKind == reflect.Uint64 {
		return float64(numValue.Uint()), nil
	} else if numKind == reflect.Float32 ||
		numKind == reflect.Float64 {
		return numValue.Float(), nil
	} else {
		return 0.0, errors.New("must be a numerical type, but was: " + numKind.String())
	}
}
