package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/kool/pkgs/release"
)

func TestLoadUploadCredentialsLocalFileWins(t *testing.T) {
	main := initRepo(t)
	linked := addWorktree(t, main)
	writeCreds(t, main, "main-token", "main-owner", "main-repo")
	writeCreds(t, linked, "local-token", "local-owner", "local-repo")
	t.Chdir(linked)

	creds, err := loadUploadCredentials()
	if err != nil {
		t.Fatal(err)
	}
	assertCreds(t, creds, "local-token", "local-owner", "local-repo")
}

func TestLoadUploadCredentialsFallsBackToMainCheckout(t *testing.T) {
	main := initRepo(t)
	linked := addWorktree(t, main)
	writeCreds(t, main, "main-token", "main-owner", "main-repo")
	t.Chdir(linked)

	creds, err := loadUploadCredentials()
	if err != nil {
		t.Fatal(err)
	}
	assertCreds(t, creds, "main-token", "main-owner", "main-repo")
}

func TestLoadUploadCredentialsMainCheckoutUsesItsFile(t *testing.T) {
	main := initRepo(t)
	writeCreds(t, main, "main-token", "main-owner", "main-repo")
	t.Chdir(main)

	creds, err := loadUploadCredentials()
	if err != nil {
		t.Fatal(err)
	}
	assertCreds(t, creds, "main-token", "main-owner", "main-repo")
}

func TestLoadUploadCredentialsInvalidLocalDoesNotFallBack(t *testing.T) {
	main := initRepo(t)
	linked := addWorktree(t, main)
	writeCreds(t, main, "main-token", "main-owner", "main-repo")
	if err := os.WriteFile(filepath.Join(linked, uploadCredentialsFile), []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(linked)

	_, err := loadUploadCredentials()
	if err == nil || !strings.Contains(err.Error(), "not valid JSON") {
		t.Fatalf("error = %v", err)
	}
	if strings.Contains(err.Error(), "main-owner") {
		t.Fatal("fell back to the main checkout")
	}
}

func TestLoadUploadCredentialsInvalidFallbackReportsThatFile(t *testing.T) {
	main := initRepo(t)
	linked := addWorktree(t, main)
	if err := os.WriteFile(filepath.Join(main, uploadCredentialsFile), []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(linked)

	_, err := loadUploadCredentials()
	if err == nil || !strings.Contains(err.Error(), "not valid JSON") || !strings.Contains(err.Error(), main) {
		t.Fatalf("error = %v", err)
	}
}

func TestLoadUploadCredentialsMissingInBothCheckouts(t *testing.T) {
	main := initRepo(t)
	linked := addWorktree(t, main)
	t.Chdir(linked)

	_, err := loadUploadCredentials()
	assertCredentialsNotFound(t, err)
	if strings.Contains(err.Error(), main) {
		t.Fatalf("error names the main checkout: %v", err)
	}
}

func TestLoadUploadCredentialsMainCheckoutMissing(t *testing.T) {
	main := initRepo(t)
	t.Chdir(main)

	_, err := loadUploadCredentials()
	assertCredentialsNotFound(t, err)
}

func TestLoadUploadCredentialsOutsideGit(t *testing.T) {
	t.Chdir(t.TempDir())

	_, err := loadUploadCredentials()
	assertCredentialsNotFound(t, err)
}

func TestReleaseHelpMentionsWorktreeCredentials(t *testing.T) {
	const sentence = "A linked worktree uses the main checkout's file when this checkout has none."
	if !strings.Contains(help, sentence) {
		t.Fatal("help missing worktree credentials sentence")
	}
}

func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	git(t, dir, "init", "-b", "master")
	git(t, dir, "config", "user.email", "release-test@example.com")
	git(t, dir, "config", "user.name", "release-test")
	git(t, dir, "config", "commit.gpgsign", "false")
	if err := os.WriteFile(filepath.Join(dir, "README"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, dir, "add", "README")
	git(t, dir, "commit", "-m", "init")
	return dir
}

func addWorktree(t *testing.T, main string) string {
	t.Helper()
	linked := filepath.Join(t.TempDir(), "linked")
	git(t, main, "worktree", "add", "--detach", linked)
	return linked
}

func writeCreds(t *testing.T, dir, token, owner, repo string) {
	t.Helper()
	body := `{"token":"` + token + `","owner":"` + owner + `","repo":"` + repo + `"}` + "\n"
	if err := os.WriteFile(filepath.Join(dir, uploadCredentialsFile), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func assertCreds(t *testing.T, creds *release.Credentials, token, owner, repo string) {
	t.Helper()
	if creds == nil {
		t.Fatal("credentials are nil")
	}
	if creds.Token != token || creds.Owner != owner || creds.Repo != repo {
		t.Fatalf("credentials = token %q owner %q repo %q", creds.Token, creds.Owner, creds.Repo)
	}
}

func assertCredentialsNotFound(t *testing.T, err error) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), ".upload-credentials.json not found") {
		t.Fatalf("error = %v", err)
	}
}

func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=release-test",
		"GIT_AUTHOR_EMAIL=release-test@example.com",
		"GIT_COMMITTER_NAME=release-test",
		"GIT_COMMITTER_EMAIL=release-test@example.com",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}
