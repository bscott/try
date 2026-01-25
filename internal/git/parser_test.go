package git

import (
	"testing"
)

func TestParseGitURI(t *testing.T) {
	tests := []struct {
		name        string
		uri         string
		wantHost    string
		wantOwner   string
		wantRepo    string
		wantErr     bool
		errContains string
	}{
		// HTTPS formats
		{
			name:      "HTTPS with .git",
			uri:       "https://github.com/user/repo.git",
			wantHost:  "github.com",
			wantOwner: "user",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:      "HTTPS without .git",
			uri:       "https://github.com/user/repo",
			wantHost:  "github.com",
			wantOwner: "user",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:      "HTTPS GitLab with .git",
			uri:       "https://gitlab.com/org/project.git",
			wantHost:  "gitlab.com",
			wantOwner: "org",
			wantRepo:  "project",
			wantErr:   false,
		},
		{
			name:      "HTTPS GitLab without .git",
			uri:       "https://gitlab.com/org/project",
			wantHost:  "gitlab.com",
			wantOwner: "org",
			wantRepo:  "project",
			wantErr:   false,
		},
		{
			name:      "HTTPS custom host",
			uri:       "https://git.example.com/team/service.git",
			wantHost:  "git.example.com",
			wantOwner: "team",
			wantRepo:  "service",
			wantErr:   false,
		},

		// SSH formats
		{
			name:      "SSH with .git",
			uri:       "git@github.com:user/repo.git",
			wantHost:  "github.com",
			wantOwner: "user",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:      "SSH without .git",
			uri:       "git@github.com:user/repo",
			wantHost:  "github.com",
			wantOwner: "user",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:      "SSH GitLab with .git",
			uri:       "git@gitlab.com:org/project.git",
			wantHost:  "gitlab.com",
			wantOwner: "org",
			wantRepo:  "project",
			wantErr:   false,
		},
		{
			name:      "SSH GitLab without .git",
			uri:       "git@gitlab.com:org/project",
			wantHost:  "gitlab.com",
			wantOwner: "org",
			wantRepo:  "project",
			wantErr:   false,
		},
		{
			name:      "SSH custom host",
			uri:       "git@git.example.com:team/service.git",
			wantHost:  "git.example.com",
			wantOwner: "team",
			wantRepo:  "service",
			wantErr:   false,
		},

		// SSH protocol prefix
		{
			name:      "SSH protocol with .git",
			uri:       "ssh://git@github.com:user/repo.git",
			wantHost:  "github.com",
			wantOwner: "user",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:      "SSH protocol without .git",
			uri:       "ssh://git@github.com:user/repo",
			wantHost:  "github.com",
			wantOwner: "user",
			wantRepo:  "repo",
			wantErr:   false,
		},

		// HTTP (not HTTPS)
		{
			name:      "HTTP with .git",
			uri:       "http://github.com/user/repo.git",
			wantHost:  "github.com",
			wantOwner: "user",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:      "HTTP without .git",
			uri:       "http://github.com/user/repo",
			wantHost:  "github.com",
			wantOwner: "user",
			wantRepo:  "repo",
			wantErr:   false,
		},

		// Edge cases and errors
		{
			name:        "Empty URI",
			uri:         "",
			wantErr:     true,
			errContains: "empty URI",
		},
		{
			name:        "Invalid format - no owner/repo",
			uri:         "https://github.com/",
			wantErr:     true,
			errContains: "invalid repository path format",
		},
		{
			name:        "Invalid format - only owner",
			uri:         "https://github.com/user",
			wantErr:     true,
			errContains: "invalid repository path format",
		},
		{
			name:        "Invalid format - malformed URL",
			uri:         "not-a-url",
			wantErr:     true,
			errContains: "no host found",
		},
		{
			name:      "Complex repo name with dash",
			uri:       "https://github.com/user/my-awesome-repo.git",
			wantHost:  "github.com",
			wantOwner: "user",
			wantRepo:  "my-awesome-repo",
			wantErr:   false,
		},
		{
			name:      "Complex owner with dash",
			uri:       "git@github.com:my-org/repo.git",
			wantHost:  "github.com",
			wantOwner: "my-org",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:      "Numeric repo name",
			uri:       "https://github.com/user/123.git",
			wantHost:  "github.com",
			wantOwner: "user",
			wantRepo:  "123",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseGitURI(tt.uri)

			// Check error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseGitURI() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.errContains != "" && err != nil {
					if !contains(err.Error(), tt.errContains) {
						t.Errorf("ParseGitURI() error = %v, should contain %v", err, tt.errContains)
					}
				}
				return
			}

			// Check parsed values
			if got.Host != tt.wantHost {
				t.Errorf("ParseGitURI() Host = %v, want %v", got.Host, tt.wantHost)
			}
			if got.Owner != tt.wantOwner {
				t.Errorf("ParseGitURI() Owner = %v, want %v", got.Owner, tt.wantOwner)
			}
			if got.Repo != tt.wantRepo {
				t.Errorf("ParseGitURI() Repo = %v, want %v", got.Repo, tt.wantRepo)
			}
			if got.Original != tt.uri {
				t.Errorf("ParseGitURI() Original = %v, want %v", got.Original, tt.uri)
			}
		})
	}
}

func TestRepoInfo_String(t *testing.T) {
	info := &RepoInfo{
		Host:  "github.com",
		Owner: "user",
		Repo:  "repo",
	}

	want := "github.com/user/repo"
	if got := info.String(); got != want {
		t.Errorf("RepoInfo.String() = %v, want %v", got, want)
	}
}

func TestRepoInfo_CloneURL(t *testing.T) {
	info := &RepoInfo{
		Host:  "github.com",
		Owner: "user",
		Repo:  "repo",
	}

	want := "https://github.com/user/repo.git"
	if got := info.CloneURL(); got != want {
		t.Errorf("RepoInfo.CloneURL() = %v, want %v", got, want)
	}
}

func TestRepoInfo_SSHURL(t *testing.T) {
	info := &RepoInfo{
		Host:  "github.com",
		Owner: "user",
		Repo:  "repo",
	}

	want := "git@github.com:user/repo.git"
	if got := info.SSHURL(); got != want {
		t.Errorf("RepoInfo.SSHURL() = %v, want %v", got, want)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && indexOf(s, substr) >= 0))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
