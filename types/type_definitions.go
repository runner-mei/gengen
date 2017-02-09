package types

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"math"
	"math/big"
	//"labix.org/v2/mgo/bson"
	"encoding/json"
	"net"
	"regexp"
	"strconv"
	"time"
)

var InvalidValueError = errors.New("value is invalid.")
var InvalidIPError = errors.New("syntex error, it is not IP.")

type Password string
type IPAddress string
type PhysicalAddress string

type TypeDefinition interface {
	Name() string
	MakeValue() interface{}
	CreateEnumerationValidator(values []string) (Validator, error)
	CreatePatternValidator(pattern string) (Validator, error)
	CreateRangeValidator(minValue, maxValue string) (Validator, error)
	CreateLengthValidator(minLength, maxLength string) (Validator, error)
	Parse(v string) (interface{}, error)
	ToInternal(v interface{}) (interface{}, error)
	ToExternal(v interface{}) interface{}
}

type integerType struct {
	DName string
}

func (self *integerType) Name() string {
	return self.DName
}

func (self *integerType) MakeValue() interface{} {
	return &sql.NullInt64{Int64: 0, Valid: false}
}

func (self *integerType) CreateEnumerationValidator(ss []string) (Validator, error) {
	if nil == ss || 0 == len(ss) {
		return nil, errors.New("values is null or empty.")
	}

	values := make([]interface{}, 0, len(ss))
	for i, s := range ss {
		v, err := strconv.ParseInt(s, 10, 64)
		if nil != err {
			return nil, fmt.Errorf("value[%d] '%v' is syntex error, %s", i, s, err.Error())
		}
		values = append(values, int64(v))
	}
	return &EnumerationValidator{Values: values}, nil
}

func (self *integerType) CreatePatternValidator(pattern string) (Validator, error) {
	panic("not supported")
}

func (self *integerType) CreateRangeValidator(minValue, maxValue string) (Validator, error) {
	var min, max int64
	var err error
	hasMin := false
	hasMax := false

	if "" != minValue {
		hasMin = true
		min, err = strconv.ParseInt(minValue, 10, 64)
		if nil != err {
			return nil, fmt.Errorf("minValue '%s' is not a integer", minValue)
		}
	}

	if "" != maxValue {
		hasMax = true
		max, err = strconv.ParseInt(maxValue, 10, 64)
		if nil != err {
			return nil, fmt.Errorf("maxValue '%s' is not a integer", maxValue)
		}
	}
	return &IntegerValidator{HasMax: hasMax, MaxValue: max,
		HasMin: hasMin, MinValue: min}, nil
}

func (self *integerType) CreateLengthValidator(minLength,
	maxLength string) (Validator, error) {
	panic("not supported")
}

func (self *integerType) ToInternal(value interface{}) (interface{}, error) {
	return ToInteger64(value)
}

func ToInteger64(value interface{}) (int64, error) {
	switch v := value.(type) {
	case json.Number:
		if i64, e := v.Int64(); nil == e {
			return i64, nil
		}
		if f64, e := v.Float64(); nil == e {
			if float64(math.MaxInt64) > f64 {
				return int64(f64), nil
			}
			return int64(0), errors.New("it is float64, value is overflow.")
		}
		return int64(0), errors.New("json.Number is not int64 and float64?")
	case string:
		i64, err := strconv.ParseInt(v, 10, 64)
		if nil == err {
			return int64(i64), nil
		}
	case int:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case uint:
		if math.MaxInt64 > int64(v) {
			return int64(v), nil
		}
		return int64(0), errors.New("it is uint, value is overflow.")
	case uint32:
		return int64(v), nil
	case uint64:
		if uint64(math.MaxInt64) > v {
			return int64(v), nil
		}
		return int64(0), errors.New("it is uint64, value is overflow.")
	case float32:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case []byte:
		i64, err := strconv.ParseInt(string(v), 10, 64)
		if nil == err {
			return int64(i64), nil
		}
	case *int64:
		return *v, nil
	case *sql.NullInt64:
		if !v.Valid {
			return 0, InvalidValueError
		}
		return v.Int64, nil
	}
	return int64(0), errors.New("ToInternal to int64 failed")
}

func (self *integerType) ToExternal(value interface{}) interface{} {
	return value
}

func (self *integerType) Parse(s string) (interface{}, error) {
	i64, e := strconv.ParseInt(s, 10, 64)
	return i64, e
}

type bigintegerType struct {
	PName string
}

func (self *bigintegerType) Name() string {
	return self.PName
}

func (self *bigintegerType) MakeValue() interface{} {
	return &sql.NullFloat64{Float64: 0, Valid: false}
}

func (self *bigintegerType) CreateEnumerationValidator(ss []string) (Validator, error) {
	if nil == ss || 0 == len(ss) {
		return nil, errors.New("values is null or empty.")
	}

	values := make([]big.Int, 0, len(ss))
	for i, s := range ss {
		v, err := parseBigInteger(s)
		if nil != err {
			return nil, fmt.Errorf("value[%d] '%v' is syntex error, %s", i, s, err.Error())
		}
		values = append(values, *v)
	}
	return &BigIntegerValidator{Values: values}, nil
}

func (self *bigintegerType) CreatePatternValidator(pattern string) (Validator, error) {
	panic("not supported")
}

func (self *bigintegerType) CreateRangeValidator(minValue, maxValue string) (Validator, error) {
	var min, max float64
	var err error
	hasMin := false
	hasMax := false

	if "" != minValue {
		hasMin = true
		min, err = strconv.ParseFloat(minValue, 64)
		if nil != err {
			return nil, fmt.Errorf("minValue '%s' is not a biginteger", minValue)
		}
	}

	if "" != maxValue {
		hasMax = true
		max, err = strconv.ParseFloat(maxValue, 64)
		if nil != err {
			return nil, fmt.Errorf("maxValue '%s' is not a biginteger", maxValue)
		}
	}
	return &DecimalValidator{HasMax: hasMax, MaxValue: max,
		HasMin: hasMin, MinValue: min}, nil
}

func (self *bigintegerType) CreateLengthValidator(minLength,
	maxLength string) (Validator, error) {
	panic("not supported")
}

func (self *bigintegerType) ToInternal(value interface{}) (interface{}, error) {
	return ToBigInteger(value)
}

var bigError = errors.New("convert to bigInteger failed")

func ToBigIntegerFromFloat64(f float64) (*big.Int, error) {
	if f < 0 {
		if f >= math.MinInt64 {
			return big.NewInt(int64(f)), nil
		}
	} else if f <= math.MaxUint64 {
		return new(big.Int).SetUint64(uint64(f)), nil
	}
	return nil, bigError
}

func parseBigInteger(s string) (*big.Int, error) {
	var bi big.Int
	if _, ok := bi.SetString(s, 10); ok {
		return &bi, nil
	} else if f64, e := strconv.ParseFloat(s, 64); nil == e {
		return ToBigIntegerFromFloat64(f64)
	}
	return nil, bigError
}

func ToBigInteger(value interface{}) (*big.Int, error) {
	switch v := value.(type) {
	case *sql.NullFloat64:
		if v.Valid {
			return ToBigIntegerFromFloat64(v.Float64)
		}
	case json.Number:
		return parseBigInteger(v.String())
	case string:
		return parseBigInteger(v)
	case int:
		return big.NewInt(int64(v)), nil
	case int32:
		return big.NewInt(int64(v)), nil
	case int64:
		return big.NewInt(v), nil
	case uint:
		return new(big.Int).SetUint64(uint64(v)), nil
	case uint32:
		return new(big.Int).SetUint64(uint64(v)), nil
	case uint64:
		return new(big.Int).SetUint64(v), nil
	case float32:
		return ToBigIntegerFromFloat64(float64(v))
	case float64:
		return ToBigIntegerFromFloat64(v)
	case []byte:
		if nil != v && 0 != len(v) {
			return parseBigInteger(string(v))
		}
	case *int64:
		return big.NewInt(*v), nil
	case *uint64:
		return new(big.Int).SetUint64(*v), nil
	}
	return nil, bigError
}

func (self *bigintegerType) ToExternal(value interface{}) interface{} {
	return value
}

func (self *bigintegerType) Parse(s string) (interface{}, error) {
	return parseBigInteger(s)
}

type decimalType struct {
	PName string
}

func (self *decimalType) Name() string {
	return self.PName
}

func (self *decimalType) MakeValue() interface{} {
	return &sql.NullFloat64{Float64: 0, Valid: false}
}

func (self *decimalType) CreateEnumerationValidator(ss []string) (Validator, error) {
	if nil == ss || 0 == len(ss) {
		return nil, errors.New("values is null or empty.")
	}

	values := make([]interface{}, 0, len(ss))
	for i, s := range ss {
		v, err := strconv.ParseFloat(s, 64)
		if nil != err {
			return nil, fmt.Errorf("value[%d] '%v' is syntex error, %s", i, s, err.Error())
		}
		values = append(values, float64(v))
	}
	return &EnumerationValidator{Values: values}, nil
}

func (self *decimalType) CreatePatternValidator(pattern string) (Validator, error) {
	panic("not supported")
}

func (self *decimalType) CreateRangeValidator(minValue, maxValue string) (Validator, error) {
	var min, max float64
	var err error
	hasMin := false
	hasMax := false

	if "" != minValue {
		hasMin = true
		min, err = strconv.ParseFloat(minValue, 64)
		if nil != err {
			return nil, fmt.Errorf("minValue '%s' is not a integer", minValue)
		}
	}

	if "" != maxValue {
		hasMax = true
		max, err = strconv.ParseFloat(maxValue, 64)
		if nil != err {
			return nil, fmt.Errorf("maxValue '%s' is not a integer", maxValue)
		}
	}
	return &DecimalValidator{HasMax: hasMax, MaxValue: max, HasMin: hasMin, MinValue: min}, nil
}

func (self *decimalType) CreateLengthValidator(minLength, maxLength string) (Validator, error) {
	panic("not supported")
}

func (self *decimalType) ToInternal(value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case json.Number:
		return v.Float64()
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case float32:
		return float64(v), nil
	case float64:
		return float64(v), nil
	case string:
		f64, err := strconv.ParseFloat(v, 64)
		if nil == err {
			return float64(f64), nil
		}
		return float64(0), err
	case []byte:
		i64, err := strconv.ParseFloat(string(v), 64)
		if nil == err {
			return float64(i64), nil
		}
	case *float64:
		return *v, nil
	case *sql.NullFloat64:
		if !v.Valid {
			return nil, InvalidValueError
		}
		return v.Float64, nil
	}
	return float64(0), errors.New("ToInternal to float64 failed")
}

func (self *decimalType) ToExternal(value interface{}) interface{} {
	return value
}

func (self *decimalType) Parse(s string) (interface{}, error) {
	f64, e := strconv.ParseFloat(s, 64)
	return f64, e
}

type stringType struct {
	PName string
}

func (self *stringType) Name() string {
	return self.PName
}

func (self *stringType) MakeValue() interface{} {
	return &sql.NullString{String: "", Valid: false}
}

func (self *stringType) CreateEnumerationValidator(values []string) (Validator, error) {
	if nil == values || 0 == len(values) {
		return nil, errors.New("values is null or empty")
	}

	new_values := make([]string, len(values))
	for i, s := range values {
		if "" == s {
			return nil, fmt.Errorf("value[%d] is empty", i)
		}
		new_values = append(new_values, string(s))
	}
	return &StringEnumerationValidator{Values: new_values}, nil
}

func (self *stringType) CreatePatternValidator(pattern string) (Validator, error) {
	if "" == pattern {
		return nil, errors.New("pattern is empty")
	}

	p, err := regexp.Compile(pattern)
	if nil != err {
		return nil, err
	}
	return &PatternValidator{Pattern: p}, nil
}

func (self *stringType) CreateRangeValidator(minValue, maxValue string) (Validator, error) {
	panic("not supported")
}

func (self *stringType) CreateLengthValidator(minLength, maxLength string) (Validator, error) {
	var err error
	var min int64 = -1
	var max int64 = -1

	if "" != minLength {
		min, err = strconv.ParseInt(minLength, 10, 32)
		if nil != err {
			return nil, fmt.Errorf("minLength '%s' is not a integer", minLength)
		}
	}

	if "" != maxLength {
		max, err = strconv.ParseInt(maxLength, 10, 32)
		if nil != err {
			return nil, fmt.Errorf("maxLength '%s' is not a integer", maxLength)
		}
	}
	return &StringLengthValidator{MaxLength: int(max), MinLength: int(min)}, nil
}

func (self *stringType) ToInternal(value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case json.Number:
		return v.String(), nil
	case string:
		return v, nil
	case *string:
		return *v, nil
	case []byte:
		return string(v), nil
	case uint:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint8:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint16:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint32:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint64:
		return strconv.FormatUint(uint64(v), 10), nil
	case int:
		return strconv.FormatInt(int64(v), 10), nil
	case int8:
		return strconv.FormatInt(int64(v), 10), nil
	case int16:
		return strconv.FormatInt(int64(v), 10), nil
	case int32:
		return strconv.FormatInt(int64(v), 10), nil
	case int64:
		return strconv.FormatInt(int64(v), 10), nil
	case float32:
		return strconv.FormatFloat(float64(v), 'e', -1, 64), nil
	case float64:
		return strconv.FormatFloat(float64(v), 'e', -1, 64), nil
	case *sql.NullString:
		if !v.Valid {
			return nil, InvalidValueError
		}
		return v.String, nil
	}

	s, e := json.Marshal(value)
	if nil == e {
		return s, nil
	}
	return "", errors.New("ToInternal to SqlString failed")
}

func (self *stringType) ToExternal(value interface{}) interface{} {
	return value
}

func (self *stringType) Parse(s string) (interface{}, error) {
	return s, nil
}

type dateTimeType struct {
	Layout string //"2006-01-02 15:04:05"
	PName  string //datetime
}

func (self *dateTimeType) Name() string {
	return self.PName
}

// NullTime represents an time that may be null.
// NullTime implements the Scanner interface so
// it can be used as a scan destination, similar to NullTime.
type NullTime struct {
	Time  time.Time
	Valid bool // Valid is true if Int64 is not NULL
}

// Scan implements the Scanner interface.
func (n *NullTime) Scan(value interface{}) error {
	if value == nil {
		n.Time, n.Valid = time.Time{}, false
		return nil
	}

	n.Time, n.Valid = value.(time.Time)
	if !n.Valid {
		if s, ok := value.(string); ok {
			var e error
			for _, layout := range []string{time.StampNano, time.StampMicro, time.StampMilli, time.Stamp} {
				if n.Time, e = time.ParseInLocation(layout, s, time.UTC); nil == e {
					n.Valid = true
					break
				}
			}
		}
	}
	return nil
}

// Value implements the driver Valuer interface.
func (n NullTime) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Time, nil
}

func (self *dateTimeType) MakeValue() interface{} {
	return &NullTime{Valid: false}
}

func (self *dateTimeType) CreateEnumerationValidator(ss []string) (Validator, error) {
	if nil == ss || 0 == len(ss) {
		return nil, errors.New("values is null or empty.")
	}

	values := make([]interface{}, 0, len(ss))
	for i, s := range ss {
		t, err := time.Parse(self.Layout, s)
		if nil != err {
			return nil, fmt.Errorf("value[%d] '%v' is syntex error, %s", i, s, err.Error())
		}
		values = append(values, t)
	}
	return &EnumerationValidator{Values: values}, nil
}

func (self *dateTimeType) CreatePatternValidator(pattern string) (Validator, error) {
	panic("not supported")
}

func (self *dateTimeType) CreateRangeValidator(minValue, maxValue string) (Validator, error) {
	var min, max time.Time
	var err error
	hasMin := false
	hasMax := false

	if "" != minValue {
		hasMin = true
		min, err = time.Parse(self.Layout, minValue)
		if nil != err {
			return nil, fmt.Errorf("minValue '%s' is not a time(%s)", minValue, self.Layout)
		}
	}

	if "" != maxValue {
		hasMax = true
		max, err = time.Parse(self.Layout, maxValue)
		if nil != err {
			return nil, fmt.Errorf("maxValue '%s' is not a time(%s)", maxValue, self.Layout)
		}
	}
	return &DateValidator{HasMax: hasMax, MaxValue: max, HasMin: hasMin, MinValue: min}, nil
}

func (self *dateTimeType) CreateLengthValidator(minLength, maxLength string) (Validator, error) {
	panic("not supported")
}

func (self *dateTimeType) ToInternal(v interface{}) (interface{}, error) {
	switch value := v.(type) {
	case string:
		t, err := time.Parse(self.Layout, value)
		if nil != err {
			return nil, err
		}
		return t, nil
	case *string:
		t, err := time.Parse(self.Layout, *value)
		if nil != err {
			return nil, err
		}
		return t, nil
	case time.Time:
		return value, nil
	case *time.Time:
		return *value, nil
	case *NullTime:
		if !value.Valid {
			return nil, InvalidValueError
		}
		return value.Time, nil
	}

	return nil, errors.New("syntex error, it is not a datetime")
}

func (self *dateTimeType) ToExternal(value interface{}) interface{} {
	return value
}

func (self *dateTimeType) Parse(s string) (interface{}, error) {
	t, e := time.Parse(self.Layout, s)
	return t, e
}

type durationType struct {
	PName string
}

func (self *durationType) Name() string {
	return self.PName
}

func (self *durationType) MakeValue() interface{} {
	return &sql.NullInt64{Int64: 0, Valid: false}
}

func (self *durationType) CreateEnumerationValidator(ss []string) (Validator, error) {
	if nil == ss || 0 == len(ss) {
		return nil, errors.New("values is null or empty.")
	}

	values := make([]int64, 0, len(ss))
	for i, s := range ss {
		v, err := time.ParseDuration(s)
		if nil != err {
			return nil, fmt.Errorf("value[%d] '%v' is syntex error, %s", i, s, err.Error())
		}
		values = append(values, int64(v))
	}
	return &IntegerEnumerationValidator{Values: values}, nil
}

func (self *durationType) CreatePatternValidator(pattern string) (Validator, error) {
	panic("not supported")
}

func (self *durationType) CreateRangeValidator(minValue, maxValue string) (Validator, error) {
	var min, max float64
	var err error
	hasMin := false
	hasMax := false

	if "" != minValue {
		hasMin = true
		min, err = strconv.ParseFloat(minValue, 64)
		if nil != err {
			return nil, fmt.Errorf("minValue '%s' is not a integer", minValue)
		}
	}

	if "" != maxValue {
		hasMax = true
		max, err = strconv.ParseFloat(maxValue, 64)
		if nil != err {
			return nil, fmt.Errorf("maxValue '%s' is not a integer", maxValue)
		}
	}
	return &DecimalValidator{HasMax: hasMax, MaxValue: max, HasMin: hasMin, MinValue: min}, nil
}

func (self *durationType) CreateLengthValidator(minLength, maxLength string) (Validator, error) {
	panic("not supported")
}

func (self *durationType) ToInternal(value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case json.Number:
		if i64, e := v.Int64(); nil == e {
			return time.Duration(i64), nil
		}
		if f64, e := v.Float64(); nil == e {
			if float64(math.MaxInt64) > f64 {
				return time.Duration(f64), nil
			}
			return time.Duration(0), errors.New("it is float64, value is overflow.")
		}
		return time.Duration(0), errors.New("json.Number is not int64 and float64?")
	case string:
		i64, err := time.ParseDuration(v)
		if nil == err {
			return time.Duration(i64), nil
		}
	case time.Duration:
		return v, nil
	case int:
		return time.Duration(v), nil
	case int32:
		return time.Duration(v), nil
	case int64:
		return v, nil
	case uint:
		return time.Duration(v), nil
	case uint32:
		return time.Duration(v), nil
	case uint64:
		if uint64(math.MaxInt64) > v {
			return time.Duration(v), nil
		}
		return time.Duration(0), errors.New("it is uint64, value is overflow.")
	case float32:
		return time.Duration(v), nil
	case float64:
		return time.Duration(v), nil
	case []byte:
		i64, err := strconv.ParseInt(string(v), 10, 64)
		if nil == err {
			return time.Duration(i64), nil
		}
	case *int64:
		return *v, nil
	case *sql.NullInt64:
		if !v.Valid {
			return nil, InvalidValueError
		}
		return v.Int64, nil
	}
	return time.Duration(0), errors.New("ToInternal to int64 failed")
}

func (self *durationType) ToExternal(value interface{}) interface{} {
	return value
}

func (self *durationType) Parse(s string) (interface{}, error) {
	return time.ParseDuration(s)
}

type ipAddressType struct {
	PName string
}

func (self *ipAddressType) Name() string {
	return self.PName
}

// NullIPAddress represents an ip address that may be null.
// NullIPAddress implements the Scanner interface so
// it can be used as a scan destination, similar to NullIPAddress.
type NullIPAddress struct {
	String string
	Valid  bool // Valid is true if Int64 is not NULL
}

// Scan implements the Scanner interface.
func (n *NullIPAddress) Scan(value interface{}) error {
	if value == nil {
		n.Valid = false
		return nil
	}

	var nullString sql.NullString
	e := nullString.Scan(value)
	if nil == e {
		n.Valid = nullString.Valid
		n.String = nullString.String
	}
	return e
}

// Value implements the driver Valuer interface.
func (n NullIPAddress) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	if "" == n.String {
		return nil, nil
	}

	ip := net.ParseIP(n.String)
	if nil == ip {
		if ip, _, e := net.ParseCIDR(n.String); nil == e {
			return IPAddress(ip.String()), nil
		}
		return nil, InvalidIPError
	}
	return IPAddress(ip.String()), nil
}

func (self *ipAddressType) MakeValue() interface{} {
	return &NullIPAddress{Valid: false}
}

func (self *ipAddressType) CreateEnumerationValidator(values []string) (Validator, error) {
	panic("not supported")
}
func (self *ipAddressType) CreatePatternValidator(pattern string) (Validator, error) {
	panic("not supported")
}
func (self *ipAddressType) CreateRangeValidator(minValue, maxValue string) (Validator, error) {
	panic("not supported")
}
func (self *ipAddressType) CreateLengthValidator(minLength, maxLength string) (Validator, error) {
	panic("not supported")
}
func (self *ipAddressType) ToInternal(v interface{}) (interface{}, error) {
	if nil == v {
		return nil, nil
	}

	switch value := v.(type) {
	case string:
		ip := net.ParseIP(value)
		if nil == ip {
			return nil, InvalidIPError
		}

		addr := IPAddress(ip.String())
		return addr, nil
	case *string:
		ip := net.ParseIP(*value)
		if nil == ip {
			return nil, InvalidIPError
		}
		addr := IPAddress(ip.String())
		return addr, nil
	case []byte:
		ip := net.ParseIP(string(value))
		if nil == ip {
			return nil, InvalidIPError
		}
		addr := IPAddress(ip.String())
		return addr, nil
	case net.IP:
		addr := IPAddress(value.String())
		return addr, nil
	case *net.IP:
		addr := IPAddress(value.String())
		return addr, nil
	case IPAddress:
		return value, nil
	case *IPAddress:
		return *value, nil
	case *NullIPAddress:
		if !value.Valid {
			return nil, InvalidValueError
		}
		return value.Value()
	}

	return nil, InvalidIPError
}

func (self *ipAddressType) ToExternal(value interface{}) interface{} {
	if nil == value {
		return value
	}
	switch v := value.(type) {
	case string:
		return v
	case IPAddress:
		return string(v)
	default:
		panic(InvalidIPError.Error())
	}
}

func (self *ipAddressType) Parse(s string) (interface{}, error) {
	ip := net.ParseIP(s)
	if nil == ip {
		return nil, InvalidIPError
	}
	return IPAddress(ip.String()), nil
}

type physicalAddressType struct {
	PName string
}

// NullIPAddress represents an ip address that may be null.
// NullIPAddress implements the Scanner interface so
// it can be used as a scan destination, similar to NullIPAddress.
type NullPhysicalAddress struct {
	String string
	Valid  bool // Valid is true if Int64 is not NULL
}

// Scan implements the Scanner interface.
func (n *NullPhysicalAddress) Scan(value interface{}) error {
	if value == nil {
		n.Valid = false
		return nil
	}

	var nullString sql.NullString
	e := nullString.Scan(value)
	if nil == e {
		n.Valid = nullString.Valid
		n.String = nullString.String
	}
	return e
}

// Value implements the driver Valuer interface.
func (n NullPhysicalAddress) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}

	if 0 == len(n.String) {
		return nil, nil
	}
	mac, err := net.ParseMAC(n.String)
	if nil != err {
		return nil, err
	}
	return PhysicalAddress(mac.String()), nil
}

func (self *physicalAddressType) MakeValue() interface{} {
	return &NullPhysicalAddress{Valid: false}
}

func (self *physicalAddressType) Name() string {
	return self.PName
}

func (self *physicalAddressType) CreateEnumerationValidator(values []string) (Validator, error) {
	panic("not supported")
}

func (self *physicalAddressType) CreatePatternValidator(pattern string) (Validator, error) {
	panic("not supported")
}

func (self *physicalAddressType) CreateRangeValidator(minValue, maxValue string) (Validator, error) {
	panic("not supported")
}

func (self *physicalAddressType) CreateLengthValidator(minLength, maxLength string) (Validator, error) {
	panic("not supported")
}

func (self *physicalAddressType) ToInternal(v interface{}) (interface{}, error) {
	if nil == v {
		return nil, nil
	}
	switch value := v.(type) {
	case string:
		if "" == value {
			return nil, nil
		}
		mac, err := net.ParseMAC(value)
		if nil != err {
			return nil, err
		}
		return PhysicalAddress(mac.String()), nil
	case *string:
		if "" == *value {
			return nil, nil
		}
		mac, err := net.ParseMAC(*value)
		if nil != err {
			return nil, err
		}
		return PhysicalAddress(mac.String()), nil
	case []byte:
		if nil == value || 0 == len(value) {
			return nil, nil
		}
		mac, err := net.ParseMAC(string(value))
		if nil != err {
			return nil, err
		}
		return PhysicalAddress(mac.String()), nil
	case net.HardwareAddr:
		return PhysicalAddress(value.String()), nil
	case *net.HardwareAddr:
		return PhysicalAddress(value.String()), nil
	case PhysicalAddress:
		return value, nil
	case *PhysicalAddress:
		return *value, nil
	case *NullPhysicalAddress:
		if !value.Valid {
			return nil, InvalidValueError
		}
		return value.Value()
	default:
		return nil, fmt.Errorf("syntex error, it is not a physicalAddress - %t", value)
	}
	//return nil, errors.New("syntex error, it is not a physicalAddress")
}

func (self *physicalAddressType) ToExternal(value interface{}) interface{} {
	if nil == value {
		return value
	}
	switch v := value.(type) {
	case string:
		return v
	case PhysicalAddress:
		return string(v)
	case net.HardwareAddr:
		return v.String()
	case *net.HardwareAddr:
		return v.String()
	default:
		panic(fmt.Errorf("syntex error, it is not a physicalAddress - %t", value))
	}
}

func (self *physicalAddressType) Parse(s string) (interface{}, error) {
	mac, e := net.ParseMAC(s)
	return mac, e
}

type booleanType struct {
	DName string
}

func (self *booleanType) Name() string {
	return self.DName
}

func (self *booleanType) MakeValue() interface{} {
	return &sql.NullBool{Valid: false}
}

func (self *booleanType) CreateEnumerationValidator(values []string) (Validator, error) {
	panic("not supported")
}

func (self *booleanType) CreatePatternValidator(pattern string) (Validator, error) {
	panic("not supported")
}

func (self *booleanType) CreateRangeValidator(minValue, maxValue string) (Validator, error) {
	panic("not supported")
}

func (self *booleanType) CreateLengthValidator(minLength, maxLength string) (Validator, error) {
	panic("not supported")
}

func (self *booleanType) ToInternal(v interface{}) (interface{}, error) {
	switch value := v.(type) {
	case string:
		return self.Parse(value)
	case *string:
		return self.Parse(*value)
	case []byte:
		return self.Parse(string(value))
	case bool:
		return value, nil
	case *bool:
		return *value, nil
	case *sql.NullBool:
		if !value.Valid {
			return nil, InvalidValueError
		}
		return value.Value()
	}

	return nil, errors.New("syntex error, it is not a boolean")
}

func (self *booleanType) ToExternal(v interface{}) interface{} {
	return v
}

func (self *booleanType) Parse(s string) (interface{}, error) {
	switch s {
	case "true", "True", "TRUE", "yes", "Yes", "YES", "1":
		return true, nil
	case "false", "False", "FALSE", "no", "No", "NO", "0":
		return false, nil
	default:
		return nil, errors.New("syntex error, it is not a boolean")
	}
}

type objectIdType struct {
	TypeDefinition
}

func (self *objectIdType) Name() string {
	return "objectId"
}

// type objectIdType struct {
// }

// func (self *objectIdType) Name() string {
// 	return "objectId"
// }

// func (self *objectIdType) CreateEnumerationValidator(values []string) (Validator, error) {
// 	panic("not supported")
// }

// func (self *objectIdType) CreatePatternValidator(pattern string) (Validator, error) {
// 	panic("not supported")
// }

// func (self *objectIdType) CreateRangeValidator(minValue, maxValue string) (Validator, error) {
// 	panic("not supported")
// }

// func (self *objectIdType) CreateLengthValidator(minLength, maxLength string) (Validator, error) {
// 	panic("not supported")
// }

// func (self *objectIdType) ToInternal(v interface{}) (interface{}, error) {
// 	switch value := v.(type) {
// 	case string:
// 		return parseObjectIdHex(value)
// 	case *string:
// 		return parseObjectIdHex(*value)
// 	case bson.ObjectId:
// 		return value, nil
// 	case *bson.ObjectId:
// 		return *value, nil
// 	}

// 	return nil, errors.New("syntex error, it is not a boolean")
// }

type SqlIdTypeDefinition struct {
	PName string
}

func (self *SqlIdTypeDefinition) Name() string {
	return self.PName
}

func (self *SqlIdTypeDefinition) MakeValue() interface{} {
	return &sql.NullInt64{Valid: false}
}

func (self *SqlIdTypeDefinition) CreateEnumerationValidator(values []string) (Validator, error) {
	panic("not supported")
}

func (self *SqlIdTypeDefinition) CreatePatternValidator(pattern string) (Validator, error) {
	panic("not supported")
}

func (self *SqlIdTypeDefinition) CreateRangeValidator(minValue, maxValue string) (Validator, error) {
	panic("not supported")
}

func (self *SqlIdTypeDefinition) CreateLengthValidator(minLength, maxLength string) (Validator, error) {
	panic("not supported")
}

func (self *SqlIdTypeDefinition) ToInternal(v interface{}) (interface{}, error) {
	switch value := v.(type) {
	case json.Number:
		i, e := value.Int64()
		return i, e
	case string:
		i64, e := strconv.ParseInt(value, 10, 0)
		return int(i64), e
	case *string:
		i64, e := strconv.ParseInt(*value, 10, 0)
		return int(i64), e
	case []byte:
		i64, e := strconv.ParseInt(string(value), 10, 0)
		return int(i64), e
	case int:
		return value, nil
	case *int:
		return *value, nil
	case int32:
		return value, nil
	case *int32:
		return *value, nil
	case int64:
		return value, nil
	case *int64:
		return *value, nil
	case *sql.NullInt64:
		if !value.Valid {
			return nil, InvalidValueError
		}
		return value.Value()
	case float64:
		return int(value), nil
	case float32:
		return int(value), nil
	}

	return nil, errors.New("syntex error, it is not a objectId")
}

func (self *SqlIdTypeDefinition) ToExternal(v interface{}) interface{} {
	return v
}

func (self *SqlIdTypeDefinition) Parse(s string) (interface{}, error) {
	i64, e := strconv.ParseInt(s, 10, 64)
	return i64, e
}

type passwordType struct {
	stringType
}

type dynamicType struct {
	stringType
}

// NullIPAddress represents an ip address that may be null.
// NullIPAddress implements the Scanner interface so
// it can be used as a scan destination, similar to NullIPAddress.
type NullPassword struct {
	String string
	Valid  bool // Valid is true if Int64 is not NULL
}

// Scan implements the Scanner interface.
func (n *NullPassword) Scan(value interface{}) error {
	if value == nil {
		n.Valid = false
		return nil
	}

	var nullString sql.NullString
	e := nullString.Scan(value)
	if nil == e {
		n.Valid = nullString.Valid
		n.String = nullString.String
	}
	return e
}

// Value implements the driver Valuer interface.
func (n NullPassword) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}

	return Password(n.String), nil
}

func (self *passwordType) MakeValue() interface{} {
	return &NullPassword{Valid: false}
}

func (self *passwordType) ToInternal(v interface{}) (interface{}, error) {
	switch value := v.(type) {
	case string:
		return Password(value), nil
	case *string:
		return Password(*value), nil
	case []byte:
		return Password(string(value)), nil
	case Password:
		return value, nil
	case *Password:
		return *value, nil
	case *NullPassword:
		if !value.Valid {
			return nil, InvalidValueError
		}
		return value.Value()
	}

	return nil, errors.New("syntex error, it is not a password")
}

func (self *passwordType) ToExternal(value interface{}) interface{} {
	if nil == value {
		return value
	}
	switch v := value.(type) {
	case string:
		return v
	case Password:
		return string(v)
	default:
		panic("syntex error, it is not a password")
	}
}

func (self *passwordType) Parse(s string) (interface{}, error) {
	return Password(s), nil
}

var (
	DATETIMELAYOUT = time.RFC3339

	BooleanType         booleanType         = booleanType{"boolean"}
	IntegerType         integerType         = integerType{"integer"}
	DecimalType         decimalType         = decimalType{"decimal"}
	StringType          stringType          = stringType{"string"}
	DateTimeType        dateTimeType        = dateTimeType{PName: "datetime", Layout: DATETIMELAYOUT}
	IPAddressType       ipAddressType       = ipAddressType{PName: "ipAddress"}
	PhysicalAddressType physicalAddressType = physicalAddressType{PName: "physicalAddress"}
	PasswordType        passwordType        = passwordType{stringType{PName: "password"}}
	ObjectIdType        objectIdType        = objectIdType{&SqlIdTypeDefinition{PName: "objectId"}}
	BigIntegerType      bigintegerType      = bigintegerType{PName: "biginteger"}
	DynamicType         dynamicType         = dynamicType{stringType{PName: "dynamic"}}
	DurationType        durationType        = durationType{PName: "duration"}

	types = map[string]TypeDefinition{"boolean": &BooleanType,
		"integer":         &IntegerType,
		"decimal":         &DecimalType,
		"string":          &StringType,
		"datetime":        &DateTimeType,
		"duration":        &DurationType,
		"ipAddress":       &IPAddressType,
		"IPAddress":       &IPAddressType,
		"physicalAddress": &PhysicalAddressType,
		"PhysicalAddress": &PhysicalAddressType,
		"password":        &PasswordType,
		"objectId":        &ObjectIdType,
		"biginteger":      &BigIntegerType,
		"bigInteger":      &BigIntegerType,
		"dynamic":         &DynamicType}

	ToGoTypes = map[string]string{"boolean": "bool",
		"integer":         "int",
		"decimal":         "float64",
		"string":          "string",
		"datetime":        "time.Time",
		"duration":        "time.Duration",
		"ipAddress":       "net.IP",
		"IPAddress":       "net.IP",
		"physicalAddress": "[]byte",
		"PhysicalAddress": "[]byte",
		"password":        "string",
		"objectId":        "ObjectID",
		"biginteger":      "int",
		"bigInteger":      "int",
		"dynamic":         ""}
)

func RegisterTypeDefinition(t TypeDefinition) {
	types[t.Name()] = t
}

func GetTypeDefinition(t string) TypeDefinition {
	if td, ok := types[t]; ok {
		return td
	}
	panic("'" + t + "' is unknown type.")
}
