package git

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// RepoInfo contains parsed information from a Git repository URI
type RepoInfo struct {
	Host     string // github.com, gitlab.com, etc.
	Owner    string // user or org
	Repo     string // repository name
	Original string // original URI
}

// ParseGitURI parses a Git repository URI and extracts repository information.
// Supports multiple formats:
// - https://github.com/user/repo.git
// - https://github.com/user/repo (no .git)
// - git@github.com:user/repo.git
// - git@github.com:user/repo
func ParseGitURI(uri string) (*RepoInfo, error) {
	if uri == "" {
		return nil, fmt.Errorf("empty URI provided")
	}

	original := uri
	info := &RepoInfo{
		Original: original,
	}

	// Try SSH format first: git@host:owner/repo.git or git@host:owner/repo
	sshRegex := regexp.MustCompile(`^(?:git@|ssh://git@)([^:]+):(.+)/([^/]+?)(?:\.git)?$`)
	if matches := sshRegex.FindStringSubmatch(uri); matches != nil {
		info.Host = matches[1]
		info.Owner = matches[2]
		info.Repo = matches[3]
		return info, nil
	}

	// Try HTTPS/HTTP format
	parsedURL, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid URI format: %w", err)
	}

	// Extract host
	if parsedURL.Host == "" {
		return nil, fmt.Errorf("no host found in URI: %s", uri)
	}
	info.Host = parsedURL.Host

	// Extract owner and repo from path
	path := strings.TrimPrefix(parsedURL.Path, "/")
	path = strings.TrimSuffix(path, ".git")

	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid repository path format: expected owner/repo, got %s", path)
	}

	info.Owner = parts[0]
	info.Repo = parts[1]

	if info.Owner == "" || info.Repo == "" {
		return nil, fmt.Errorf("could not extract owner and repo from URI: %s", uri)
	}

	return info, nil
}

// String returns a human-readable representation of the RepoInfo
func (r *RepoInfo) String() string {
	return fmt.Sprintf("%s/%s/%s", r.Host, r.Owner, r.Repo)
}

// CloneURL returns the HTTPS clone URL for the repository
func (r *RepoInfo) CloneURL() string {
	return fmt.Sprintf("https://%s/%s/%s.git", r.Host, r.Owner, r.Repo)
}

// SSHURL returns the SSH clone URL for the repository
func (r *RepoInfo) SSHURL() string {
	return fmt.Sprintf("git@%s:%s/%s.git", r.Host, r.Owner, r.Repo)
}
