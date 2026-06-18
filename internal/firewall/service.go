package firewall

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Service manages firewall state and port rules through the detected backend.
type Service struct {
	runner commandRunner
}

// NewService creates a firewall service with bounded command execution time.
func NewService() Service {
	return Service{runner: commandRunner{timeout: 12 * time.Second}}
}

// Snapshot describes the detected firewall backend and its visible rules.
type Snapshot struct {
	Backend   string `json:"backend"`
	Available bool   `json:"available"`
	Enabled   bool   `json:"enabled"`
	Message   string `json:"message"`
	Rules     []Rule `json:"rules"`
}

// Rule is a normalized firewall port rule for API and UI consumers.
type Rule struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Action   string `json:"action"`
	Source   string `json:"source"`
}

// PortRuleRequest is the only accepted input shape for mutating firewall rules.
type PortRuleRequest struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
}

// Snapshot detects firewall availability and lists currently visible rules.
func (s Service) Snapshot(ctx context.Context) Snapshot {
	backend := detectBackend()
	if backend == "" {
		return Snapshot{Available: false, Message: "未检测到 ufw、firewalld、nftables 或 iptables"}
	}

	enabled, message := s.enabled(ctx, backend)
	return Snapshot{
		Backend:   backend,
		Available: true,
		Enabled:   enabled,
		Message:   message,
		Rules:     s.rules(ctx, backend),
	}
}

// Enable starts or enables the detected firewall backend when supported.
func (s Service) Enable(ctx context.Context) error {
	backend := detectBackend()
	if backend == "" {
		return errors.New("未检测到可用防火墙")
	}
	switch backend {
	case "ufw":
		_, err := s.runner.run(ctx, "ufw", "enable")
		return err
	case "firewalld":
		_, err := s.runner.run(ctx, "systemctl", "start", "firewalld")
		return err
	case "nft":
		_, err := s.runner.run(ctx, "systemctl", "start", "nftables")
		return err
	default:
		return nil
	}
}

// Disable stops or disables the detected firewall backend when supported.
func (s Service) Disable(ctx context.Context) error {
	backend := detectBackend()
	if backend == "" {
		return errors.New("未检测到可用防火墙")
	}
	switch backend {
	case "ufw":
		_, err := s.runner.run(ctx, "ufw", "disable")
		return err
	case "firewalld":
		_, err := s.runner.run(ctx, "systemctl", "stop", "firewalld")
		return err
	case "nft":
		_, err := s.runner.run(ctx, "systemctl", "stop", "nftables")
		return err
	default:
		return nil
	}
}

// AddRule opens a TCP or UDP port after validating the structured request.
func (s Service) AddRule(ctx context.Context, req PortRuleRequest) error {
	protocol, err := normalizeRule(req)
	if err != nil {
		return err
	}
	backend := detectBackend()
	if backend == "" {
		return errors.New("未检测到可用防火墙")
	}
	port := strconv.Itoa(req.Port)
	switch backend {
	case "ufw":
		_, err = s.runner.run(ctx, "ufw", "allow", port+"/"+protocol)
	case "firewalld":
		_, err = s.runner.run(ctx, "firewall-cmd", "--permanent", "--add-port="+port+"/"+protocol)
		if err == nil {
			_, err = s.runner.run(ctx, "firewall-cmd", "--reload")
		}
	case "nft":
		_, err = s.runner.run(ctx, "nft", "add", "rule", "inet", "filter", "input", protocol, "dport", port, "accept")
	default:
		_, err = s.runner.run(ctx, "iptables", "-I", "INPUT", "-p", protocol, "--dport", port, "-j", "ACCEPT")
	}
	return err
}

// DeleteRule removes a TCP or UDP port rule when the backend supports deletion.
func (s Service) DeleteRule(ctx context.Context, req PortRuleRequest) error {
	protocol, err := normalizeRule(req)
	if err != nil {
		return err
	}
	backend := detectBackend()
	if backend == "" {
		return errors.New("未检测到可用防火墙")
	}
	port := strconv.Itoa(req.Port)
	switch backend {
	case "ufw":
		_, err = s.runner.run(ctx, "ufw", "delete", "allow", port+"/"+protocol)
	case "firewalld":
		_, err = s.runner.run(ctx, "firewall-cmd", "--permanent", "--remove-port="+port+"/"+protocol)
		if err == nil {
			_, err = s.runner.run(ctx, "firewall-cmd", "--reload")
		}
	case "nft":
		err = errors.New("nftables 删除规则需要句柄，当前版本请使用系统 nft 管理已有规则")
	default:
		_, err = s.runner.run(ctx, "iptables", "-D", "INPUT", "-p", protocol, "--dport", port, "-j", "ACCEPT")
	}
	return err
}

func (s Service) enabled(ctx context.Context, backend string) (bool, string) {
	switch backend {
	case "ufw":
		output, err := s.runner.run(ctx, "ufw", "status")
		if err != nil {
			return false, err.Error()
		}
		text := string(output)
		return strings.Contains(text, "Status: active"), firstLine(text)
	case "firewalld":
		output, err := s.runner.run(ctx, "firewall-cmd", "--state")
		if err != nil {
			return false, err.Error()
		}
		text := strings.TrimSpace(string(output))
		return text == "running", text
	case "nft":
		output, err := s.runner.run(ctx, "systemctl", "is-active", "nftables")
		text := strings.TrimSpace(string(output))
		return err == nil && text == "active", text
	default:
		output, err := s.runner.run(ctx, "iptables", "-S")
		if err != nil {
			return false, err.Error()
		}
		return true, fmt.Sprintf("iptables 规则 %d 行", len(strings.Split(strings.TrimSpace(string(output)), "\n")))
	}
}

func (s Service) rules(ctx context.Context, backend string) []Rule {
	switch backend {
	case "ufw":
		return s.ufwRules(ctx)
	case "firewalld":
		return s.firewalldRules(ctx)
	case "iptables":
		return s.iptablesRules(ctx)
	default:
		return []Rule{}
	}
}

func (s Service) ufwRules(ctx context.Context) []Rule {
	output, err := s.runner.run(ctx, "ufw", "status")
	if err != nil {
		return []Rule{}
	}
	rules := []Rule{}
	for _, line := range strings.Split(string(output), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 3 || !strings.Contains(fields[0], "/") {
			continue
		}
		port, protocol, ok := strings.Cut(fields[0], "/")
		if !ok {
			continue
		}
		number, err := strconv.Atoi(port)
		if err != nil {
			continue
		}
		rules = append(rules, Rule{Port: number, Protocol: protocol, Action: fields[1], Source: strings.Join(fields[2:], " ")})
	}
	return rules
}

func (s Service) firewalldRules(ctx context.Context) []Rule {
	output, err := s.runner.run(ctx, "firewall-cmd", "--list-ports")
	if err != nil {
		return []Rule{}
	}
	rules := []Rule{}
	for _, item := range strings.Fields(string(output)) {
		port, protocol, ok := strings.Cut(item, "/")
		if !ok {
			continue
		}
		number, err := strconv.Atoi(port)
		if err != nil {
			continue
		}
		rules = append(rules, Rule{Port: number, Protocol: protocol, Action: "ALLOW", Source: "any"})
	}
	return rules
}

func (s Service) iptablesRules(ctx context.Context) []Rule {
	output, err := s.runner.run(ctx, "iptables", "-S", "INPUT")
	if err != nil {
		return []Rule{}
	}
	rules := []Rule{}
	for _, line := range strings.Split(string(output), "\n") {
		fields := strings.Fields(line)
		for i, field := range fields {
			if field != "--dport" || i+1 >= len(fields) {
				continue
			}
			port, err := strconv.Atoi(fields[i+1])
			if err != nil {
				continue
			}
			protocol := findFlagValue(fields, "-p")
			action := findFlagValue(fields, "-j")
			rules = append(rules, Rule{Port: port, Protocol: protocol, Action: action, Source: "any"})
		}
	}
	return rules
}

type commandRunner struct {
	timeout time.Duration
}

func (r commandRunner) run(ctx context.Context, name string, args ...string) ([]byte, error) {
	if r.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.timeout)
		defer cancel()
	}
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.CombinedOutput()
}

func detectBackend() string {
	for _, candidate := range []string{"ufw", "firewall-cmd", "nft", "iptables"} {
		if _, err := exec.LookPath(candidate); err == nil {
			if candidate == "firewall-cmd" {
				return "firewalld"
			}
			return candidate
		}
	}
	return ""
}

func normalizeRule(req PortRuleRequest) (string, error) {
	if req.Port < 1 || req.Port > 65535 {
		return "", errors.New("端口必须在 1-65535 之间")
	}
	protocol := strings.ToLower(strings.TrimSpace(req.Protocol))
	if protocol == "" {
		protocol = "tcp"
	}
	if protocol != "tcp" && protocol != "udp" {
		return "", errors.New("协议只支持 tcp 或 udp")
	}
	return protocol, nil
}

func firstLine(value string) string {
	lines := strings.Split(strings.TrimSpace(value), "\n")
	if len(lines) == 0 {
		return ""
	}
	return strings.TrimSpace(lines[0])
}

func findFlagValue(fields []string, flag string) string {
	for i, field := range fields {
		if field == flag && i+1 < len(fields) {
			return fields[i+1]
		}
	}
	return ""
}
