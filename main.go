package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/CharonWare/ecs-cicd/internal/ci"
)

func main() {
	// All env vars are required, fail fast if they are empty
	project := os.Getenv("PROJECT")
	branch := os.Getenv("BRANCH")
	token := os.Getenv("PAT_TOKEN")
	ecr := os.Getenv("ECR")
	region := os.Getenv("AWS_DEFAULT_REGION")

	if project == "" || token == "" || ecr == "" || region == "" {
		log.Fatal("Missing required environment variables: PROJECT, BRANCH, PAT_TOKEN, ECR, AWS_DEFAULT_REGION")
	}
	if branch == "" {
		branch = "main"
	}
	// Create the authorised URL in order to clone/pull the repo using the PAT token
	authURL := "https://" + token + "@github.com/" + project

	// Set up the repos directory
	if err := os.Mkdir("repos", 0775); err != nil {
		fmt.Errorf("error creating the repos directory: %w", err)
	}
	os.Chdir("repos")

	// Use the project var to determine what the name of the directory will be
	ci_dir, err := getDirectory(project)
	if err != nil {
		fmt.Errorf("unable to get directory using PROJECT var: %w", err)
	}

	// Check if the directory exists, if not, create it by cloning
	buildRequired, err := checkOrClone(ci_dir, authURL, branch)
	if err != nil {
		fmt.Errorf("unable to check for updates or clone git repository: %w", err)
	}

	// Check the local hash against the remote hash, if they match then buildRequired = false
	// If they don't match, initiate the CI/CD process
	if !buildRequired {
		fmt.Print("No build required")
	} else {
		fmt.Print("Initiating new image build: ")
		versionTag, err := ci.DockerBuild(ci_dir, ecr)
		if err != nil {
			fmt.Printf("CI process failed: %v\n", err)
		} else {
			fmt.Printf("build result: %s\n", versionTag)
		}

		if err := ci.PushToEcr(ecr, versionTag, region); err != nil {
			fmt.Errorf("%w", err)
		} else {
			fmt.Println("Successfully pushed to ECR")
		}
	}
}

// The project env var is [username || organisation]/repository, we just want repository for the directory name
func getDirectory(project string) (string, error) {
	split_proj := strings.Split(project, "/")
	dir := strings.TrimSuffix(split_proj[1], ".git")

	return dir, nil
}

// Create the repository if it doesn't exist, this will trigger a build
// If it already exists, check for new commits
func checkOrClone(directory, authURL, branch string) (bool, error) {
	stateFile := directory + "/.last_commit"

	buildRequired := false
	// If repo doesn't exist, clone it
	if _, err := os.Stat(directory); errors.Is(err, os.ErrNotExist) {
		cmd := exec.Command("git", "clone", "--branch", branch, authURL, directory)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return buildRequired, fmt.Errorf("git clone failed: %w", err)
		}

		// Get the HEAD commit hash
		commitHash, err := ci.GetHeadHash(directory)
		if err != nil {
			return buildRequired, err
		}

		// Create the state file with the commit hash
		if err := os.WriteFile(stateFile, []byte(commitHash), 0644); err != nil {
			return buildRequired, fmt.Errorf("failed to write state file: %w", err)
		}

		// If the repo needed to be cloned then a build is required
		buildRequired = true
		return buildRequired, nil
	}

	// If repo does exit, fetch new commits
	cmd := exec.Command("git", "-C", directory, "fetch", "origin", branch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return buildRequired, fmt.Errorf("git fetch failed: %w", err)
	}

	// Get remote branch HEAD
	remoteHash, err := ci.GetRemoteHash(authURL, branch)
	if err != nil {
		return buildRequired, err
	}

	// Compare the hashes, if they are the same, nothing is required
	// If they are different, a build is required, set bool to true and return
	var storedHash string
	if data, err := os.ReadFile(stateFile); err == nil {
		storedHash = strings.TrimSpace(string(data))
	}

	if remoteHash == storedHash {
		return buildRequired, nil
	} else {
		buildRequired = true
		return buildRequired, nil
	}
}
