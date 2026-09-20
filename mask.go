package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"syscall"
)

type RuleConfig struct {
	Name          string `json:"name"`
	Enabled       bool   `json:"enabled"`
	Pattern       string `json:"pattern"`
	Replacement   string `json:"replacement"`
	CaseSensitive bool   `json:"case_sensitive"`
}

type Config struct {
	Rules []RuleConfig `json:"rules"`
}

type CompiledRule struct {
	Name        string
	Regex       *regexp.Regexp
	Replacement string
}

func getSystemVars() (string, string) {
	username := ""

	if u, err := user.Current(); err == nil {
		username = u.Username

		if idx := strings.LastIndex(username, "\\"); idx != -1 {
			username = username[idx+1:]
		}
	}

	hostname, _ := os.Hostname()

	return username, hostname
}

func LoadConfig() (Config, error) {
	execPath, err := os.Executable()
	if err != nil {
		return Config{}, err
	}

	configPath := filepath.Join(filepath.Dir(execPath), "config.json")

	file, err := os.Open(configPath)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	var cfg Config

	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func CompileRules(cfg Config) []CompiledRule {
	username, hostname := getSystemVars()

	var compiled []CompiledRule

	for _, rule := range cfg.Rules {
		if !rule.Enabled {
			continue
		}

		pattern := rule.Pattern

		if strings.Contains(pattern, "$USERNAME") {
			if len(username) <= 2 {
				continue
			}

			pattern = strings.ReplaceAll(
				pattern,
				"$USERNAME",
				regexp.QuoteMeta(username),
			)
		}

		if strings.Contains(pattern, "$HOSTNAME") {
			if len(hostname) <= 1 {
				continue
			}

			pattern = strings.ReplaceAll(
				pattern,
				"$HOSTNAME",
				regexp.QuoteMeta(hostname),
			)
		}

		if !rule.CaseSensitive {
			pattern = "(?i)" + pattern
		}

		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}

		compiled = append(compiled, CompiledRule{
			Name:        rule.Name,
			Regex:       re,
			Replacement: rule.Replacement,
		})
	}

	return compiled
}

func maskText(text string, rules []CompiledRule) string {
	for _, rule := range rules {
		text = rule.Regex.ReplaceAllString(text, rule.Replacement)
	}

	return text
}

func copyToClipboard(content string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")

	case "windows":
		cmd = exec.Command("clip.exe")

	case "linux":
		if _, err := exec.LookPath("wl-copy"); err == nil {
			cmd = exec.Command("wl-copy")
		} else if _, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		} else if _, err := exec.LookPath("xsel"); err == nil {
			cmd = exec.Command("xsel", "--clipboard", "--input")
		} else if _, err := exec.LookPath("clip.exe"); err == nil {
			cmd = exec.Command("clip.exe") // WSL
		} else {
			return fmt.Errorf("no clipboard utility found")
		}

	default:
		return fmt.Errorf("unsupported OS")
	}

	cmd.Stdin = strings.NewReader(content)

	return cmd.Run()
}

func main() {
	copyClipboard := flag.Bool(
		"c",
		false,
		"copy masked output to clipboard",
	)

	flag.Parse()

	signal.Ignore(syscall.SIGPIPE)

	cfg, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"\033[1;31m[mask error] Could not load config.json: %v\033[0m\n",
			err,
		)
		os.Exit(1)
	}

	rules := CompileRules(cfg)

	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()

	var buf []byte

	// Only keep a copy when -c is used.
	var clipBuffer strings.Builder

	for {
		b, err := reader.ReadByte()

		if err != nil {
			if len(buf) > 0 {
				masked := maskText(string(buf), rules)

				writer.WriteString(masked)

				if *copyClipboard {
					clipBuffer.WriteString(masked)
				}

				writer.Flush()
			}

			if err == io.EOF {
				break
			}

			break
		}

		buf = append(buf, b)

		if b == '\n' || b == '\r' {
			masked := maskText(string(buf), rules)

			writer.WriteString(masked)

			if *copyClipboard {
				clipBuffer.WriteString(masked)
			}

			writer.Flush()

			buf = buf[:0]
		}
	}

	if *copyClipboard && clipBuffer.Len() > 0 {
		if err := copyToClipboard(clipBuffer.String()); err != nil {
			fmt.Fprintf(
				os.Stderr,
				"\033[1;33m[mask warning] Failed to copy to clipboard: %v\033[0m\n",
				err,
			)
		}
	}
}
