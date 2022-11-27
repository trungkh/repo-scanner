package sqlq

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"repo-scanner/internal/utils/utinterface"
	"repo-scanner/internal/utils/uttime"
)

type NullableTime string

func (NullableTime) Cast(val interface{}) (res interface{}, err error) {
	if val == nil {
		return NullableTime(""), err
	}

	var tim time.Time
	tim, err = uttime.Parse(val)
	if err != nil {
		return res, err
	}

	return NullableTime(tim.Format(time.RFC3339Nano)), err
}

func (ox NullableTime) Value() *time.Time {
	if ox != "" {
		if ptr, err := time.Parse(time.RFC3339Nano, string(ox)); err == nil {
			return &ptr
		}
	}
	return nil
}

func (ox NullableTime) Valuex() *uttime.Time {
	if ox != "" {
		if ptr, err := time.Parse(time.RFC3339Nano, string(ox)); err == nil {
			tim := uttime.Time(ptr)
			return &tim
		}
	}
	return nil
}

func (ox NullableTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(ox.Valuex())
}

type NullableString string

func (NullableString) Cast(val interface{}) (res interface{}, err error) {
	if val == nil {
		return NullableString(""), err
	}

	return NullableString(fmt.Sprintf("%v", val)), err
}

func (ox NullableString) Value() string {
	return string(ox)
}

type NullableInt64 int64

func (NullableInt64) Cast(val interface{}) (res interface{}, err error) {
	if val == nil {
		return NullableInt64(0), err
	}

	return NullableInt64(utinterface.ToInt(val, 0)), err
}

func (ox NullableInt64) Value() int64 {
	return int64(ox)
}

type NullableFloat64 float64

func (NullableFloat64) Cast(val interface{}) (res interface{}, err error) {
	if val == nil {
		return NullableFloat64(0), err
	}

	return NullableFloat64(utinterface.ToFloat(val, 0)), err
}

func (ox NullableFloat64) Value() float64 {
	return float64(ox)
}

type NullableMetas map[string]interface{}

func (NullableMetas) Cast(val interface{}) (res interface{}, err error) {
	res = make(NullableMetas)

	if val == nil {
		return res, err
	}

	switch valx := val.(type) {
	case map[string]interface{}:
		res = NullableMetas(valx)

	case string:
		err = json.Unmarshal([]byte(valx), &res)
		if err != nil {
			return res, err
		}

	case []byte:
		err = json.Unmarshal(valx, &res)
		if err != nil {
			return res, err
		}

	default:
		err = errors.New("Cannot parse to metas")
		return res, err
	}

	return res, err
}

func (ox NullableMetas) Value() map[string]interface{} {
	return map[string]interface{}(ox)
}

type NullableBool string

func (NullableBool) Cast(val interface{}) (res interface{}, err error) {
	if val == nil {
		return NullableBool(""), err
	}

	switch valx := val.(type) {
	case bool:
		val = utinterface.ToString(valx)

	case string:
		val = ""
		switch strings.ToLower(valx) {
		case "true", "1":
			val = "true"

		case "false", "0":
			val = "false"
		}

	case int, int32, int64:
		val = ""
		switch utinterface.ToInt(valx, -1) {
		case 1:
			val = "true"

		case 0:
			val = "false"
		}

	case byte:
		val = ""
		switch valx {
		case 1:
			val = "true"

		case 0:
			val = "false"
		}

	default:
		val = ""
	}

	return NullableBool(fmt.Sprint(val)), err
}

func (ox NullableBool) IsNull() bool {
	return ox == ""
}

func (ox NullableBool) Value(def bool) bool {
	if ox.IsNull() {
		return def
	}

	return utinterface.ToBool(ox, def)
}

func (ox NullableBool) MarshalJSON() ([]byte, error) {
	var val interface{}
	if !ox.IsNull() {
		val = ox.Value(false)
	}

	return json.Marshal(val)
}
