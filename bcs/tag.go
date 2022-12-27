package bcs

import (
	"fmt"
	"strings"
)

const tagName = "bcs"

const (
	tagValue_Optional int64 = 1 << iota // optional
	tagValue_Ignore                     // -
)

func parseTagValue(tag string) (int64, error) {
	var r int64
	tagSegs := strings.Split(tag, ",")
	for _, seg := range tagSegs {
		seg := strings.TrimSpace(seg)
		if seg == "" {
			continue
		}
		switch seg {
		case "optional":
			r |= tagValue_Optional
		case "-":
			return tagValue_Ignore, nil
		default:
			return 0, fmt.Errorf("unknown tag: %s in %s", seg, tag)
		}
	}

	return r, nil
}
