package pkg

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// fetchServiceLogs retrieves logs from secret-vm-* systemd services
func fetchServicesLogs() (string, error) {
	var buf bytes.Buffer
	cmd := exec.Command("journalctl", "-u", "secret-vm-network-setup", "-u", "secret-vm-startup", "-u", "secret-vm-attest-rest", "-u", "secret-vm-docker-start")
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return buf.String(), nil
}

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

// formatBytes converts bytes to a human-readable string in B, kB, MB, GB, etc.
func formatBytes(b uint64) string {
    const base = 1024.0
    sizes := []string{"B", "kB", "MB", "GB", "TB", "PB"}
    f := float64(b)
    i := 0
    for f >= base && i < len(sizes)-1 {
        f /= base
        i++
    }
    // show two decimals for MB and above
    if i > 1 {
        return fmt.Sprintf("%.2f %s", f, sizes[i])
    }
    return fmt.Sprintf("%.0f %s", f, sizes[i])
}