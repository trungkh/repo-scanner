package uttime

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/gearintellix/u2"

	"repo-scanner/internal/utils/serror"
	"repo-scanner/internal/utils/utint"
	"repo-scanner/internal/utils/utinterface"
	"repo-scanner/internal/utils/utstring"
)

const (
	// Year4Digits for Years in 4 digits
	Year4Digits = "2006"

	// Year2Digits for Years in 2 digits
	Year2Digits = "06"

	// Month2Digits for Months in 2 digits
	Month2Digits = "01"

	// Month1Digits for Months in 1 digits
	Month1Digits = "1"

	// Day2Digits for Days in 2 digits
	Day2Digits = "02"

	// Day1Digits for Days in 1 digits
	Day1Digits = "2"

	// Hour2Digits for Hours in 2 digits
	Hour2Digits = "15"

	// Minute2Digits for Minutes in 2 digits
	Minute2Digits = "04"

	// Second2Digits for Second in 2 digits
	Second2Digits = "05"

	// Timezone for Timezone Location
	Timezone = "MST"
)

// DateFormat date format
type DateFormat = string

const (
	// DefaultDateFormat for Default Date Format
	DefaultDateFormat DateFormat = "Y-m-d"

	// DefaultDateWithTimezoneFormat for Default Date Format with Timezone
	DefaultDateWithTimezoneFormat DateFormat = "Y-m-d TZ"

	// INDateFormat for Indonesian Date Format
	INDateFormat DateFormat = "d-m-Y"

	// DefaultDateTimeFormat for Default Date Time Format
	DefaultDateTimeFormat DateFormat = "Y-m-d H:i:s"

	// INDateTimeFormat for Indonesian Date Time Format
	INDateTimeFormat DateFormat = "d-m-Y H:i:s"

	// DefaultDateTimeWithTimezoneFormat for Default Date Time Format With Timezone
	DefaultDateTimeWithTimezoneFormat DateFormat = "Y-m-d H:i:s TZ"

	// DefaultTimeFormat for Default Time Format
	DefaultTimeFormat DateFormat = "H:i:s"
)

var (
	EmptyTime   time.Time
	EmptyTimeFN func() time.Time

	materials = []string{
		// standard format
		"!" + time.RFC3339Nano,
		"!" + time.RFC3339,

		// usual format
		DefaultDateTimeFormat,
		"Y-m-d",
		"YmdHis",
		"Ymd",
		"Y-m-dTH:i:s.999999999",
		"Y-m-dTH:i:s",

		// usual format with timezone
		DefaultDateTimeWithTimezoneFormat,
		"Y-m-d H:i:s Z07:00",
		"Y-m-dZ07:00",
		"YmdHisZ07:00",
		"YmdZ07:00",

		// any format
		"Y-m-d H:i",
		"Y-m-d H",
		"Y-m",
		"Y",
		"!" + time.RFC822Z,
		"!" + time.RFC822,
		"!" + time.RFC1123Z,
		"!" + time.RFC1123,
		"!" + time.RFC850,
	}
)

//> Helper

func getEmptyTime() time.Time {
	if EmptyTimeFN != nil {
		return EmptyTimeFN()
	}

	return EmptyTime
}

// Most for makesure return time without error
func Most(tim time.Time, errx serror.SError) time.Time {
	if errx != nil {
		errx.Panic()
	}

	return tim
}

// ParseToGoFormat to parsing format to go format
//
// Deprecated: use GoLayout instead
func ParseToGoFormat(format DateFormat) string {
	return GoLayout(format)
}

// GoLayout to get golang time layout
func GoLayout(format DateFormat) string {
	rl := map[string]string{
		"Y":  Year4Digits,
		"y":  Year2Digits,
		"m":  Month2Digits,
		"M":  Month1Digits,
		"d":  Day2Digits,
		"D":  Day1Digits,
		"H":  Hour2Digits,
		"i":  Minute2Digits,
		"s":  Second2Digits,
		"TZ": Timezone,
	}

	for k, v := range rl {
		format = strings.ReplaceAll(format, k, v)
	}

	return format
}

// GetTimezone to get timezone
func GetTimezone(zone string) (res *time.Location, errx serror.SError) {
	res = time.Local

	switch {
	case strings.HasPrefix(zone, "+"), strings.HasPrefix(zone, "-"):
		offset := utstring.Sub(zone, 1, 0)
		if !utint.IsInteger(offset) {
			errx = serror.Newf("Invalid timezone offset %s", offset)
			return res, errx
		}

		offx := int(utint.StringToInt(offset, 0) * 60 * 60)
		if string(zone[0]) == "-" {
			offx *= -1
		}

		res = time.FixedZone(fmt.Sprintf("UTC%s", zone), offx)

	case zone == "UTC", zone == "0":
		res = time.UTC

	case zone == "@":
		res = time.Local

	default:
		var err error
		res, err = time.LoadLocation(zone)
		if err != nil {
			errx = serror.NewFromErrorc(err, "Failed to load location")
			return res, errx
		}
	}

	return res, errx
}

// WithTimezone to replace timezone with recalculate timestamp value
func WithTimezone(tim time.Time, zone string) (res time.Time, errx serror.SError) {
	res = tim

	var loc *time.Location
	loc, errx = GetTimezone(zone)
	if errx != nil {
		errx.AddComments("while get timezone")
		return res, errx
	}

	res = res.In(loc)
	return res, errx
}

// ForceTimezone to replace timezone without impact timestamp value
func ForceTimezone(tim time.Time, zone string) (res time.Time, errx serror.SError) {
	res = tim

	var loc *time.Location
	loc, errx = GetTimezone(zone)
	if errx != nil {
		errx.AddComments("while get timezone")
		return res, errx
	}

	var (
		dfmt = GoLayout(DefaultDateTimeFormat)
		val  = tim.Format(dfmt)

		err error
	)

	res, err = time.ParseInLocation(dfmt, val, loc)
	if err != nil {
		errx = serror.NewFromErrorc(err, "Failed to parse in location")
		return res, errx
	}

	return res, errx
}

//> Getter

// Now for current time
func Now() time.Time {
	return time.Now()
}

// NowWithTimezone for current time with timezone
func NowWithTimezone(zone string) (res time.Time, errx serror.SError) {
	res = Now()
	res, errx = WithTimezone(res, zone)
	if errx != nil {
		errx.AddComments("while with timezone")
		return res, errx
	}

	return res, errx
}

// NowForceTimezone for current time with timezone
func NowForceTimezone(zone string) (res time.Time, errx serror.SError) {
	res = Now()
	res, errx = ForceTimezone(res, zone)
	if errx != nil {
		errx.AddComments("while force timezone")
		return res, errx
	}

	return res, errx
}

// MostNowWithTimezone for current time with timezone
func MostNowWithTimezone(zone string) time.Time {
	return Most(NowWithTimezone(zone))
}

// MostNowForceTimezone for current time force timezone
func MostNowForceTimezone(zone string) time.Time {
	return Most(NowForceTimezone(zone))
}

//> Composer and Parser

// Compose to compose time
func Compose(year int, month int, day int, hour int, minute int, second int) (time.Time, serror.SError) {
	vv := "__year__-__month__-__day__ __hour__:__minute__:__second__"
	vv = u2.Binding(vv, map[string]string{
		"year":   utstring.LeftPad(utstring.IntToString(year), 4, "0"),
		"month":  utstring.LeftPad(utstring.IntToString(month), 2, "0"),
		"day":    utstring.LeftPad(utstring.IntToString(day), 2, "0"),
		"hour":   utstring.LeftPad(utstring.IntToString(hour), 2, "0"),
		"minute": utstring.LeftPad(utstring.IntToString(minute), 2, "0"),
		"second": utstring.LeftPad(utstring.IntToString(second), 2, "0"),
	})

	res, errx := ParseWithFormat(DefaultDateTimeFormat, vv)
	if errx != nil {
		errx.AddComments("while parse with format")
		return res, errx
	}

	return res, nil
}

// ComposeUTC to compose time with default UTC zone
func ComposeUTC(year int, month int, day int, hour int, minute int, second int) (time.Time, serror.SError) {
	vv := "__year__-__month__-__day__ __hour__:__minute__:__second__"
	vv = u2.Binding(vv, map[string]string{
		"year":   utstring.LeftPad(utstring.IntToString(year), 4, "0"),
		"month":  utstring.LeftPad(utstring.IntToString(month), 2, "0"),
		"day":    utstring.LeftPad(utstring.IntToString(day), 2, "0"),
		"hour":   utstring.LeftPad(utstring.IntToString(hour), 2, "0"),
		"minute": utstring.LeftPad(utstring.IntToString(minute), 2, "0"),
		"second": utstring.LeftPad(utstring.IntToString(second), 2, "0"),
	})

	res, errx := ParseUTCWithFormat(DefaultDateTimeFormat, vv)
	if errx != nil {
		errx.AddComments("while parse utc with format")
		return res, errx
	}

	return res, nil
}

// ParseWithFormat to parse time with format
func ParseWithFormat(format DateFormat, value string) (res time.Time, errx serror.SError) {
	res = getEmptyTime()

	if len(format) <= 0 {
		format = "!" + time.RFC3339
	}

	switch {
	case strings.HasPrefix(format, "!"):
		format = utstring.Sub(format, 1, 0)

	default:
		format = GoLayout(format)
	}

	var err error
	res, err = time.ParseInLocation(format, value, time.Local)
	if err != nil {
		errx = serror.NewFromErrorc(err, "Failed to parse time")
		return res, errx
	}

	return res, errx
}

// ParseUTCWithFormat to parse time with format and default UTC zone
func ParseUTCWithFormat(format DateFormat, value string) (res time.Time, errx serror.SError) {
	res = getEmptyTime()

	if len(format) <= 0 {
		format = "!" + time.RFC3339
	}

	switch {
	case strings.HasPrefix(format, "!"):
		format = utstring.Sub(format, 1, 0)

	default:
		format = GoLayout(format)
	}

	var err error
	res, err = time.Parse(format, value)
	if err != nil {
		errx = serror.NewFromErrorc(err, "Failed to parse time")
		return res, errx
	}

	return res, errx
}

// ParseWithFormatAndTimezone to parse time with format and timezone
func ParseWithFormatAndTimezone(format DateFormat, value string, zone string) (res time.Time, errx serror.SError) {
	res = getEmptyTime()

	res, errx = ParseWithFormat(format, value)
	if errx != nil {
		errx.AddComments("while parse with format")
		return res, errx
	}

	res, errx = WithTimezone(res, zone)
	if errx != nil {
		errx.AddComments("while with timezone")
		return res, errx
	}

	return res, errx
}

// ParseUTCWithFormatAndTimezone to parse time with format and timezone, default UTC zone
func ParseUTCWithFormatAndTimezone(format DateFormat, value string, zone string) (res time.Time, errx serror.SError) {
	res = getEmptyTime()

	res, errx = ParseUTCWithFormat(format, value)
	if errx != nil {
		errx.AddComments("while parse utc with format")
		return res, errx
	}

	res, errx = WithTimezone(res, zone)
	if errx != nil {
		errx.AddComments("while with timezone")
		return res, errx
	}

	return res, errx
}

// ParseWithFormatAndForceTimezone to parse time and format and force timezone
func ParseWithFormatAndForceTimezone(format DateFormat, value string, zone string) (res time.Time, errx serror.SError) {
	res = getEmptyTime()

	res, errx = ParseWithFormat(format, value)
	if errx != nil {
		errx.AddComments("while parse with format")
		return res, errx
	}

	res, errx = ForceTimezone(res, zone)
	if errx != nil {
		errx.AddComments("while force timezone")
		return res, errx
	}

	return res, errx
}

// ParseFromInteger to parse time from integer
func ParseFromInteger(value int64) (res time.Time, errx serror.SError) {
	res = time.Unix(value, 0)
	return res, errx
}

// ParseUTCFromInteger to parse time from integer with default UTC zone
func ParseUTCFromInteger(value int64) (res time.Time, errx serror.SError) {
	res = time.Unix(value, 0)
	res, errx = ForceTimezone(res, "UTC")
	if errx != nil {
		errx.AddComments("while force timezone")
		return res, errx
	}

	return res, errx
}

// ParseFromIntegerWithTimezone to parse time from integer with timezone
func ParseFromIntegerWithTimezone(value int64, zone string) (res time.Time, errx serror.SError) {
	res, errx = ParseFromInteger(value)
	if errx != nil {
		errx.AddComments("while parse from integer")
		return res, errx
	}

	res, errx = WithTimezone(res, zone)
	if errx != nil {
		errx.AddComments("while with timezone")
		return res, errx
	}

	return res, errx
}

// ParseUTCFromIntegerWithTimezone to parse time from integer with timezone and default UTC zone
func ParseUTCFromIntegerWithTimezone(value int64, zone string) (res time.Time, errx serror.SError) {
	res, errx = ParseUTCFromInteger(value)
	if errx != nil {
		errx.AddComments("while parse utc from integer")
		return res, errx
	}

	res, errx = WithTimezone(res, zone)
	if errx != nil {
		errx.AddComments("while with timezone")
		return res, errx
	}

	return res, errx
}

// ParseFromIntegerForceTimezone to parse time from integer with timezone
func ParseFromIntegerForceTimezone(value int64, zone string) (res time.Time, errx serror.SError) {
	res, errx = ParseFromInteger(value)
	if errx != nil {
		errx.AddComments("while parse from integer")
		return res, errx
	}

	res, errx = ForceTimezone(res, zone)
	if errx != nil {
		errx.AddComments("while force timezone")
		return res, errx
	}

	return res, errx
}

//> Parse Helper

// Parse to parse time from anything
func Parse(value interface{}) (time.Time, serror.SError) {
	return ParseWithTimezone(value, "@")
}

// ParseUTC to parse time with default UTC zone from anything
func ParseUTC(value interface{}) (time.Time, serror.SError) {
	return ParseUTCWithTimezone(value, "@")
}

// MostParse for most parse time from anything
func MostParse(value interface{}) time.Time {
	return Most(Parse(value))
}

// MostParseUTC for most parse time with default UTC zone from anything
func MostParseUTC(value interface{}) time.Time {
	return Most(ParseUTC(value))
}

// ParseWithTimezone to parse time with timezone from anything
func ParseWithTimezone(value interface{}, zone string) (res time.Time, errx serror.SError) {
	res = getEmptyTime()

	if utinterface.IsNil(value) {
		errx = serror.New("Cannot parse from nil")
		return res, errx
	}

	val := reflect.ValueOf(value)
	switch val.Kind() {
	case reflect.String:
		res, errx = ParseFromStringWithTimezone(val.String(), zone)
		if errx != nil {
			errx.AddComments("while parse from string with timezone")
			return res, errx
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		res, errx = ParseFromIntegerWithTimezone(val.Int(), zone)
		if errx != nil {
			errx.AddComments("while parse from integer with timezone")
			return res, errx
		}

	case reflect.Float32, reflect.Float64:
		res, errx = ParseFromIntegerWithTimezone(int64(val.Float()), zone)
		if errx != nil {
			errx.AddComments("while parse from integer with timezone")
			return res, errx
		}

	default:
		switch val.Type().String() {
		case "time.Time":
			res, errx = WithTimezone(value.(time.Time), zone)
			if errx != nil {
				errx.AddComments("while with timezone")
				return res, errx
			}
		}
	}

	return res, errx
}

// ParseUTCWithTimezone to parse time with timezone and default UTC zone from anything
func ParseUTCWithTimezone(value interface{}, zone string) (res time.Time, errx serror.SError) {
	res = getEmptyTime()

	if utinterface.IsNil(value) {
		errx = serror.New("Cannot parse from nil")
		return res, errx
	}

	val := reflect.ValueOf(value)
	switch val.Kind() {
	case reflect.String:
		res, errx = ParseUTCFromStringWithTimezone(val.String(), zone)
		if errx != nil {
			errx.AddComments("while parse utc from string with timezone")
			return res, errx
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		res, errx = ParseUTCFromIntegerWithTimezone(val.Int(), zone)
		if errx != nil {
			errx.AddComments("while parse utc from integer with timezone")
			return res, errx
		}

	case reflect.Float32, reflect.Float64:
		res, errx = ParseUTCFromIntegerWithTimezone(int64(val.Float()), zone)
		if errx != nil {
			errx.AddComments("while parse utc from integer with timezone")
			return res, errx
		}

	default:
		switch val.Type().String() {
		case "time.Time":
			res, errx = WithTimezone(value.(time.Time), zone)
			if errx != nil {
				errx.AddComments("while with timezone")
				return res, errx
			}
		}
	}

	return res, errx
}

// ParseForceTimezone to parse time with force timezone from anything
func ParseForceTimezone(value interface{}, zone string) (res time.Time, errx serror.SError) {
	res = getEmptyTime()

	if utinterface.IsNil(value) {
		errx = serror.New("Cannot parse from nil")
		return res, errx
	}

	val := reflect.ValueOf(value)
	switch val.Kind() {
	case reflect.String:
		res, errx = ParseFromStringForceTimezone(val.String(), zone)
		if errx != nil {
			errx.AddComments("while parse from string force timezone")
			return res, errx
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		res, errx = ParseFromIntegerForceTimezone(val.Int(), zone)
		if errx != nil {
			errx.AddComments("while parse from integer force timezone")
			return res, errx
		}

	case reflect.Float32, reflect.Float64:
		res, errx = ParseFromIntegerForceTimezone(int64(val.Float()), zone)
		if errx != nil {
			errx.AddComments("while parse from integer force timezone")
			return res, errx
		}

	default:
		switch val.Type().String() {
		case "time.Time":
			res, errx = ForceTimezone(value.(time.Time), zone)
			if errx != nil {
				errx.AddComments("while force timezone")
				return res, errx
			}
		}
	}

	return res, errx
}

// MostParseWithTimezone for most parse time with timezone from anything
func MostParseWithTimezone(value interface{}, zone string) time.Time {
	return Most(ParseWithTimezone(value, zone))
}

// MostUTCParseWithTimezone for most parse time with timezone and default UTC zone from anything
func MostUTCParseWithTimezone(value interface{}, zone string) time.Time {
	return Most(ParseUTCWithTimezone(value, zone))
}

// MostParseForceTimezone for most parse time force timezone from anything
func MostParseForceTimezone(value interface{}, zone string) time.Time {
	return Most(ParseForceTimezone(value, zone))
}

// ParseFromString to parse time from string
func ParseFromString(value string) (res time.Time, errx serror.SError) {
	res = getEmptyTime()

	for _, f := range materials {
		res, errx = ParseWithFormat(f, value)
		if errx == nil {
			return res, errx
		}
	}

	if utint.IsInteger(value) {
		res, errx = ParseFromInteger(utint.StringToInt(value, 0))
		if errx == nil {
			return res, errx
		}
	}

	errx = serror.Newf("Cannot parse time from %s", value)
	return res, errx
}

// ParseUTCFromString to parse time from string with default UTC zone
func ParseUTCFromString(value string) (res time.Time, errx serror.SError) {
	res = getEmptyTime()

	for _, f := range materials {
		res, errx = ParseUTCWithFormat(f, value)
		if errx == nil {
			return res, errx
		}
	}

	if utint.IsInteger(value) {
		res, errx = ParseUTCFromInteger(utint.StringToInt(value, 0))
		if errx == nil {
			return res, errx
		}
	}

	errx = serror.Newf("Cannot parse time from %s", value)
	return res, errx
}

// ParseFromStringWithTimezone to parse time with timezone from string
func ParseFromStringWithTimezone(value string, zone string) (res time.Time, errx serror.SError) {
	res, errx = ParseFromString(value)
	if errx != nil {
		errx.AddComments("while parse from string")
		return res, errx
	}

	res, errx = WithTimezone(res, zone)
	if errx != nil {
		errx.AddComments("while with timezone")
		return res, errx
	}

	return res, errx
}

// ParseUTCFromStringWithTimezone to parse time with timezone and default UTC zone from string
func ParseUTCFromStringWithTimezone(value string, zone string) (res time.Time, errx serror.SError) {
	res, errx = ParseUTCFromString(value)
	if errx != nil {
		errx.AddComments("while parse utc from string")
		return res, errx
	}

	res, errx = WithTimezone(res, zone)
	if errx != nil {
		errx.AddComments("while with timezone")
		return res, errx
	}

	return res, errx
}

// ParseFromStringForceTimezone to parse time force timezone from string
func ParseFromStringForceTimezone(value string, zone string) (res time.Time, errx serror.SError) {
	res, errx = ParseFromString(value)
	if errx != nil {
		errx.AddComments("while parse from string")
		return res, errx
	}

	res, errx = ForceTimezone(res, zone)
	if errx != nil {
		errx.AddComments("while force timezone")
		return res, errx
	}

	return res, errx
}

//> Formating

// ToString to convert time to string
//
// Deprecated: use Format instead
func ToString(format DateFormat, value time.Time) string {
	return Format(format, value)
}

// Format to formating time to string
func Format(format DateFormat, value time.Time) string {
	return value.Format(GoLayout(format))
}
