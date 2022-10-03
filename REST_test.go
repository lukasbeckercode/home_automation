package main

import "testing"

func Test_getanalogpartbyname(t *testing.T) {
	result, err := getAnalogPartByName("TEMP1")
	if err != nil || result == nil {
		t.Error("Expected err to be nil, result to contain a part, "+
			"but got {} for err and {} for part", err, result)
	}
}
