package timeops

import (
	"testing"
	"time"
)

func TestConvertToTimestampString(t *testing.T) {
	var tests = []struct {
		input time.Time
	}{
		{input: time.Now()},
		{input: time.Now()},
		{input: time.Now()},
	}
	for _, test := range tests {
		ts := ConvertToTimestampString(test.input)
		t.Logf("timestamp: %v <END>", ts)
	}
}

func TestStringToTimestamp(t *testing.T) {
	var tests = []struct {
		input string
		want  error
	}{
		{input: "01-02-2006 15:04:05", want: nil}, // PASS
		{input: "05-30-2021 13:22:24", want: nil}, // PASS
		{input: "5-30-2021 13:22:24", want: nil},  // FAIL
		{input: "2021-5-30 12:22:24", want: nil},  // FAIL
	}
	for _, test := range tests {
		ts, err := ConvertStringToTimestamp(test.input)
		if err != test.want {
			t.Errorf("FAIL: %v", err)
		} else {
			t.Logf("PASS - timestamp: %v", ts)
		}
	}
}

func TestConvertToDateString(t *testing.T) {
	var tests = []struct {
		input time.Time
		want  string
	}{
		{input: time.Now(), want: "05-30-2021"},
		{input: time.Now(), want: "05-30-2021"},
		{input: time.Now(), want: "05-30-2021"},
	}
	for _, test := range tests {
		ts := ConvertToDateString(test.input)
		if ts != test.want {
			t.Errorf("FAIL: %v; want: %v", ts, test.want)
		} else {
			t.Logf("timestamp: %v <END>", ts)
		}
	}
}

func TestConvertDateStringToTime(t *testing.T) {
	var tests = []struct {
		input string
		want  error
	}{
		{input: "01-02-2006", want: nil}, // PASS
		{input: "05-30-2021", want: nil}, // PASS
		{input: "5-30-2021", want: nil},  // FAIL
		{input: "2021-5-30", want: nil},  // FAIL
	}
	for _, test := range tests {
		ts, err := ConvertDateStringToTime(test.input)
		if err != test.want {
			t.Errorf("FAIL: %v", err)
		} else {
			t.Logf("PASS - timestamp: %v", ts)
		}
	}
}
