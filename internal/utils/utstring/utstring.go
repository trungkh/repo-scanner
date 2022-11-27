package utstring

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rivo/uniseg"
	uuid "github.com/satori/go.uuid"
)

// IntToString to cast from integer to string
//
// **@Params:** [ `v`: value ]
//
// **@Returns:** [ `$1`: string ]
func IntToString(v int) string {
	return Int64ToString(int64(v))
}

// UintToString to cast from unsign integer to string
//
// **@Params:** [ `v`: value ]
//
// **@Returns:** [ `$1`: string ]
func UintToString(v uint) string {
	return Uint64ToString(uint64(v))
}

// Int64ToString to cast from big integer to string
//
// **@Params:** [ `v`: value ]
//
// **@Returns:** [ `$1`: string ]
func Int64ToString(v int64) string {
	return strconv.FormatInt(v, 10)
}

// Uint64ToString to cast from unsign big integer to string
//
// **@Params:** [ `v`: value ]
//
// **@Returns:** [ `$1`: string ]
func Uint64ToString(v uint64) string {
	return strconv.FormatUint(v, 10)
}

// BoolToString to cast from boolean to string
//
// **@Params:** [ `v`: value ]
//
// **@Returns:** [ `$1`: string ]
func BoolToString(v bool) string {
	return strconv.FormatBool(v)
}

// FloatToString to cast from float to string
//
// **@Params:** [ `v`: value ]
//
// **@Returns:** [ `$1`: string ]
func FloatToString(v float64) string {
	str := fmt.Sprintf("%f", v)
	if strings.Contains(str, ".") {
		str = strings.TrimRight(strings.TrimRight(str, "0"), ".")
	}

	return str
}

// IsNumber function
//
// **@Params:** [ `v`: value ]
//
// **@Returns:** [ `$1`: number flag ]
func IsNumber(v string) bool {
	if _, err := strconv.Atoi(v); err == nil {
		return true
	}
	return false
}

// Length function
//
// **@Params:** [ `s`: string ]
//
// **@Returns:** [ `$1`: length ]
func Length(s string) int64 {
	return int64(uniseg.GraphemeClusterCount(s))
}

// LeftPad function
//
// **@Params:** [ `s`: string; `l`: length; `p`: pad char ]
//
// **@Returns:** [ `$1`: padded string ]
func LeftPad(s string, l int, p string) string {
	for i := Length(s); i < int64(l); i++ {
		s = p + s
	}
	return s
}

// RightPad function
//
// **@Params:** [ `s`: string; `l`: length; `p`: pad char ]
//
// **@Returns:** [ `$1`: padded string ]
func RightPad(s string, l int, p string) string {
	for i := Length(s); i < int64(l); i++ {
		s += p
	}
	return s
}

// Sub function
//
// **@Params:** [ `s`: string; `f`: from; `p`: length ]
//
// **@Returns:** [ `$1`: string ]
func Sub(s string, f int, l int) string {
	if l == 0 {
		l = len(s) - f
	} else if l < 0 {
		l = len(s) + l
	}

	return s[f : f+l]
}

// RandString function
func RandString(length int, material string) string {
	e, r := len(material), ""

	rand.Seed(time.Now().UnixNano())
	for i := 0; i < length; i++ {
		c := int(math.Floor(rand.Float64() * float64(e)))
		r += string([]rune(material)[c])
	}
	return r
}

// ExRandString function
func ExRandString(length int) string {
	return RandString(length, "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
}

// WrapDoubleQuote function
//
// **@Params:** [ `s`: string ]
//
// **@Returns:** [ `$1`: wrapped string ]
func WrapDoubleQuote(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, "'") + `"`
}

// WrapSingleQuote function
//
// **@Params:** [ `s`: string ]
//
// **@Returns:** [ `$1`: wrapped string ]
func WrapSingleQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `"`) + "'"
}

// Slug for making slug string
func Slug(value string) string {
	a := "1234567890abcdefghijklmnopqrstuvwxyz "
	vals := []rune(strings.ToLower(value))

	res := ""
	for i := 0; i < len(vals); i++ {
		if strings.Contains(a, string(vals[i])) {
			res += string(vals[i])
		}
	}
	res = strings.ReplaceAll(strings.TrimSpace(res), " ", "-")

	return res
}

// Env to get environment variable
func Env(key string, def ...string) string {
	return Chains(append([]string{os.Getenv(key)}, def...)...)
}

// Index to find string with offset
func Index(s string, value string, offset int) int {
	if offset > 0 {
		s = Sub(s, offset, 0)
	}

	if offset < 0 {
		offset = 0
	}

	i := strings.Index(s, value)
	if i >= 0 {
		return i + offset
	}
	return i
}

// Indexs to get all match string indexs
func Indexs(s string, value string) []int {
	res := []int{}

	i := 0
	for {
		i = Index(s, value, i)
		if i >= 0 {
			res = append(res, i)
			i += len(value)
			continue
		}
		break
	}
	return res
}

// Trim whitespace
func Trim(value string) string {
	return strings.TrimSpace(value)
}

// Chains function
func Chains(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// MD5 function
func MD5(val string) string {
	mdcp := md5.New()

	_, err := mdcp.Write([]byte(val))
	if err != nil {
		return ""
	}

	return hex.EncodeToString(mdcp.Sum(nil))
}

func SHA1(val string) string {
	sha := sha1.New()
	_, err := sha.Write([]byte(val))
	if err != nil {
		return ""
	}

	return hex.EncodeToString(sha.Sum(nil))
}

func GUID() string {
	guid, err := uuid.NewV1()
	if err == nil {
		return guid.String()
	}

	guid, err = uuid.NewV4()
	if err == nil {
		return guid.String()
	}

	return "00000000-0000-0000-0000-000000000000"
}
