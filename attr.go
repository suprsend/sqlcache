package sqlcache

import (
	"regexp"
	"strconv"
)

var (
	attrRegexp = regexp.MustCompile(`(@cache-ttl|@cache-max-rows|@cache-skip-empty-resultset) (\d+)`)
)

type attributes struct {
	ttl     int
	maxRows int
	//
	skipEmptyResultset bool
}

func getAttrs(query string) *attributes {
	matches := attrRegexp.FindAllStringSubmatch(query, 3)
	if !(len(matches) == 2 || len(matches) == 3) { // TODO: add better condition
		return nil
	}

	var attrs attributes
	for _, match := range matches {
		if len(match) != 3 {
			return nil
		}
		switch match[1] {
		case "@cache-ttl":
			ttl, _ := strconv.Atoi(match[2])
			attrs.ttl = ttl
		case "@cache-max-rows":
			maxRows, _ := strconv.Atoi(match[2])
			attrs.maxRows = maxRows
		case "@cache-skip-empty-resultset":
			skipVal, _ := strconv.Atoi(match[2])
			attrs.skipEmptyResultset = skipVal > 0
		}
	}

	return &attrs
}
