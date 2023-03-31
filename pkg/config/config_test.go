// Copyright Contributors to the Open Cluster Management project

package config

import (
	"os"
	"testing"
)

// Test "new()"
func TestNew(t *testing.T) {
	type testConfig struct {
		ThanosURL string
		HttpPort  int
		KruizeURL string
	}
	testCase := struct {
		name      string
		ThanosURL string
		HttpPort  int
		KruizeURL string
		expected  testConfig
	}{
		"should execute \"new()\" and return a new config object instance",
		getEnv("THANOS_SERVER_URL", "https://localhost:5555"),
		getEnvAsInt("HTTP_PORT", 4020),
		getEnv("KRUIZE_URL", "http://localhost:8080"),
		testConfig{getEnv("THANOS_SERVER_URL", "https://localhost:5555"), getEnvAsInt("HTTP_PORT", 4020), getEnv("KRUIZE_URL", "http://localhost:8080")},
	}

	testCfg := new()

	if testCfg == nil {
		t.Errorf("case (%v) output: (%v) is not the expected value: (%v)", testCase.name, testCfg, testCase.expected)
	}

	if testCfg.ThanosURL != testCase.expected.ThanosURL {
		t.Errorf("case (%v) output: (%v) is not the expected value: (%v)", testCase.name, testCfg.ThanosURL, testCase.expected.ThanosURL)
	}

	if testCfg.HttpPort != testCase.expected.HttpPort {
		t.Errorf("case (%v) output: (%v) is not the expected value: (%v)", testCase.name, testCfg.HttpPort, testCase.expected.HttpPort)
	}

	if testCfg.KruizeURL != testCase.expected.KruizeURL {
		t.Errorf("case (%v) output: (%v) is not the expected value: (%v)", testCase.name, testCfg.KruizeURL, testCase.expected.KruizeURL)
	}
}

// Test getEnv
func TestGetEnv(t *testing.T) {
	testCase := struct {
		name       string
		val        string
		defaultVal string
		expected   string
	}{
		"should execute getEnv and return 3",
		"test-1",
		"",
		"test-1",
	}

	envVar := "THANOS_SERVER_URL"
	if val := getEnv(envVar, ""); val == Cfg.ThanosURL {
		t.Errorf("case (%v) output: (%v) is not the expected value: (%v)", testCase.name, val, "\"\"")
	}

	if err := os.Setenv(envVar, testCase.val); err != nil {
		t.Errorf("failed to set environment variable: %v", err)
	}

	if val := getEnv(envVar, ""); val != testCase.expected {
		t.Errorf("case (%v) output: (%v) is not the expected value: (%v)", testCase.name, val, testCase.expected)
	}
}

// Test getEnvAsInt
func TestGetEnvAsInt(t *testing.T) {
	testCases := []struct {
		name       string
		val        string
		defaultVal int
		expected   int
	}{
		{
			"should execute getEnvAsInt and return 3",
			"3",
			0,
			3,
		},
		{
			"should execute getEnvAsInt and return 3",
			"three",
			3,
			3,
		},
	}

	envVar := "NUM_OF_MANAGED_CLUSTERS"
	for _, testCase := range testCases {
		os.Setenv(envVar, testCase.val)
		if val := os.Getenv(envVar); val != testCase.val {
			t.Errorf("failed to set environment variable '%v' to '%v'", envVar, testCase.val)
		}

		if ok := getEnvAsInt(envVar, testCase.defaultVal); ok != testCase.expected {
			t.Logf("case (%v) output: (%v) is not the expected value: (%v)", testCase.name, ok, testCase.expected)
		}
	}
}

// Test getEnvAsBool
func TestGetEnvAsBool(t *testing.T) {
	testCases := []struct {
		name       string
		val        string
		defaultVal bool
		expected   bool
	}{
		{
			"should execute getEnvAsBool and return true",
			"true",
			false,
			true,
		},
		{
			"should execute getEnvAsBool and return true",
			"unknown",
			true,
			true,
		},
	}

	envVar := "IS_HUB_CLUSTER"
	for _, testCase := range testCases {
		os.Setenv(envVar, testCase.val)
		if val := os.Getenv(envVar); val != testCase.val {
			t.Errorf("failed to set environment variable '%v' to '%v'", envVar, testCase.val)
		}

		if ok := getEnvAsBool(envVar, testCase.defaultVal); !ok {
			t.Logf("case (%v) output: (%v) is not the expected value: (%v)", testCase.name, ok, testCase.expected)
		}
	}
}
