package flagutil

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
)

/* Additional flag implementations similar to the StringToString/StringToInt flag implementation in the pflag project. */

// -- stringToBool Value
type stringToBoolValue struct {
	value   *map[string]bool
	changed bool
}

func newStringToBoolValue(val map[string]bool, p *map[string]bool) *stringToBoolValue {
	ssv := new(stringToBoolValue)
	ssv.value = p
	*ssv.value = val
	return ssv
}

// Format: a=true,b=false,c
func (s *stringToBoolValue) Set(val string) error {
	ss := strings.Split(val, ",")
	out := make(map[string]bool, len(ss))
	for _, pair := range ss {
		// TODO: interpret omitted "=" as "true"
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			return fmt.Errorf("%s must be formatted as key=value", pair)
		}
		var err error
		out[kv[0]], err = strconv.ParseBool(kv[1])
		if err != nil {
			return fmt.Errorf("parsing value for %q: %w", kv[0], err)
		}
	}
	if !s.changed {
		*s.value = out
	} else {
		for k, v := range out {
			(*s.value)[k] = v
		}
	}
	s.changed = true
	return nil
}

func (s *stringToBoolValue) Type() string {
	return "stringToBool"
}

func (s *stringToBoolValue) String() string {
	var buf bytes.Buffer
	i := 0
	for k, v := range *s.value {
		if i > 0 {
			buf.WriteRune(',')
		}
		buf.WriteString(k)
		buf.WriteRune('=')
		buf.WriteString(strconv.FormatBool(v))
		i++
	}
	return "[" + buf.String() + "]"
}

// -- stringToOptString Value
type stringToOptStringValue struct {
	value   *map[string]*string
	changed bool
}

func newStringToOptStringValue(val map[string]*string, p *map[string]*string) *stringToOptStringValue {
	ssv := new(stringToOptStringValue)
	ssv.value = p
	*ssv.value = val
	return ssv
}

// Format: a,b=2
func (s *stringToOptStringValue) Set(val string) error {
	var ss []string
	r := csv.NewReader(strings.NewReader(val))
	var err error
	ss, err = r.Read()
	if err != nil {
		return err
	}

	out := make(map[string]*string, len(ss))
	for _, pair := range ss {
		k, v, found := strings.Cut(pair, "=")
		if !found {
			out[k] = nil
		} else {
			out[k] = &v
		}
	}
	if !s.changed {
		*s.value = out
	} else {
		for k, v := range out {
			(*s.value)[k] = v
		}
	}
	s.changed = true
	return nil
}

func (s *stringToOptStringValue) Type() string {
	return "stringToOptString"
}

func (s *stringToOptStringValue) String() string {
	records := make([]string, 0, len(*s.value)>>1)
	for k, v := range *s.value {
		if v != nil {
			records = append(records, k+"="+*v)
		} else {
			records = append(records, k)
		}
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.Write(records); err != nil {
		panic(err)
	}
	w.Flush()
	return "[" + strings.TrimSpace(buf.String()) + "]"
}
