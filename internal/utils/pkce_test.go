package utils_test

import (
	"testing"

	"insighta/internal/utils"
)

func TestGenerateState(t *testing.T) {
	s1 := utils.GenerateState()
	s2 := utils.GenerateState()

	if s1 == "" {
		t.Fatal("expected non-empty state")
	}
	if s1 == s2 {
		t.Error("expected unique states")
	}
}

func TestGenerateVerifier(t *testing.T) {
	v := utils.GenerateVerifier()
	if len(v) < 40 {
		t.Errorf("verifier too short: %d chars", len(v))
	}
}

func TestGenerateChallenge(t *testing.T) {
	verifier := utils.GenerateVerifier()
	challenge := utils.GenerateChallenge(verifier)

	if challenge == "" {
		t.Fatal("expected non-empty challenge")
	}
	if challenge == verifier {
		t.Error("challenge should differ from verifier")
	}

	// Same verifier → same challenge (deterministic)
	challenge2 := utils.GenerateChallenge(verifier)
	if challenge != challenge2 {
		t.Error("challenge should be deterministic")
	}
}
