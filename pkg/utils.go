package pkg

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// fetchDockerLogsWithSelector retrieves logs from a container based on
// a provided name or numeric index. If name is non-empty, it tries to find
// a container with exactly that name and errors if none found. If useIndex is true,
// it tries to pick the container at that zero-based index and errors if out of range.
// If neither name nor index is provided, it falls back to the first running container.
func fetchDockerLogsWithSelector(name string, index int, useIndex bool, lines int) (string, error) {
	// List running containers as "<ID> <Name>\n"
	out, err := exec.Command("docker", "ps", "--format", "{{.ID}} {{.Names}}").Output()
	if err != nil {
		return "", fmt.Errorf("failed to list Docker containers: %w", err)
	}
	entries := strings.Split(strings.TrimSpace(string(out)), "\n")
	linesArg := "--tail=" + strconv.Itoa(lines)

	// If a name was specified, try to find it
	if name != "" {
		for _, entry := range entries {
			parts := strings.Fields(entry)
			if len(parts) < 2 {
				continue
			}
			if parts[1] == name {
				return getContainerLogs(parts[0], linesArg)
			}
		}
		return "", fmt.Errorf("no running container with name %q found", name)
	}

	// If an index was specified, try to use it
	if useIndex {
		if index < 0 || index >= len(entries) {
			return "", fmt.Errorf("container index %d out of range", index)
		}
		parts := strings.Fields(entries[index])
		if len(parts) >= 1 {
			return getContainerLogs(parts[0], linesArg)
		}
		return "", fmt.Errorf("failed to parse container entry at index %d", index)
	}

	// No selector provided: fallback to first container
	if len(entries) > 0 {
		parts := strings.Fields(entries[0])
		if len(parts) >= 1 {
			return getContainerLogs(parts[0], linesArg)
		}
	}

	return "", fmt.Errorf("no running containers found")
}

// getContainerLogs executes `docker logs` for the given container ID.
func getContainerLogs(containerID, linesArg string) (string, error) {
	var buf bytes.Buffer
	cmd := exec.Command("docker", "logs", linesArg, containerID)
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return buf.String(), nil
}
