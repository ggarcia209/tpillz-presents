package timeops

import (
	"fmt"
	"log"
	"strconv"
	"time"
)

// ConvertToTimeStamp string converts the given Time value to a
// string in MM/DD/YYYY HH:MM:SS.MS format.
func ConvertToTimestampString(t time.Time) string {
	y, m, d := t.Date()
	hr, min, sec := t.Hour(), t.Minute(), t.Second()
	vals := []int{int(m), d, y, hr, min, sec}
	strs := []string{}
	for _, val := range vals {
		str := strconv.Itoa(val)
		if val < 10 {
			// add leading zero
			str = "0" + str
		}
		strs = append(strs, str)
	}
	fmt := fmt.Sprintf("%s-%s-%s %s:%s:%s", strs[0], strs[1], strs[2], strs[3], strs[4], strs[5])
	return fmt
}

// ConvertStringToTimestamp converts timestamp string in 'MM-DD-YYYY HH:MM:SS' format
// to a time.Time object
func ConvertStringToTimestamp(s string) (time.Time, error) {
	fmt := "01-02-2006 15:04:05"
	time, err := time.Parse(fmt, s)
	if err != nil {
		log.Printf("ConvertStringToTimestamp failed: %v", err)
		return time, err
	}
	return time, nil
}

// Convert time.Time object to date string in 'MM-DD-YYYY' format
func ConvertToDateString(t time.Time) string {
	y, m, d := t.Date()
	vals := []int{int(m), d, y}
	strs := []string{}
	for _, val := range vals {
		str := strconv.Itoa(val)
		if val < 10 {
			// add leading zero
			str = "0" + str
		}
		strs = append(strs, str)
	}
	fmt := fmt.Sprintf("%s-%s-%s", strs[0], strs[1], strs[2])
	return fmt
}

func ConvertDateStringToTime(s string) (time.Time, error) {
	fmt := "01-02-2006"
	time, err := time.Parse(fmt, s)
	if err != nil {
		log.Printf("ConvertDateStringToTimefailed: %v", err)
		return time, err
	}
	return time, nil
}
