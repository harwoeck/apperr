package errdetails

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// A Code is an unsigned 32-bit error code as defined by the gRPC spec.
// Only values 0 through 16 (OK through Unauthenticated) are valid.
// Use [Code.Valid] to check whether a Code is within the defined range.
type Code uint32

const (
	// OK represents a successful status. It is NOT a valid error code and
	// must not be used with [apperr.AppError]. It is exported only for
	// completeness with the gRPC code enum.
	OK                 Code = 0
	Canceled           Code = 1
	Unknown            Code = 2
	InvalidArgument    Code = 3
	DeadlineExceeded   Code = 4
	NotFound           Code = 5
	AlreadyExists      Code = 6
	PermissionDenied   Code = 7
	ResourceExhausted  Code = 8
	FailedPrecondition Code = 9
	Aborted            Code = 10
	OutOfRange         Code = 11
	Unimplemented      Code = 12
	Internal           Code = 13
	Unavailable        Code = 14
	DataLoss           Code = 15
	Unauthenticated    Code = 16
)

// Valid reports whether c is a defined gRPC status code (0–16 inclusive).
func (c Code) Valid() bool {
	return c <= Unauthenticated
}

// codeNames is the single source of truth for Code string representations.
// Values use the canonical google.rpc.Code names.
var codeNames = [17]string{
	"OK",
	"CANCELLED",
	"UNKNOWN",
	"INVALID_ARGUMENT",
	"DEADLINE_EXCEEDED",
	"NOT_FOUND",
	"ALREADY_EXISTS",
	"PERMISSION_DENIED",
	"RESOURCE_EXHAUSTED",
	"FAILED_PRECONDITION",
	"ABORTED",
	"OUT_OF_RANGE",
	"UNIMPLEMENTED",
	"INTERNAL",
	"UNAVAILABLE",
	"DATA_LOSS",
	"UNAUTHENTICATED",
}

// codesByName maps the SCREAMING_SNAKE_CASE string back to a Code value.
// Populated by init from [codeNames] so the mapping is defined once.
var codesByName map[string]Code

func init() {
	codesByName = make(map[string]Code, len(codeNames))
	for i, name := range codeNames {
		codesByName[name] = Code(i)
	}
}

func (c Code) String() string {
	if c.Valid() {
		return codeNames[c]
	}
	return "Code(" + strconv.FormatInt(int64(c), 10) + ")"
}

// MarshalText implements [encoding.TextMarshaler].
func (c Code) MarshalText() ([]byte, error) {
	if !c.Valid() {
		return nil, fmt.Errorf("errdetails: cannot text-marshal invalid code %d", uint32(c))
	}
	return []byte(codeNames[c]), nil
}

// UnmarshalText implements [encoding.TextUnmarshaler].
func (c *Code) UnmarshalText(text []byte) error {
	v, ok := codesByName[string(text)]
	if !ok {
		return fmt.Errorf("errdetails: unknown code %q", string(text))
	}
	*c = v
	return nil
}

// MarshalJSON implements [json.Marshaler]. It encodes Code as a
// SCREAMING_SNAKE_CASE JSON string (e.g. "NOT_FOUND"), matching the
// canonical google.rpc.Code proto enum names.
func (c Code) MarshalJSON() ([]byte, error) {
	if !c.Valid() {
		return nil, fmt.Errorf("errdetails: cannot JSON-marshal invalid code %d", uint32(c))
	}
	return json.Marshal(codeNames[c])
}

// UnmarshalJSON implements [json.Unmarshaler]. It accepts the
// SCREAMING_SNAKE_CASE string form (e.g. "NOT_FOUND").
func (c *Code) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("errdetails: cannot unmarshal code: expected string")
	}
	v, ok := codesByName[s]
	if !ok {
		return fmt.Errorf("errdetails: unknown JSON code %q", s)
	}
	*c = v
	return nil
}
