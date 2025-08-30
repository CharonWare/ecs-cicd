package ci

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// getHeadHash returns the HEAD commit hash of a local repo
func GetHeadHash(directory string) (string, error) {
	cmd := exec.Command("git", "-C", directory, "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD hash: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// getRemoteHash gets the latest commit hash of a remote branch
func GetRemoteHash(project, branch string) (string, error) {
	cmd := exec.Command("git", "ls-remote", project, branch)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get remote hash: %w", err)
	}
	parts := strings.Fields(string(out))
	if len(parts) == 0 {
		return "", fmt.Errorf("no remote hash found for %s", branch)
	}
	return parts[0], nil
}

func DockerBuild(directory, ecr string) (string, error) {
	err := os.Chdir(directory)
	if err != nil {
		return "", fmt.Errorf("error changing directory: %w", err)
	}

	// Pull the latest from the repo
	cmd := exec.Command("git", "pull")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git pull failed: %w", err)
	}

	// Build from the Dockerfile
	// Tag with both a timestamp and as latest
	timestamp := time.Now().Format("2006-01-02t150405")
	versionTag := ecr + ":" + timestamp
	latestTag := ecr + ":latest"
	cmd = exec.Command("docker", "build", "-t", versionTag, "-t", latestTag, ".")
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("docker build failed: %w", err)
	}

	// Update the last_commit file with the new hash
	newHash, err := GetHeadHash(".")
	if err != nil {
		return "", fmt.Errorf("unable to get HEAD hash: %w", err)
	}

	if err := os.WriteFile(".last_commit", []byte(newHash), 0644); err != nil {
		return "", fmt.Errorf("failed to write state file: %w", err)
	}

	return versionTag, nil
}

func PushToEcr(ecr, versionTag, region string) error {
	cmdGetPw := exec.Command("aws", "ecr", "get-login-password", "--region", region)

	cmdLogin := exec.Command("docker", "login", "--username", "AWS", "--password-stdin", ecr)
	cmdLogin.Stdin, _ = cmdGetPw.StdoutPipe()
	cmdLogin.Stdout = os.Stdout
	cmdLogin.Stderr = os.Stderr

	if err := cmdLogin.Start(); err != nil {
		return fmt.Errorf("failed to start docker login: %w", err)
	}
	if err := cmdGetPw.Run(); err != nil {
		return fmt.Errorf("failed to get ECR login password: %w", err)
	}
	if err := cmdLogin.Wait(); err != nil {
		return fmt.Errorf("docker login failed: %w", err)
	}

	cmd := exec.Command("docker", "push", versionTag)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker push failed: %w", err)
	}

	return nil
}
