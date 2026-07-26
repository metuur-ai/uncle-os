package ids

import (
	"sort"
	"strings"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// Suggest is suggest_ids (`bin/company-os:1220-1232`): up to three registered
// IDs closest to an unknown one (GPF-R-2.3). `governance explain` is its only
// caller today (`:365`).
//
// The match runs on the LOCAL NAME — everything after "://" — because every
// candidate shares its scheme once the scheme filter has been applied, and
// comparing full URIs would score that shared prefix as similarity. Full
// canonical IDs come back out, which is what a user can act on.
//
// scheme is optional; "" compares against the whole registry.
func Suggest(ws *workspace.Workspace, unknown, scheme string) ([]string, error) {
	entries, err := Load(ws)
	if err != nil {
		return nil, err
	}

	// dict.setdefault: the FIRST full id wins for a repeated local name, and the
	// candidate order is registry order, which is what breaks score ties.
	var bare []string
	full := map[string]string{}
	for _, e := range entries {
		if scheme != "" && !strings.HasPrefix(e.ID, scheme+"://") {
			continue
		}
		name := localName(e.ID)
		if _, seen := full[name]; seen {
			continue
		}
		full[name] = e.ID
		bare = append(bare, name)
	}

	hits := closeMatches(localName(unknown), bare, 3, 0.3)
	out := make([]string, 0, len(hits))
	for _, h := range hits {
		out = append(out, full[h])
	}
	return out, nil
}

// localName is `s.split("://")[-1]`: the last segment, or the whole string when
// there is no separator.
func localName(s string) string {
	if i := strings.LastIndex(s, "://"); i >= 0 {
		return s[i+len("://"):]
	}
	return s
}

// closeMatches is difflib.get_close_matches(word, possibilities, n, cutoff).
//
// It is transliterated rather than approximated because the scores decide what a
// user is told to type next, and no Go standard-library ratio agrees with
// SequenceMatcher's. The three-stage filter is kept even though it is only a
// speed optimisation in Python: real_quick_ratio and quick_ratio are upper
// bounds on ratio, so dropping them cannot change the result, but keeping them
// keeps this readable against difflib.py side by side.
//
// Ordering is heapq.nlargest over (score, candidate) TUPLES, so a score tie is
// broken by the candidate string descending, not by input order.
func closeMatches(word string, possibilities []string, n int, cutoff float64) []string {
	b := []rune(word)
	type scored struct {
		ratio float64
		text  string
	}
	var result []scored
	for _, x := range possibilities {
		a := []rune(x)
		if realQuickRatio(a, b) < cutoff {
			continue
		}
		if quickRatio(a, b) < cutoff {
			continue
		}
		r := ratio(a, b)
		if r < cutoff {
			continue
		}
		result = append(result, scored{r, x})
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].ratio != result[j].ratio {
			return result[i].ratio > result[j].ratio
		}
		return result[i].text > result[j].text
	})
	if len(result) > n {
		result = result[:n]
	}
	out := make([]string, 0, len(result))
	for _, r := range result {
		out = append(out, r.text)
	}
	return out
}

// calculateRatio is difflib._calculate_ratio.
func calculateRatio(matches, length int) float64 {
	if length == 0 {
		return 1.0
	}
	return 2.0 * float64(matches) / float64(length)
}

// realQuickRatio is an upper bound based on length alone.
func realQuickRatio(a, b []rune) float64 {
	la, lb := len(a), len(b)
	m := la
	if lb < m {
		m = lb
	}
	return calculateRatio(m, la+lb)
}

// quickRatio is an upper bound based on the multiset intersection of a and b.
func quickRatio(a, b []rune) float64 {
	fullbcount := map[rune]int{}
	for _, r := range b {
		fullbcount[r]++
	}
	avail := map[rune]int{}
	matches := 0
	for _, r := range a {
		numb, seen := avail[r]
		if !seen {
			numb = fullbcount[r]
		}
		avail[r] = numb - 1
		if numb > 0 {
			matches++
		}
	}
	return calculateRatio(matches, len(a)+len(b))
}

// autojunkMin is SequenceMatcher.__chain_b's `n >= 200` (difflib.py), where n is
// len(seq2).
//
// A previous note here called autojunk inert "because no canonical ID is that
// long". That was wrong about which sequence it measures: get_close_matches sets
// seq2 to the WORD, and the word is the unknown id a user typed at
// `governance explain`, not a registry entry. Measured against CPython's difflib:
// identical results for every word under 200 runes, and a divergence on 200 of
// 200 sampled words at or above it — Python returning [] where this returned up
// to three suggestions.
const autojunkMin = 200

// ratio is SequenceMatcher(None, a, b).ratio(). With no junk the total matched
// size is the sum over get_matching_blocks, and the merge pass in that function
// does not change the sum — so the recursion below is enough and the block list
// is not built.
func ratio(a, b []rune) float64 {
	b2j := map[rune][]int{}
	for j, r := range b {
		b2j[r] = append(b2j[r], j)
	}
	// __chain_b's autojunk pass: a character appearing in more than 1% of a long
	// seq2 becomes "popular" and is dropped from the index, so the DP below can
	// no longer start a match on it. It stays visible to find_longest_match's
	// extension loops, which is why those are no longer omitted.
	if len(b) >= autojunkMin {
		ncut := len(b)/100 + 1
		for r, idxs := range b2j {
			if len(idxs) > ncut {
				delete(b2j, r)
			}
		}
	}

	matches := 0
	type span struct{ alo, ahi, blo, bhi int }
	queue := []span{{0, len(a), 0, len(b)}}
	for len(queue) > 0 {
		q := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		i, j, k := longestMatch(a, b, b2j, q.alo, q.ahi, q.blo, q.bhi)
		if k == 0 {
			continue
		}
		matches += k
		if q.alo < i && q.blo < j {
			queue = append(queue, span{q.alo, i, q.blo, j})
		}
		if i+k < q.ahi && j+k < q.bhi {
			queue = append(queue, span{i + k, q.ahi, j + k, q.bhi})
		}
	}
	return calculateRatio(matches, len(a)+len(b))
}

// longestMatch is SequenceMatcher.find_longest_match. Its last two extension
// loops are omitted: both are guarded by isbjunk, and bjunk is the isjunk-based
// set, which is empty here because get_close_matches passes isjunk=None. The
// first two are NOT omitted — `not isbjunk(...)` is vacuously true, so they run,
// and they are the only thing that can grow a match across a character autojunk
// pruned out of b2j.
func longestMatch(a, b []rune, b2j map[rune][]int, alo, ahi, blo, bhi int) (int, int, int) {
	besti, bestj, bestsize := alo, blo, 0
	j2len := map[int]int{}
	for i := alo; i < ahi; i++ {
		newj2len := map[int]int{}
		for _, j := range b2j[a[i]] {
			if j < blo {
				continue
			}
			if j >= bhi {
				break
			}
			k := j2len[j-1] + 1
			newj2len[j] = k
			if k > bestsize {
				besti, bestj, bestsize = i-k+1, j-k+1, k
			}
		}
		j2len = newj2len
	}
	for besti > alo && bestj > blo && a[besti-1] == b[bestj-1] {
		besti, bestj, bestsize = besti-1, bestj-1, bestsize+1
	}
	for besti+bestsize < ahi && bestj+bestsize < bhi &&
		a[besti+bestsize] == b[bestj+bestsize] {
		bestsize++
	}
	return besti, bestj, bestsize
}
