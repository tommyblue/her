package main

import (
	"fmt"
	"testing"
)

func Test_loadConfig(t *testing.T) {
	err := loadConfig("unexistent")
	if err.Error() != fmt.Errorf("error opening the config file").Error() {
		t.Error("Expected error")
	}

	err = loadConfig("../../config.example.toml")
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
}
