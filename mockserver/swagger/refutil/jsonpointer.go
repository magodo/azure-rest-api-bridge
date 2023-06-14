package refutil

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-openapi/jsonpointer"
)

type offsetProgress struct {
	decodedToken []string
	tkIndex      int
	offset       int64
	finished     bool
}

// return map[p.String()]offset
// use a map to keep progress of every pointer and do the parse in one pass.
// use `-1` as error result.
func JSONPointerOffsetMulti(ps []jsonpointer.Pointer, document string) (map[string]int64, error) {
	dec := json.NewDecoder(strings.NewReader(document))

	progressMap := make(map[string]offsetProgress, 0)
	for _, p := range ps {
		t := p.DecodedTokens()
		if len(t) > 0 {
			progressMap[p.String()] = offsetProgress{
				decodedToken: t,
				tkIndex:      0,
				offset:       0,
				finished:     false,
			}
		}
	}

	tks := make(map[string]string, 0)
	for k, p := range progressMap {
		if len(p.decodedToken) > 0 {
			tks[k] = p.decodedToken[0]
		}
	}

	tk, err := dec.Token()
	if err != nil {
		return nil, err
	}
	switch tk := tk.(type) {
	case json.Delim:
		switch tk {
		case '{':
			progressMap, err = offsetSingleObjectForMulti(dec, progressMap, 0)
			if err != nil {
				return nil, err
			}
		case '[':
			progressMap, err = offsetSingleArrayForMulti(dec, progressMap, 0)
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("invalid token %#v", tk)
		}
	default:
	}

	result := make(map[string]int64)
	errKey := make([]string, 0)
	for k, p := range progressMap {
		if p.finished {
			result[k] = p.offset
		} else {
			result[k] = 0 // not found
			errKey = append(errKey, k)
		}
	}

	if len(errKey) > 0 {
		return nil, fmt.Errorf("token reference %q not found", errKey)
	}

	return result, nil
}

func offsetSingleObjectForMulti(dec *json.Decoder, progressMap map[string]offsetProgress, depth int) (map[string]offsetProgress, error) {
	tks := make(map[string]string, 0)
	for k, p := range progressMap {
		if p.tkIndex == depth {
			tks[k] = p.decodedToken[p.tkIndex]
		}
	}
	if len(tks) == 0 {
		return progressMap, nil
	}

	// used for next level call
	var nextLevelProgress map[string]offsetProgress
	// first token after a key needs special handling
	isFirstTokenAfterHit := false
	for dec.More() {
		tmpProgressMap := nextLevelProgress
		nextLevelProgress = make(map[string]offsetProgress, 0)
		hit := false
		offset := dec.InputOffset()
		tk, err := dec.Token()
		if err != nil {
			return nil, err
		}
		switch tk := tk.(type) {
		case json.Delim:
			switch tk {
			case '{':
				if isFirstTokenAfterHit {
					nextLevelProgress, err = offsetSingleObjectForMulti(dec, tmpProgressMap, depth+1)
					if err != nil {
						return nil, err
					}
				} else if err := drainSingle(dec); err != nil {
					return nil, err
				}
			case '[':
				if isFirstTokenAfterHit {
					nextLevelProgress, err = offsetSingleArrayForMulti(dec, tmpProgressMap, depth+1)
					if err != nil {
						return nil, err
					}
				} else if err := drainSingle(dec); err != nil {
					return nil, err
				}
			}
		case string:
			for k, t := range tks {
				if tk == t {
					p := progressMap[k]
					p.offset = offset
					if p.tkIndex == len(p.decodedToken)-1 {
						p.finished = true
					}
					if p.tkIndex < len(p.decodedToken)-1 {
						p.tkIndex++
					}
					progressMap[k] = p
					nextLevelProgress[k] = p
					isFirstTokenAfterHit = true
					hit = true
				}
			}
		default:
		}
		for k, p := range nextLevelProgress {
			progressMap[k] = p
		}
		isFirstTokenAfterHit = hit
	}
	// Consumes the ending delim
	if depth > 0 {
		_, err := dec.Token()
		if err != nil {
			return nil, err
		}
	}
	return progressMap, nil
}

func offsetSingleArrayForMulti(dec *json.Decoder, progressMap map[string]offsetProgress, depth int) (map[string]offsetProgress, error) {
	idxs := make(map[string]string, 0)
	for k, p := range progressMap {
		if p.tkIndex == depth {
			idxs[k] = p.decodedToken[p.tkIndex]
		}
	}
	if len(idxs) == 0 {
		return progressMap, nil
	}

	var tmpProgressMap map[string]offsetProgress
	for i := 0; dec.More(); i++ {
		hit := false
		tmpProgressMap = make(map[string]offsetProgress, 0)

		offset := dec.InputOffset()
		tk, err := dec.Token()
		if err != nil {
			return nil, err
		}
		for k, idx := range idxs {
			idx, err := strconv.Atoi(idx)
			if err != nil {
				return nil, err
			}
			if i == idx {
				p := progressMap[k]
				p.offset = offset
				if p.tkIndex == len(p.decodedToken)-1 {
					p.finished = true
				}
				if p.tkIndex < len(p.decodedToken)-1 {
					p.tkIndex++
				}
				progressMap[k] = p
				tmpProgressMap[k] = p
				hit = true
			}
		}

		switch tk := tk.(type) {
		case json.Delim:
			switch tk {
			case '{':
				if hit {
					tmpProgressMap, err = offsetSingleObjectForMulti(dec, tmpProgressMap, depth+1)
					if err != nil {
						return nil, err
					}
				} else if err := drainSingle(dec); err != nil {
					return nil, err
				}
			case '[':
				if hit {
					tmpProgressMap, err = offsetSingleArrayForMulti(dec, tmpProgressMap, depth+1)
					if err != nil {
						return nil, err
					}
				} else if err := drainSingle(dec); err != nil {
					return nil, err
				}
			}
		}
		for t, p := range tmpProgressMap {
			progressMap[t] = p
		}
	}

	// Consumes the ending delimF
	if depth > 0 {
		_, err := dec.Token()
		if err != nil {
			return nil, err
		}
	}
	return progressMap, nil
}

// drainSingle drains a single level of object or array.
// The decoder has to guarantee the begining delim (i.e. '{' or '[') has been consumed.
func drainSingle(dec *json.Decoder) error {
	for dec.More() {
		tk, err := dec.Token()
		if err != nil {
			return err
		}
		switch tk := tk.(type) {
		case json.Delim:
			switch tk {
			case '{':
				if err := drainSingle(dec); err != nil {
					return err
				}
			case '[':
				if err := drainSingle(dec); err != nil {
					return err
				}
			}
		}
	}
	// Consumes the ending delim
	if _, err := dec.Token(); err != nil {
		return err
	}
	return nil
}
