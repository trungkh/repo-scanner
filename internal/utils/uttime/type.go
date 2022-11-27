package uttime

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"repo-scanner/internal/utils/serror"
)

type Time time.Time

func (ox Time) String() string {
	return Format(time.RFC3339, time.Time(ox))
}

func (ox Time) MarshalJSON() ([]byte, error) {
	val := fmt.Sprintf("\"%s\"", Format(time.RFC3339, time.Time(ox)))
	return []byte(val), nil
}

func (ox *Time) UnmarshalJSON(data []byte) error {
	var val interface{}
	err := json.Unmarshal(data, &val)
	if err != nil {
		return serror.NewFromErrorc(err, "Failed to unmarshal json")
	}

	valx := reflect.ValueOf(val)
	switch valx.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val = int64(valx.Int() / 1000)

	case reflect.Float32, reflect.Float64:
		val = int64(valx.Float() / 1000)
	}

	tim, errx := Parse(val)
	if errx != nil {
		errx.AddComments("while parsing")
		return errx
	}

	*ox = Time(tim)
	return nil
}

func ToTime(tim time.Time) Time {
	return Time(tim)
}

func ToTimep(tim *time.Time) *Time {
	if tim == nil {
		return nil
	}

	timx := ToTime(*tim)
	return &timx
}

type Date Time

func (ox Date) String() string {
	return Format(DefaultDateFormat, time.Time(ox))
}

func (ox Date) MarshalJSON() ([]byte, error) {
	val := fmt.Sprintf("\"%s\"", Format(DefaultDateFormat, time.Time(ox)))
	return []byte(val), nil
}

func (ox *Date) UnmarshalJSON(data []byte) error {
	var val Time
	err := json.Unmarshal(data, &val)
	if err != nil {
		return serror.NewFromErrorc(err, "Failed to unmarshal json")
	}

	*ox = Date(val)
	return nil
}

func ToDate(tim time.Time) Date {
	return Date(tim)
}

func ToDatep(tim *time.Time) *Date {
	if tim == nil {
		return nil
	}

	datx := ToDate(*tim)
	return &datx
}
