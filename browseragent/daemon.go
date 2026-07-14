package browseragent

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// DaemonMeta is written to {BaseDir}/server.json for daemon discovery.
type DaemonMeta struct {
	PID           int       `json:"pid"`
	Addr          string    `json:"addr"`
	BaseURL       string    `json:"base_url"`
	BaseDir       string    `json:"base_dir"`
	StartedAt     time.Time `json:"started_at"`
	DaemonVersion string    `json:"daemon_version,omitempty"`
}

var sessionIDPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]{0,63}$`)

const generateSessionIDChars = "abcdefghijklmnopqrstuvwxyz0123456789"

// ValidateSessionID returns nil when id matches the allowed session id rules.
func ValidateSessionID(id string) error {
	if id == "" {
		return errors.New("session id is empty")
	}
	if strings.Contains(id, "..") {
		return fmt.Errorf("session id %q contains invalid sequence \"..\"", id)
	}
	if !sessionIDPattern.MatchString(id) {
		return fmt.Errorf("session id %q is invalid", id)
	}
	return nil
}

// GenerateSessionID returns sess- followed by 6 random lowercase alphanumeric characters.
func GenerateSessionID() string {
	suffix := make([]byte, 6)
	randBytes := make([]byte, 6)
	if _, err := rand.Read(randBytes); err != nil {
		panic(fmt.Errorf("generate session id: %w", err))
	}
	for i := range suffix {
		suffix[i] = generateSessionIDChars[int(randBytes[i])%len(generateSessionIDChars)]
	}
	return "sess-" + string(suffix)
}

// WriteDaemonMeta atomically writes meta as JSON with a trailing newline.
func WriteDaemonMeta(path string, meta DaemonMeta) error {
	data, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("marshal daemon meta: %w", err)
	}
	data = append(data, '\n')

	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".server.json.*")
	if err != nil {
		return fmt.Errorf("create temp daemon meta file: %w", err)
	}
	tmpPath := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write daemon meta: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close daemon meta temp file: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("rename daemon meta file: %w", err)
	}
	cleanup = false
	return nil
}

// ReadDaemonMeta parses daemon meta from path.
func ReadDaemonMeta(path string) (DaemonMeta, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return DaemonMeta{}, fmt.Errorf("read daemon meta: %w", err)
	}
	var meta DaemonMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return DaemonMeta{}, fmt.Errorf("parse daemon meta: %w", err)
	}
	return meta, nil
}

// RemoveDaemonMeta removes the daemon meta file. Missing files are not an error.
func RemoveDaemonMeta(path string) error {
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove daemon meta: %w", err)
	}
	return nil
}

// SessionDirPath returns the on-disk path for a session directory.
func SessionDirPath(baseDir, sessionID string) string {
	return filepath.Join(baseDir, "sessions", sessionID)
}

// SessionDirExists reports whether the session directory exists.
func SessionDirExists(baseDir, sessionID string) bool {
	info, err := os.Stat(SessionDirPath(baseDir, sessionID))
	if err != nil {
		return false
	}
	return info.IsDir()
}