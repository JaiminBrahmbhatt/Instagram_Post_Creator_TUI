package db

import (
	"os"
	"testing"
)

func TestGetConfig(t *testing.T) {
	// Setup DB
	db, err := InitDB(":memory:")
	if err != nil {
		t.Fatal(err)
	}

	// 1. Test Default (Empty)
	val := db.GetConfig("non_existent", "")
	if val != "" {
		t.Errorf("Expected empty string, got %s", val)
	}

	// 2. Test Env Fallback
	os.Setenv("TEST_ENV_VAR", "env_value")
	val = db.GetConfig("test_key", "TEST_ENV_VAR")
	if val != "env_value" {
		t.Errorf("Expected 'env_value', got %s", val)
	}

	// 3. Test DB Override
	err = db.SetSetting("test_key", "db_value")
	if err != nil {
		t.Fatal(err)
	}
	val = db.GetConfig("test_key", "TEST_ENV_VAR")
	if val != "db_value" {
		t.Errorf("Expected 'db_value', got %s", val)
	}

	// Cleanup
	os.Unsetenv("TEST_ENV_VAR")
}

func TestGetConfigBool(t *testing.T) {
	db, _ := InitDB(":memory:")

	if db.GetConfigBool("bool_key", "") {
		t.Error("Expected false for missing key")
	}

	os.Setenv("BOOL_ENV", "true")
	if !db.GetConfigBool("bool_key", "BOOL_ENV") {
		t.Error("Expected true from env")
	}

	db.SetSetting("bool_key", "false")
	if db.GetConfigBool("bool_key", "BOOL_ENV") {
		t.Error("Expected false from DB override")
	}
	
	db.SetSetting("bool_key", "1")
	if !db.GetConfigBool("bool_key", "BOOL_ENV") {
		t.Error("Expected true from DB '1'")
	}
	
	os.Unsetenv("BOOL_ENV")
}
