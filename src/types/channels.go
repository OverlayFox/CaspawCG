package types

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

var channelRangeRe = regexp.MustCompile(`^(\d+)-(\d+)$`)
var channelSingleRe = regexp.MustCompile(`^\d+$`)

// ParseChannelExpression parses a channel expression such as "1", "1-3", or "1,3-5"
// into a sorted, deduplicated slice of channel numbers. A blank (whitespace-only)
// input returns nil, meaning "all channels" to callers.
func ParseChannelExpression(input string) ([]int, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return nil, nil
	}

	seen := make(map[int]struct{})
	for _, token := range strings.Split(trimmed, ",") {
		part := strings.TrimSpace(token)

		if match := channelRangeRe.FindStringSubmatch(part); match != nil {
			start, _ := strconv.Atoi(match[1])
			end, _ := strconv.Atoi(match[2])
			if start > end {
				return nil, fmt.Errorf("invalid range %q", part)
			}
			for i := start; i <= end; i++ {
				seen[i] = struct{}{}
			}
			continue
		}

		if channelSingleRe.MatchString(part) {
			n, _ := strconv.Atoi(part)
			seen[n] = struct{}{}
			continue
		}

		return nil, fmt.Errorf("invalid token %q", part)
	}

	channels := make([]int, 0, len(seen))
	for n := range seen {
		channels = append(channels, n)
	}
	slices.Sort(channels)
	return channels, nil
}
