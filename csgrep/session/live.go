package session

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type LiveInfo struct {
	PID       int
	SessionID string
	CWD       string
}

type LiveSession struct {
	Session
	PID    int
	Active bool
}

var uuidRe = regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)

func FindLiveProcesses() []LiveInfo {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil
	}

	var live []LiveInfo
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}

		cmdline, err := os.ReadFile(filepath.Join("/proc", e.Name(), "cmdline"))
		if err != nil {
			continue
		}

		args := strings.Split(string(cmdline), "\x00")
		if len(args) == 0 {
			continue
		}

		exe := filepath.Base(args[0])
		if exe != "claude" {
			continue
		}

		if containsArg(args, "--chrome-native-host") || containsArgs(args, "daemon", "run") {
			continue
		}

		info := LiveInfo{PID: pid}

		cwd, err := os.Readlink(filepath.Join("/proc", e.Name(), "cwd"))
		if err == nil {
			info.CWD = cwd
		}

		for i, arg := range args {
			if arg == "--resume" && i+1 < len(args) {
				candidate := args[i+1]
				if uuidRe.MatchString(candidate) {
					info.SessionID = candidate
				}
				break
			}
		}

		live = append(live, info)
	}
	return live
}

func MatchLiveSessions(filter Filter) []LiveSession {
	processes := FindLiveProcesses()
	if len(processes) == 0 {
		return nil
	}

	liveByID := make(map[string]LiveInfo)
	var unmatched []LiveInfo
	for _, p := range processes {
		if p.SessionID != "" {
			liveByID[p.SessionID] = p
		} else {
			unmatched = append(unmatched, p)
		}
	}

	files := FindFiles(Filter{Dir: "", Interactive: false})

	var results []LiveSession

	for _, f := range files {
		s, err := ParseFast(f)
		if err != nil || s == nil {
			continue
		}

		if p, ok := liveByID[s.ID]; ok {
			if !filter.Matches(s) {
				continue
			}
			results = append(results, LiveSession{
				Session: *s,
				PID:     p.PID,
				Active:  true,
			})
			delete(liveByID, s.ID)
		}
	}

	seenSessions := make(map[string]bool)
	for _, r := range results {
		seenSessions[r.ID] = true
	}

	for _, p := range unmatched {
		best := findBestSessionForCWD(p.CWD)
		if best == "" {
			continue
		}
		sid := strings.TrimSuffix(filepath.Base(best), ".jsonl")
		if seenSessions[sid] {
			continue
		}
		s, err := ParseFast(best)
		if err != nil || s == nil {
			continue
		}
		if !filter.Matches(s) {
			continue
		}
		seenSessions[s.ID] = true
		results = append(results, LiveSession{
			Session: *s,
			PID:     p.PID,
			Active:  true,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].LastTime.After(results[j].LastTime)
	})

	return results
}

func findBestSessionForCWD(cwd string) string {
	if cwd == "" {
		return ""
	}

	dir := cwd
	for dir != "/" && dir != "." {
		sessDir := findSessionsDir(dir)
		files := globJSONL(sessDir)
		if len(files) > 0 {
			return newestFile(files)
		}
		dir = filepath.Dir(dir)
	}
	return ""
}

func newestFile(files []string) string {
	var best string
	var bestTime int64
	for _, f := range files {
		info, err := os.Stat(f)
		if err != nil {
			continue
		}
		t := info.ModTime().Unix()
		if t > bestTime {
			bestTime = t
			best = f
		}
	}
	return best
}

func containsArg(args []string, target string) bool {
	for _, a := range args {
		if a == target {
			return true
		}
	}
	return false
}

func containsArgs(args []string, a, b string) bool {
	for i, arg := range args {
		if arg == a && i+1 < len(args) && args[i+1] == b {
			return true
		}
	}
	return false
}
