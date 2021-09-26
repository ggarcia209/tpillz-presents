package hashops

import (
	"testing"
)

func TestGetMD5Hash(t *testing.T) {
	var tests = []string{
		"test001",
		"this is a test",
		"th1s is a test",
		"this is A test",
		"this is test",
	}
	seen := make(map[string]bool)

	for _, test := range tests {
		hash := GetMD5Hash(test)
		if seen[hash] {
			t.Errorf("Hash %v collided!", hash)
		} else {
			seen[hash] = true
		}
		t.Log(hash)
	}
}

func TestGetMD5Hash64Bit(t *testing.T) {
	var tests = []string{
		"test001",
		"this is a test",
		"th1s is a test",
		"this is A test",
		"this is test",
	}
	seen := make(map[string]bool)

	for _, test := range tests {
		hash := GetMD5Hash64Bit(test)
		if seen[hash] {
			t.Errorf("Hash %v collided!", hash)
		} else {
			seen[hash] = true
		}
		t.Log(hash)
	}
}
