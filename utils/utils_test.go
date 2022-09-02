package utils_test

import (
	"iex-indicators/utils"
	"os"
	"testing"
)

var key = "ENV_KEY"
var value = "ENV_VALUE"

func TestMain(m *testing.M) {
	// log.Println("Do stuff BEFORE the tests!")
	os.Setenv(key, value)
	exitVal := m.Run()
	// log.Println("Do stuff AFTER the tests!")

	os.Exit(exitVal)
}

func TestGetEnv(t *testing.T) {

	var badKey = "XKDNVMCKAW"
	var goodValue = "good"

	myValue := utils.GetEnv(key, "bad")

	if myValue != value {
		t.Error("Bad Value found")
	}

	myValue = utils.GetEnv(badKey, goodValue)

	if myValue != goodValue {
		t.Error("Good Value not found")
	}
}

func TestJulDate(t *testing.T) {

	t.Logf("%s", utils.JulDate())
}