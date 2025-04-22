package pkg

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// fetchDockerLogs returns the last 'lines' lines of logs from the
// "secret-vm-docker" container or, if not found, the first running container.
func fetchDockerLogs(lines int) (string, error) {
	out, err := exec.Command("docker", "ps", "-q", "--filter", "name=secret-vm-docker").Output()
	if err != nil {
		return "", err
	}
	id := strings.TrimSpace(string(out))
	if id == "" {
		out, err = exec.Command("docker", "ps", "-q").Output()
		if err != nil {
			return "", err
		}
		ids := strings.Fields(string(out))
		if len(ids) == 0 {
			return "", fmt.Errorf("no running containers")
		}
		id = ids[0]
	}
	cmd := exec.Command("docker", "logs", "--tail", strconv.Itoa(lines), id)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return buf.String(), nil
}
