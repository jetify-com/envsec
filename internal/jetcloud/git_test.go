package jetcloud

import (
	"testing"
)

func TestNormalizeGitRepoURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"git@github.com:username/repo.git", "github.com/username/repo"},
		{"https://github.com/username/repo.git", "github.com/username/repo"},
		{"http://github.com/username/repo.git", "github.com/username/repo"},
		{"https://www.github.com/username/repo.git", "github.com/username/repo"},
		{"http://www.github.com/username/repo.git", "github.com/username/repo"},
		{"git@github.com:username/repo", "github.com/username/repo"},
		{"https://github.com/username/repo", "github.com/username/repo"},
		{"http://github.com/username/repo", "github.com/username/repo"},
		{"https://www.github.com/username/repo", "github.com/username/repo"},
		{"http://www.github.com/username/repo", "github.com/username/repo"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := normalizeGitRepoURL(test.input)
			if result != test.expected {
				t.Errorf("Expected %s, but got %s", test.expected, result)
			}
		})
	}
}
