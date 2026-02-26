package docker

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strings"
)

// containerInspect mirrors the fields we care about from `docker inspect`.
type containerInspect struct {
	Name   string `json:"Name"`
	Config struct {
		Image  string            `json:"Image"`
		Env    []string          `json:"Env"`
		Labels map[string]string `json:"Labels"`
	} `json:"Config"`
	HostConfig struct {
		RestartPolicy struct {
			Name string `json:"Name"`
		} `json:"RestartPolicy"`
		Binds []string `json:"Binds"` // host:container[:options]
	} `json:"HostConfig"`
	NetworkSettings struct {
		Ports map[string][]struct {
			HostIP   string `json:"HostIp"`
			HostPort string `json:"HostPort"`
		} `json:"Ports"`
	} `json:"NetworkSettings"`
}

// ContainerInfo holds the subset of inspect data used for compose reconstruction.
type ContainerInfo struct {
	ServiceName   string
	Image         string
	RestartPolicy string
	Ports         []string // "hostPort:containerPort/proto"
	Volumes       []string // "host:container[:options]"
	Environment   []string // "KEY=VALUE" (compose-managed vars only)
}

// Project groups containers that share a compose project label.
type Project struct {
	Name       string
	Containers []ContainerInfo
}

// DiscoverProjects finds all running Docker Compose projects via label inspection.
// It uses a single batched `docker inspect` call for efficiency.
func DiscoverProjects() ([]Project, error) {
	ids, err := runningContainerIDs()
	if err != nil {
		return nil, fmt.Errorf("list containers: %w", err)
	}
	if len(ids) == 0 {
		return nil, nil
	}

	containers, err := inspectContainers(ids)
	if err != nil {
		return nil, fmt.Errorf("inspect containers: %w", err)
	}

	projectMap := make(map[string][]ContainerInfo)

	for _, c := range containers {
		labels := c.Config.Labels

		projectName, ok := labels["com.docker.compose.project"]
		if !ok {
			continue // not a compose container
		}

		serviceName, ok := labels["com.docker.compose.service"]
		if !ok {
			continue
		}

		info := ContainerInfo{
			ServiceName:   serviceName,
			Image:         c.Config.Image,
			RestartPolicy: normaliseRestartPolicy(c.HostConfig.RestartPolicy.Name),
			Volumes:       c.HostConfig.Binds,
			Ports:         extractPorts(c.NetworkSettings.Ports),
			Environment:   filterEnv(c.Config.Env),
		}

		projectMap[projectName] = append(projectMap[projectName], info)
	}

	var projects []Project
	for name, containers := range projectMap {
		// Sort containers by service name for deterministic output.
		sort.Slice(containers, func(i, j int) bool {
			return containers[i].ServiceName < containers[j].ServiceName
		})
		projects = append(projects, Project{Name: name, Containers: containers})
	}

	// Sort projects by name.
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Name < projects[j].Name
	})

	return projects, nil
}

// ToCompose renders the project as a minimal but valid compose YAML string.
// It only includes fields that are actually set — no empty keys.
func (p *Project) ToCompose() []byte {
	var b strings.Builder

	b.WriteString("# Reconstructed by StackSnap from running containers.\n")
	b.WriteString("# This is a best-effort snapshot — review before use.\n")
	b.WriteString("services:\n")

	for _, c := range p.Containers {
		b.WriteString("  " + c.ServiceName + ":\n")
		b.WriteString("    image: " + c.Image + "\n")

		if c.RestartPolicy != "" {
			b.WriteString("    restart: " + c.RestartPolicy + "\n")
		}

		if len(c.Ports) > 0 {
			b.WriteString("    ports:\n")
			for _, port := range c.Ports {
				b.WriteString("      - \"" + port + "\"\n")
			}
		}

		if len(c.Volumes) > 0 {
			b.WriteString("    volumes:\n")
			for _, vol := range c.Volumes {
				b.WriteString("      - \"" + vol + "\"\n")
			}
		}

		if len(c.Environment) > 0 {
			b.WriteString("    environment:\n")
			for _, env := range c.Environment {
				// Split on first = only.
				k, v, _ := strings.Cut(env, "=")
				b.WriteString("      " + k + ": \"" + escapeYAML(v) + "\"\n")
			}
		}
	}

	return []byte(b.String())
}

// ── helpers ──────────────────────────────────────────────────────────────────

func runningContainerIDs() ([]string, error) {
	out, err := exec.Command("docker", "ps", "-q").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}

	var ids []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			ids = append(ids, line)
		}
	}
	return ids, nil
}

func inspectContainers(ids []string) ([]containerInspect, error) {
	args := append([]string{"inspect"}, ids...)
	out, err := exec.Command("docker", args...).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}

	var containers []containerInspect
	if err := json.Unmarshal(out, &containers); err != nil {
		return nil, fmt.Errorf("parse inspect output: %w", err)
	}
	return containers, nil
}

func normaliseRestartPolicy(name string) string {
	switch name {
	case "always", "unless-stopped", "on-failure":
		return name
	case "no", "":
		return "" // omit — "no" is the default
	default:
		return name
	}
}

func extractPorts(ports map[string][]struct {
	HostIP   string `json:"HostIp"`
	HostPort string `json:"HostPort"`
}) []string {
	var result []string
	for containerPort, bindings := range ports {
		for _, b := range bindings {
			if b.HostPort == "" {
				continue
			}
			// containerPort is like "80/tcp" — strip protocol for display.
			cp := strings.Split(containerPort, "/")[0]
			entry := b.HostPort + ":" + cp
			if b.HostIP != "" && b.HostIP != "0.0.0.0" {
				entry = b.HostIP + ":" + entry
			}
			result = append(result, entry)
		}
	}
	sort.Strings(result)
	return result
}

// filterEnv removes environment variables that Docker or Compose injects
// internally — they're noise in a reconstructed compose file.
func filterEnv(env []string) []string {
	skip := map[string]bool{
		"PATH": true, "HOME": true, "HOSTNAME": true,
		"TERM": true, "LANG": true, "LC_ALL": true,
	}

	var result []string
	for _, e := range env {
		k, _, _ := strings.Cut(e, "=")
		if skip[k] {
			continue
		}
		result = append(result, e)
	}
	return result
}

// escapeYAML escapes double-quotes inside a YAML double-quoted string.
func escapeYAML(s string) string {
	return strings.ReplaceAll(s, `"`, `\"`)
}
