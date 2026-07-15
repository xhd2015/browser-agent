// Chaos fuzzy harness: real Chrome + browser-agent against seeds from
// --links PATH (markdown/text URL extraction) or --random-links.
// Dice picks the next op.
//
// Usage:
//
//	go run . --links path/to/links.md --seed 42 --max-ops 40
//	go run . --random-links --dry-run --seed 7 --max-ops 20
//	go run . --links path.md --session-id sess-xxx   # reuse existing session
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/xhd2015/browser-agent/script/debug/chaos-useful-links/seedload"
)

const identityExpr = `({href: location.href, title: document.title, ready: document.readyState})`

func main() {
	var (
		seed            = flag.Int64("seed", time.Now().UnixNano()%1_000_000, "RNG seed for dice (reproducible)")
		maxOps          = flag.Int("max-ops", 40, "maximum chaos operations")
		maxTabs         = flag.Int("max-tabs", 5, "max content tabs (session /go excluded)")
		linksPath       = flag.String("links", "", "path to markdown/text link file (mutex with --random-links)")
		randomLinks     = flag.Bool("random-links", false, "use built-in public seed catalog (mutex with --links)")
		includeArchived = flag.Bool("include-archived", false, "include links under Historical/Archived/Deprecated headings")
		maxSeeds        = flag.Int("max-seeds", 0, "cap seeds after dedupe (0=all)")
		kindFilter      = flag.String("kind", "", "optional seed kind filter when metadata present")
		envFilter       = flag.String("env", "", "optional seed env filter when metadata present")
		outDir          = flag.String("out", "", "artifact directory (default: ./out/<run-id>)")
		sessionID       = flag.String("session-id", "", "reuse session; default: run session new")
		binPath         = flag.String("browser-agent", "", "browser-agent binary (default: PATH)")
		dryRun          = flag.Bool("dry-run", false, "plan ops only; no Chrome/session commands")
		opTimeout       = flag.Duration("op-timeout", 25*time.Second, "per-op timeout")
		waitExt         = flag.Duration("wait-extension", 90*time.Second, "wait for extension connect after session new")
		jsonSummary     = flag.Bool("json", false, "print run.json path and exit code only (still writes artifacts)")
	)
	flag.Parse()

	here, err := os.Getwd()
	if err != nil {
		fatal(2, "cwd: %v", err)
	}
	scriptDir := here
	// When invoked from module root, prefer script dir for default out/.
	if _, err := os.Stat(filepath.Join(here, "script/debug/chaos-useful-links")); err == nil {
		if filepath.Base(here) != "chaos-useful-links" {
			cand := filepath.Join(here, "script/debug/chaos-useful-links")
			if st, err := os.Stat(cand); err == nil && st.IsDir() {
				scriptDir = cand
			}
		}
	}

	opts := seedload.Options{
		IncludeArchived: *includeArchived,
		MaxSeeds:        *maxSeeds,
		Kind:            strings.TrimSpace(*kindFilter),
		Env:             strings.TrimSpace(*envFilter),
	}
	resolved, err := seedload.ResolveSeedSource(*linksPath, *randomLinks, opts)
	if err != nil {
		fatal(2, "seed source: %v", err)
	}
	if resolved == nil || len(resolved.Seeds) == 0 {
		fatal(2, "no seeds resolved")
	}
	seeds := toMainSeeds(resolved.Seeds)
	srcLabel := "random-links"
	if resolved.Source.Type == "links" {
		srcLabel = resolved.Source.Path
		if srcLabel == "" {
			srcLabel = *linksPath
		}
	}
	fmt.Fprintf(os.Stderr, "chaos: loaded %d seeds from %s\n", len(seeds), srcLabel)

	runID := time.Now().UTC().Format("20060102T150405Z")
	if *outDir == "" {
		*outDir = filepath.Join(scriptDir, "out", runID)
	}
	if err := os.MkdirAll(filepath.Join(*outDir, "issues"), 0o755); err != nil {
		fatal(2, "mkdir out: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(*outDir, "screenshots"), 0o755); err != nil {
		fatal(2, "mkdir screenshots: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(*outDir, "session-info"), 0o755); err != nil {
		fatal(2, "mkdir session-info: %v", err)
	}

	// Persist resolved corpus snapshot under out/.
	resolvedPath := filepath.Join(*outDir, "corpus.resolved.json")
	if err := writeJSON(resolvedPath, resolved); err != nil {
		fatal(2, "write corpus.resolved.json: %v", err)
	}

	bin := *binPath
	if bin == "" {
		bin, err = exec.LookPath("browser-agent")
		if err != nil && !*dryRun {
			fatal(2, "browser-agent not on PATH: %v (build/install or pass --browser-agent)", err)
		}
		if bin == "" {
			bin = "browser-agent"
		}
	}

	corpusLabel := "random-links"
	if strings.TrimSpace(*linksPath) != "" {
		corpusLabel = *linksPath
	}

	run := &RunRecord{
		RunID:     runID,
		SeedRNG:   *seed,
		MaxOps:    *maxOps,
		MaxTabs:   *maxTabs,
		Corpus:    corpusLabel,
		Chrome:    "default",
		DryRun:    *dryRun,
		StartedAt: time.Now().UTC(),
		Counts:    map[string]int{},
		Steps:     nil,
		Issues:    nil,
	}

	ctx := context.Background()
	agent := &AgentCLI{Bin: bin, Timeout: *opTimeout}
	dice := newDice(*seed)

	if !*dryRun {
		sid := strings.TrimSpace(*sessionID)
		if sid == "" {
			sid = strings.TrimSpace(os.Getenv("BROWSER_AGENT_SESSION_ID"))
		}
		if sid == "" {
			fmt.Fprintf(os.Stderr, "chaos: running browser-agent session new (default Chrome)…\n")
			newID, raw, err := agent.sessionNew(ctx)
			if err != nil {
				fatal(2, "%v\n%s", err, raw)
			}
			sid = newID
			fmt.Fprintf(os.Stderr, "chaos: session-id %s\n", sid)
			fmt.Fprintf(os.Stderr, "chaos: keep the /go?session= tab open; load unpacked extension if not connected.\n")
		}
		agent.SessionID = sid
		run.SessionID = sid

		// Wait for extension.
		deadline := time.Now().Add(*waitExt)
		for {
			info, _, _, err := agent.infoJSON(ctx)
			if err == nil && extensionConnected(info) {
				fmt.Fprintf(os.Stderr, "chaos: extension connected\n")
				break
			}
			if time.Now().After(deadline) {
				fatal(2, "extension not connected within %s; open session page and Load unpacked extension", *waitExt)
			}
			fmt.Fprintf(os.Stderr, "chaos: waiting for extension… (%v)\n", err)
			time.Sleep(2 * time.Second)
		}
	} else {
		run.SessionID = "dry-run"
		fmt.Fprintf(os.Stderr, "chaos: dry-run mode (no Chrome)\n")
	}

	var contentTabs []TabState
	seedByID := map[string]Seed{}
	for _, s := range seeds {
		seedByID[s.ID] = s
	}

	for step := 1; step <= *maxOps; step++ {
		op := dice.pickOp(len(contentTabs), *maxTabs)
		planned := PlannedOp{Step: step, Op: op}

		var rec StepRecord
		start := time.Now()

		switch op {
		case opOpenSeed:
			s := dice.pickSeed(seeds)
			planned.SeedID = s.ID
			rec = execOpenSeed(ctx, agent, s, *dryRun, &contentTabs, *outDir, step)

		case opNavigateSeed:
			tab := dice.pickTab(contentTabs)
			s := dice.pickSeed(seeds)
			planned.SeedID = s.ID
			planned.TabID = tab.TabID
			rec = execNavigate(ctx, agent, tab, s, *dryRun, &contentTabs)

		case opEvalIdentity:
			tab := dice.pickTab(contentTabs)
			planned.TabID = tab.TabID
			planned.SeedID = tab.SeedID
			rec = execEval(ctx, agent, tab, *dryRun)

		case opScreenshot:
			tab := dice.pickTab(contentTabs)
			planned.TabID = tab.TabID
			planned.SeedID = tab.SeedID
			rec = execScreenshot(ctx, agent, tab, *dryRun, *outDir, step)

		case opSessionInfo:
			rec = execSessionInfo(ctx, agent, *dryRun, *outDir, step)

		case opLogs:
			tab := dice.pickTab(contentTabs)
			planned.TabID = tab.TabID
			planned.SeedID = tab.SeedID
			rec = execLogs(ctx, agent, tab, *dryRun)

		case opBackgroundEval:
			var info map[string]any
			if !*dryRun {
				info, _, _, _ = agent.infoJSON(ctx)
			}
			active := activeTabIDFromInfo(info)
			tab := dice.pickBackgroundTab(contentTabs, active)
			planned.TabID = tab.TabID
			planned.SeedID = tab.SeedID
			rec = execEval(ctx, agent, tab, *dryRun)
			rec.Op = opBackgroundEval

		case opRacePair:
			tab := dice.pickTab(contentTabs)
			planned.TabID = tab.TabID
			planned.SeedID = tab.SeedID
			rec = execRace(ctx, agent, tab, *dryRun, *outDir, step)

		default:
			rec = StepRecord{Op: op, Class: classFailOther, Error: "unknown op"}
		}

		rec.Step = step
		if rec.Op == "" {
			rec.Op = op
		}
		if rec.DurationMS == 0 {
			rec.DurationMS = time.Since(start).Milliseconds()
		}
		if rec.SeedID == "" {
			rec.SeedID = planned.SeedID
		}
		if rec.TabID == 0 {
			rec.TabID = planned.TabID
		}

		run.Steps = append(run.Steps, rec)
		run.Counts[string(rec.Class)]++

		fmt.Printf("  [%02d] %-16s class=%-16s seed=%s tab=%d %dms\n",
			step, rec.Op, rec.Class, rec.SeedID, rec.TabID, rec.DurationMS)
		if rec.Error != "" {
			fmt.Fprintf(os.Stderr, "         error: %s\n", truncate(rec.Error, 200))
		}

		if isP0orP1(rec.Class) || rec.Class == classFailAttach || rec.Class == classFailOther {
			issue := buildIssue(run, rec, seedByID, contentTabs)
			run.Issues = append(run.Issues, issue)
			writeIssue(*outDir, issue)
		}
	}

	run.TabsAtEnd = contentTabs
	fin := time.Now().UTC()
	run.FinishedAt = &fin

	runPath := filepath.Join(*outDir, "run.json")
	if err := writeJSON(runPath, run); err != nil {
		fatal(2, "write run.json: %v", err)
	}

	// Summary
	p0p1 := 0
	for _, iss := range run.Issues {
		if iss.Severity == sevP0 || iss.Severity == sevP1 {
			p0p1++
		}
	}

	if !*jsonSummary {
		fmt.Println()
		fmt.Printf("run_id     %s\n", run.RunID)
		fmt.Printf("session    %s\n", run.SessionID)
		fmt.Printf("seed       %d\n", run.SeedRNG)
		fmt.Printf("ops        %d\n", len(run.Steps))
		fmt.Printf("results   ")
		// stable-ish print
		order := []ResultClass{
			classOKLoaded, classOKAuthWall, classOKSlow, classOKInfo,
			classFailRouting, classFailTimeout, classFailDisconnect, classFailAttach, classFailCrash, classFailOther,
		}
		for _, c := range order {
			if n := run.Counts[string(c)]; n > 0 {
				fmt.Printf(" %s=%d", c, n)
			}
		}
		fmt.Println()
		if len(run.Issues) > 0 {
			fmt.Println("issues:")
			for _, iss := range run.Issues {
				fmt.Printf("  %s  %-16s  %s  step=%d\n", iss.Severity, iss.Category, iss.SeedID, iss.Step)
			}
		} else {
			fmt.Println("issues:    (none)")
		}
		fmt.Printf("artifacts: %s\n", *outDir)
	} else {
		fmt.Println(runPath)
	}

	if p0p1 > 0 {
		os.Exit(1)
	}
}

func toMainSeeds(in []seedload.Seed) []Seed {
	out := make([]Seed, 0, len(in))
	for _, s := range in {
		out = append(out, Seed{
			ID:     s.ID,
			URL:    s.URL,
			Kind:   s.Kind,
			Env:    s.Env,
			Market: s.Market,
			Title:  s.Title,
			Tags:   s.Tags,
		})
	}
	return out
}

func execOpenSeed(ctx context.Context, agent *AgentCLI, s Seed, dry bool, tabs *[]TabState, outDir string, step int) StepRecord {
	rec := StepRecord{Op: opOpenSeed, SeedID: s.ID}
	if dry {
		fakeID := int64(1000 + len(*tabs) + step)
		*tabs = append(*tabs, TabState{TabID: fakeID, SeedID: s.ID, URL: s.URL})
		rec.TabID = fakeID
		rec.Class = classOKLoaded
		return rec
	}
	start := time.Now()
	tabID, stdout, stderr, err := agent.createTab(ctx, s.URL)
	rec.DurationMS = time.Since(start).Milliseconds()
	rec.Stdout = truncate(stdout, 2000)
	rec.Stderr = truncate(stderr, 1000)
	if err != nil {
		rec.Class = classifyError(err)
		rec.Error = err.Error()
		return rec
	}
	*tabs = append(*tabs, TabState{TabID: tabID, SeedID: s.ID, URL: s.URL})
	rec.TabID = tabID
	// Brief settle then identity check.
	time.Sleep(1500 * time.Millisecond)
	evalRec := execEval(ctx, agent, TabState{TabID: tabID, SeedID: s.ID, URL: s.URL}, false)
	rec.Class = evalRec.Class
	rec.EvalHref = evalRec.EvalHref
	rec.EvalTitle = evalRec.EvalTitle
	rec.Error = evalRec.Error
	if rec.Class == classOKLoaded || rec.Class == classOKAuthWall || rec.Class == classOKSlow {
		// keep
	} else if evalRec.Error != "" {
		// open succeeded; classification from eval
	}
	return rec
}

func execNavigate(ctx context.Context, agent *AgentCLI, tab TabState, s Seed, dry bool, tabs *[]TabState) StepRecord {
	rec := StepRecord{Op: opNavigateSeed, SeedID: s.ID, TabID: tab.TabID}
	if dry {
		updateTab(tabs, tab.TabID, s.ID, s.URL)
		rec.Class = classOKLoaded
		return rec
	}
	start := time.Now()
	stdout, stderr, err := agent.cdpNavigate(ctx, tab.TabID, s.URL)
	rec.DurationMS = time.Since(start).Milliseconds()
	rec.Stdout = truncate(stdout, 2000)
	rec.Stderr = truncate(stderr, 1000)
	if err != nil {
		rec.Class = classifyError(err)
		rec.Error = err.Error()
		return rec
	}
	updateTab(tabs, tab.TabID, s.ID, s.URL)
	time.Sleep(1500 * time.Millisecond)
	evalRec := execEval(ctx, agent, TabState{TabID: tab.TabID, SeedID: s.ID, URL: s.URL}, false)
	rec.Class = evalRec.Class
	rec.EvalHref = evalRec.EvalHref
	rec.EvalTitle = evalRec.EvalTitle
	rec.Error = evalRec.Error
	return rec
}

func execEval(ctx context.Context, agent *AgentCLI, tab TabState, dry bool) StepRecord {
	rec := StepRecord{Op: opEvalIdentity, TabID: tab.TabID, SeedID: tab.SeedID}
	if dry {
		rec.Class = classOKLoaded
		rec.EvalHref = tab.URL
		return rec
	}
	start := time.Now()
	stdout, stderr, err := agent.eval(ctx, tab.TabID, identityExpr)
	dur := time.Since(start)
	rec.DurationMS = dur.Milliseconds()
	rec.Stdout = truncate(stdout, 2000)
	rec.Stderr = truncate(stderr, 1000)
	href, title, _ := parseEvalIdentity(stdout)
	rec.EvalHref = href
	rec.EvalTitle = title
	rec.Class = classifyEval(tab.URL, href, dur, err)
	if err != nil {
		rec.Error = err.Error()
	}
	return rec
}

func execScreenshot(ctx context.Context, agent *AgentCLI, tab TabState, dry bool, outDir string, step int) StepRecord {
	rec := StepRecord{Op: opScreenshot, TabID: tab.TabID, SeedID: tab.SeedID}
	path := filepath.Join(outDir, "screenshots", fmt.Sprintf("step-%02d.png", step))
	if dry {
		rec.Class = classOKInfo
		rec.Evidence = map[string]string{"screenshot": path}
		return rec
	}
	start := time.Now()
	stdout, stderr, err := agent.screenshot(ctx, tab.TabID, path)
	dur := time.Since(start)
	rec.DurationMS = dur.Milliseconds()
	rec.Stdout = truncate(stdout, 500)
	rec.Stderr = truncate(stderr, 1000)
	rec.Evidence = map[string]string{"screenshot": path}
	if err != nil {
		rec.Class = classifyError(err)
		rec.Error = err.Error()
		return rec
	}
	// Screenshot succeeded — check routing via eval.
	evalRec := execEval(ctx, agent, tab, false)
	if evalRec.Class == classFailRouting {
		rec.Class = classFailRouting
		rec.EvalHref = evalRec.EvalHref
		rec.Error = "screenshot ok but identity eval indicates session-page routing"
		return rec
	}
	if dur >= slowThreshold {
		rec.Class = classOKSlow
	} else {
		rec.Class = classOKInfo
	}
	return rec
}

func execSessionInfo(ctx context.Context, agent *AgentCLI, dry bool, outDir string, step int) StepRecord {
	rec := StepRecord{Op: opSessionInfo}
	path := filepath.Join(outDir, "session-info", fmt.Sprintf("step-%02d.json", step))
	if dry {
		rec.Class = classOKInfo
		rec.Evidence = map[string]string{"session_info": path}
		return rec
	}
	start := time.Now()
	info, stdout, stderr, err := agent.infoJSON(ctx)
	rec.DurationMS = time.Since(start).Milliseconds()
	rec.Stdout = truncate(stdout, 4000)
	rec.Stderr = truncate(stderr, 1000)
	_ = os.WriteFile(path, []byte(stdout), 0o644)
	rec.Evidence = map[string]string{"session_info": path}
	if err != nil {
		rec.Class = classifyError(err)
		rec.Error = err.Error()
		return rec
	}
	if !extensionConnected(info) {
		rec.Class = classFailDisconnect
		rec.Error = "session info reports extension not connected"
		return rec
	}
	rec.Class = classOKInfo
	return rec
}

func execLogs(ctx context.Context, agent *AgentCLI, tab TabState, dry bool) StepRecord {
	rec := StepRecord{Op: opLogs, TabID: tab.TabID, SeedID: tab.SeedID}
	if dry {
		rec.Class = classOKInfo
		return rec
	}
	start := time.Now()
	stdout, stderr, err := agent.logs(ctx, tab.TabID)
	rec.DurationMS = time.Since(start).Milliseconds()
	rec.Stdout = truncate(stdout, 2000)
	rec.Stderr = truncate(stderr, 1000)
	if err != nil {
		rec.Class = classifyError(err)
		rec.Error = err.Error()
		return rec
	}
	rec.Class = classOKInfo
	return rec
}

func execRace(ctx context.Context, agent *AgentCLI, tab TabState, dry bool, outDir string, step int) StepRecord {
	rec := StepRecord{Op: opRacePair, TabID: tab.TabID, SeedID: tab.SeedID}
	if dry {
		rec.Class = classOKInfo
		return rec
	}
	path := filepath.Join(outDir, "screenshots", fmt.Sprintf("step-%02d-race.png", step))
	start := time.Now()
	var (
		evalOut, evalErrS, shotOut, shotErrS string
		evalErr, shotErr                     error
	)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		evalOut, evalErrS, evalErr = agent.eval(ctx, tab.TabID, identityExpr)
	}()
	go func() {
		defer wg.Done()
		shotOut, shotErrS, shotErr = agent.screenshot(ctx, tab.TabID, path)
	}()
	wg.Wait()
	rec.DurationMS = time.Since(start).Milliseconds()
	rec.Stdout = truncate(evalOut+"\n"+shotOut, 2000)
	rec.Stderr = truncate(evalErrS+"\n"+shotErrS, 1000)
	rec.Evidence = map[string]string{"screenshot": path}

	if evalErr != nil {
		rec.Class = classifyError(evalErr)
		rec.Error = "eval: " + evalErr.Error()
		return rec
	}
	if shotErr != nil {
		rec.Class = classifyError(shotErr)
		rec.Error = "screenshot: " + shotErr.Error()
		return rec
	}
	href, title, _ := parseEvalIdentity(evalOut)
	rec.EvalHref = href
	rec.EvalTitle = title
	rec.Class = classifyEval(tab.URL, href, time.Since(start), nil)
	return rec
}

func updateTab(tabs *[]TabState, tabID int64, seedID, url string) {
	for i := range *tabs {
		if (*tabs)[i].TabID == tabID {
			(*tabs)[i].SeedID = seedID
			(*tabs)[i].URL = url
			return
		}
	}
}

func buildIssue(run *RunRecord, rec StepRecord, seeds map[string]Seed, tabs []TabState) Issue {
	url := ""
	if s, ok := seeds[rec.SeedID]; ok {
		url = s.URL
	}
	for _, t := range tabs {
		if t.TabID == rec.TabID && t.URL != "" {
			url = t.URL
			break
		}
	}
	seq := make([]string, 0, rec.Step)
	for _, st := range run.Steps {
		if st.Step <= rec.Step {
			seq = append(seq, string(st.Op))
		}
	}
	actual := rec.Error
	if actual == "" {
		actual = fmt.Sprintf("class=%s href=%s", rec.Class, rec.EvalHref)
	}
	return Issue{
		IssueID:    fmt.Sprintf("iss-%03d", len(run.Issues)+1),
		Severity:   severityFor(rec.Class),
		Category:   rec.Class,
		SeedID:     rec.SeedID,
		URL:        url,
		OpSequence: seq,
		Expected:   expectedFor(rec.Class, rec.Op),
		Actual:     actual,
		Step:       rec.Step,
		Evidence:   rec.Evidence,
		CreatedAt:  time.Now().UTC(),
	}
}

func writeIssue(outDir string, issue Issue) {
	path := filepath.Join(outDir, "issues", issue.IssueID+".json")
	_ = writeJSON(path, issue)
}

func writeJSON(path string, v any) error {
	raw, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	return os.WriteFile(path, raw, 0o644)
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func fatal(code int, format string, args ...any) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(code)
}
