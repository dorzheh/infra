// Provides mount/umount ability

package utils

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	mountinfoFormat = "%d %d %d:%d %s %s %s"
)

// /proc/self/mountinfo representation
type procEntry struct {
	id, parent, major, minor int
	source, mountpoint, opts string
}

func parseMountTable() ([]*procEntry, error) {
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return parseInfoFile(f)
}

func parseInfoFile(r io.Reader) ([]*procEntry, error) {
	s   := bufio.NewScanner(r)
	out := []*procEntry{}
	for s.Scan() {
		if err := s.Err(); err != nil {
			return nil, err
		}
		p    := &procEntry{}
		text := s.Text()
		if _, err := fmt.Sscanf(text, mountinfoFormat,
			&p.id, &p.parent, &p.major, &p.minor,
			&p.source, &p.mountpoint, &p.opts); err != nil {
			return nil, fmt.Errorf("Scanning '%s' failed: %s", text, err)
		}
		out = append(out, p)
	}
	return out, nil
}

// Looks at /proc/self/mountinfo to determine of the specified
// mountpoint has been mounted
func Mounted(device, mountpoint string) (bool, error) {
	entries, err := parseMountTable()
	if err != nil {
		return false, err
	}
	// Search the table for the mountpoint
	for _, entry := range entries {
		if entry.mountpoint == mountpoint || strings.Contains(entry.opts, device) {
			return true, nil
		}
	}
	return false, nil
}
