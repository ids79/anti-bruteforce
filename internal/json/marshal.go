package internaljson

import (
	"encoding/json"
	"strings"
	"time"
)

type JSONDate time.Time

type Duration time.Duration

func (j JSONDate) MarshalJSON() ([]byte, error) {
	st := time.Time(j).Format("2006-01-02 03:04 PM")
	return json.Marshal(st)
}

func (j *JSONDate) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	t, err := time.Parse("2006-01-02 03:04 PM", s)
	if err != nil {
		return err
	}
	*j = JSONDate(t)
	return nil
}

func (d Duration) MarshalJSON() ([]byte, error) {
	st := time.Duration(d).String()
	return json.Marshal(st)
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v string
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	t, err := time.ParseDuration(v)
	if err != nil {
		return err
	}
	*d = Duration(t)
	return nil
}
