/*
Copyright 2025 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package secret

import (
	"testing"

	prowapi "sigs.k8s.io/prow/pkg/apis/prowjobs/v1"
	"sigs.k8s.io/prow/pkg/secretutil"
)

func TestUpdateCensoringConfig(t *testing.T) {
	// Create a test agent
	agent := &agent{
		secretsMap:        make(map[string]secretReloader),
		ReloadingCensorer: secretutil.NewCensorer(),
	}

	// Test with nil config - should not panic
	agent.UpdateCensoringConfig(nil)

	// Test with config with minimum secret length
	minLength := 5
	config := &prowapi.CensoringOptions{
		MinimumSecretLength: &minLength,
	}

	agent.UpdateCensoringConfig(config)

	// Verify the censorer was updated (we can't directly test the minimum length
	// without accessing private fields, but we can test that it doesn't panic and
	// that secrets are loaded correctly)
	testSecrets := [][]byte{
		[]byte("short"),          // 5 chars, should be censored with minLength=5
		[]byte("verylongsecret"), // longer, should be censored
		[]byte("ab"),             // 2 chars, should not be censored with minLength=5
	}

	agent.ReloadingCensorer.RefreshBytes(testSecrets...)

	// Test censoring behavior
	input := []byte("This contains short and verylongsecret and ab")
	agent.ReloadingCensorer.Censor(&input)

	// With minimum length 5, "short" and "verylongsecret" should be censored, but "ab" should not
	result := string(input)

	if result == "This contains short and verylongsecret and ab" {
		t.Error("Expected some censoring to occur")
	}

	// "ab" should not be censored since it's shorter than minimum length
	if !containsString(result, "ab") {
		t.Error("Expected 'ab' to remain uncensored due to minimum length requirement")
	}
}

func TestUpdateCensoringConfigFromDecorationConfig(t *testing.T) {
	// Test with nil config - should not panic
	UpdateCensoringConfigFromDecorationConfig(nil)

	// Test with decoration config without censoring options
	decorationConfig := &prowapi.DecorationConfig{}
	UpdateCensoringConfigFromDecorationConfig(decorationConfig)

	// Test with decoration config with censoring options
	minLength := 3
	decorationConfig = &prowapi.DecorationConfig{
		CensoringOptions: &prowapi.CensoringOptions{
			MinimumSecretLength: &minLength,
		},
	}
	UpdateCensoringConfigFromDecorationConfig(decorationConfig)

	// Should not panic and should complete successfully
}

func TestApplyDecorationCensoringConfig(t *testing.T) {
	// Test with nil config - should not panic
	ApplyDecorationCensoringConfig(nil)

	// Test with valid decoration config
	minLength := 4
	decorationConfig := &prowapi.DecorationConfig{
		CensorSecrets: &[]bool{true}[0],
		CensoringOptions: &prowapi.CensoringOptions{
			MinimumSecretLength: &minLength,
		},
	}

	ApplyDecorationCensoringConfig(decorationConfig)

	// Should complete without panic
}

// containsString checks if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsString(s[1:], substr))))
}
