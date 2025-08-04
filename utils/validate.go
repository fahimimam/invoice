package utils

import (
	emailverifier "github.com/AfterShip/email-verifier"
	"math/rand"
	"reflect"
	"regexp"
	"time"
)

type UserType string

const (
	Employer UserType = "employer"
	Employee UserType = "employee"
)

func IsValueEmpty(val interface{}) bool {
	if val == nil {
		return true
	}

	switch v := val.(type) {
	// String types
	case string:
		return v == ""
	case *string:
		return v == nil || *v == ""

	// Numeric types
	case int, int8, int16, int32, int64:
		return v == 0
	case uint, uint8, uint16, uint32, uint64:
		return v == 0
	case float32, float64:
		return v == 0.0
	case complex64, complex128:
		return v == 0
	case *int, *int8, *int16, *int32, *int64:
		return v == nil || reflect.ValueOf(v).Elem().Int() == 0
	case *uint, *uint8, *uint16, *uint32, *uint64:
		return v == nil || reflect.ValueOf(v).Elem().Uint() == 0
	case *float32, *float64:
		return v == nil || reflect.ValueOf(v).Elem().Float() == 0.0
	case *complex64, *complex128:
		return v == nil || reflect.ValueOf(v).Elem().Complex() == 0

	// Boolean
	case bool:
		return !v
	case *bool:
		return v == nil || !*v

	// Time
	case time.Time:
		return v.IsZero()
	case *time.Time:
		return v == nil || v.IsZero()

	// Slice/Array/Map
	case []interface{}:
		return len(v) == 0
	case map[interface{}]interface{}:
		return len(v) == 0
	case []byte:
		return len(v) == 0

	// Interface and pointer cases
	case interface{}:
		if reflect.ValueOf(v).IsNil() {
			return true
		}
		return IsValueEmpty(reflect.ValueOf(v).Elem().Interface())

	default:
		// For custom types, you might want to implement String() or IsZero()
		// and handle them here
		if rv := reflect.ValueOf(v); rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				return true
			}
			return IsValueEmpty(rv.Elem().Interface())
		}

		// For structs, you could recursively check all fields if needed
		return false
	}
}

func GenerateCode(low, hi int) int {
	return low + rand.Intn(hi-low)
}

func IsInValidEmail(email string) bool {
	verifier := emailverifier.NewVerifier()
	ret, err := verifier.Verify(email)
	if err != nil {
		return true
	}
	if !ret.Syntax.Valid {
		return true
	}
	return false
}

func IsValidPhoneNumber(phoneNumber string) bool {
	regex := "^(?:\\+?88|0088)?01[15-9]\\d{8}$"
	reg := regexp.MustCompile(regex)
	if !reg.MatchString(phoneNumber) {
		return false
	}
	return true
}

func IsUserType(userType UserType) bool {
	if userType == Employer || userType == Employee {
		return true
	}
	return false
}
