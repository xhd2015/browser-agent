package main

import (
	"strings"
	"time"
)

const slowThreshold = 8 * time.Second

// classifyEval maps an identity eval result into a ResultClass.
func classifyEval(targetURL, evalHref string, duration time.Duration, err error) ResultClass {
	if err != nil {
		return classifyError(err)
	}
	href := strings.TrimSpace(evalHref)
	if isSessionControlURL(href) && !isSessionControlURL(targetURL) {
		return classFailRouting
	}
	if looksLikeAuthWall(href) {
		return classOKAuthWall
	}
	if duration >= slowThreshold {
		return classOKSlow
	}
	return classOKLoaded
}

func classifyError(err error) ResultClass {
	if err == nil {
		return classFailOther
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "timeout") || strings.Contains(msg, "deadline exceeded") || strings.Contains(msg, "context deadline"):
		return classFailTimeout
	case strings.Contains(msg, "extension not connected") || strings.Contains(msg, "not connected") || strings.Contains(msg, "websocket"):
		return classFailDisconnect
	case strings.Contains(msg, "attach") || strings.Contains(msg, "debugger") || strings.Contains(msg, "not allowed") || strings.Contains(msg, "-32000"):
		return classFailAttach
	case strings.Contains(msg, "crash") || strings.Contains(msg, "connection refused") || strings.Contains(msg, "broken pipe"):
		return classFailCrash
	default:
		return classFailOther
	}
}

func severityFor(c ResultClass) Severity {
	switch c {
	case classFailCrash:
		return sevP0
	case classFailRouting, classFailTimeout, classFailDisconnect:
		return sevP1
	case classFailAttach, classFailOther:
		return sevP2
	default:
		return sevP3
	}
}

func isP0orP1(c ResultClass) bool {
	s := severityFor(c)
	return s == sevP0 || s == sevP1
}

func isSessionControlURL(u string) bool {
	u = strings.ToLower(strings.TrimSpace(u))
	if u == "" {
		return false
	}
	return strings.Contains(u, "/go?session=") ||
		strings.Contains(u, "/go?session%3d") ||
		(strings.Contains(u, "127.0.0.1:43761") && strings.Contains(u, "session=")) ||
		(strings.Contains(u, "localhost:43761") && strings.Contains(u, "session="))
}

func looksLikeAuthWall(href string) bool {
	h := strings.ToLower(href)
	needles := []string{
		"login", "sso", "signin", "sign-in", "auth", "oauth",
		"accounts.google", "okta", "cas.", "passport", "idp.",
		"saml", "openid", "keycloak",
	}
	for _, n := range needles {
		if strings.Contains(h, n) {
			return true
		}
	}
	return false
}

func expectedFor(c ResultClass, op OpName) string {
	switch c {
	case classFailRouting:
		return "eval/screenshot targets the content tab, not the /go session page"
	case classFailTimeout:
		return "job completes within timeout with a structured error if page is slow"
	case classFailDisconnect:
		return "extension stays connected while /go session page remains open"
	case classFailAttach:
		return "debugger attach succeeds on capturable http(s) tabs"
	case classFailCrash:
		return "daemon and extension remain alive"
	default:
		return "op " + string(op) + " completes without agent reliability failure"
	}
}
