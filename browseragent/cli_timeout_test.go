package browseragent

import "testing"

func TestJobTimeoutMS(t *testing.T) {
	const def int64 = 15000

	cases := []struct {
		name string
		args []string
		want int64
	}{
		{"default", nil, def},
		{"empty flag ignored", []string{"--timeout", ""}, def},
		{"ms integer", []string{"--timeout", "60000"}, 60000},
		{"equals form", []string{"--timeout=45000"}, 45000},
		{"duration seconds", []string{"--timeout", "30s"}, 30000},
		{"duration minute", []string{"--timeout", "1m"}, 60000},
		{"duration ms", []string{"--timeout", "500ms"}, 500},
		{"zero falls back", []string{"--timeout", "0"}, def},
		{"negative falls back", []string{"--timeout", "-1"}, def},
		{"invalid falls back", []string{"--timeout", "nope"}, def},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := jobTimeoutMS(tc.args, def)
			if got != tc.want {
				t.Fatalf("jobTimeoutMS(%v)=%d want %d", tc.args, got, tc.want)
			}
		})
	}
}

func TestTakePositionalSkipsTimeout(t *testing.T) {
	args := []string{"--session-id", "sess-x", "--timeout", "60s", "--tab-id", "123", "document.title"}
	got := takePositional(args, 0)
	if got != "document.title" {
		t.Fatalf("takePositional= %q want document.title", got)
	}
}
