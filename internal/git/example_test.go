package git_test

import (
	"fmt"
	"log"

	"github.com/bscott/try/internal/git"
)

func ExampleParseGitURI() {
	// Parse an HTTPS URL
	info, err := git.ParseGitURI("https://github.com/user/repo.git")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(info.Host)
	fmt.Println(info.Owner)
	fmt.Println(info.Repo)
	// Output:
	// github.com
	// user
	// repo
}

func ExampleParseGitURI_ssh() {
	// Parse an SSH URL
	info, err := git.ParseGitURI("git@github.com:user/repo.git")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(info.Host)
	fmt.Println(info.Owner)
	fmt.Println(info.Repo)
	// Output:
	// github.com
	// user
	// repo
}

func ExampleRepoInfo_String() {
	info := &git.RepoInfo{
		Host:  "github.com",
		Owner: "user",
		Repo:  "repo",
	}

	fmt.Println(info.String())
	// Output:
	// github.com/user/repo
}

func ExampleRepoInfo_CloneURL() {
	info := &git.RepoInfo{
		Host:  "github.com",
		Owner: "user",
		Repo:  "repo",
	}

	fmt.Println(info.CloneURL())
	// Output:
	// https://github.com/user/repo.git
}

func ExampleRepoInfo_SSHURL() {
	info := &git.RepoInfo{
		Host:  "github.com",
		Owner: "user",
		Repo:  "repo",
	}

	fmt.Println(info.SSHURL())
	// Output:
	// git@github.com:user/repo.git
}
