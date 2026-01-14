package vcard

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// DedupIndex provides efficient contact deduplication
type DedupIndex struct {
	byPhone map[string][]*Contact
	byEmail map[string][]*Contact
	byName  map[string][]*Contact
}

// NewDedupIndex creates an index from a slice of contacts
func NewDedupIndex(contacts []*Contact) *DedupIndex {
	idx := &DedupIndex{
		byPhone: make(map[string][]*Contact),
		byEmail: make(map[string][]*Contact),
		byName:  make(map[string][]*Contact),
	}

	for _, c := range contacts {
		idx.Add(c)
	}

	return idx
}

// Add indexes a contact for dedup lookups
func (idx *DedupIndex) Add(c *Contact) {
	// Index by all phone suffixes
	for _, phone := range c.Phones {
		key := NormalizePhoneForDedup(phone)
		if key != "" {
			idx.byPhone[key] = append(idx.byPhone[key], c)
		}
	}

	// Index by all normalized emails
	for _, email := range c.Emails {
		key := NormalizeEmailForDedup(email)
		if key != "" {
			idx.byEmail[key] = append(idx.byEmail[key], c)
		}
	}

	// Index by normalized name
	key := NormalizeNameForDedup(c.DisplayName())
	if key != "" {
		idx.byName[key] = append(idx.byName[key], c)
	}
}

// FindDuplicates returns contacts that likely match the given contact
func (idx *DedupIndex) FindDuplicates(c *Contact) []*Contact {
	seen := make(map[*Contact]struct{})
	var matches []*Contact

	addMatch := func(candidate *Contact) {
		if candidate == c {
			return
		}
		if _, ok := seen[candidate]; ok {
			return
		}
		seen[candidate] = struct{}{}
		matches = append(matches, candidate)
	}

	// Strong match: same phone (suffix match handles country codes)
	for _, phone := range c.Phones {
		key := NormalizePhoneForDedup(phone)
		for _, candidate := range idx.byPhone[key] {
			addMatch(candidate)
		}
	}

	// Strong match: same email (after normalization)
	for _, email := range c.Emails {
		key := NormalizeEmailForDedup(email)
		for _, candidate := range idx.byEmail[key] {
			addMatch(candidate)
		}
	}

	// Weak match: same name - only if we also have partial overlap
	nameKey := NormalizeNameForDedup(c.DisplayName())
	for _, candidate := range idx.byName[nameKey] {
		if hasAnyOverlap(c, candidate) {
			addMatch(candidate)
		}
	}

	return matches
}

// IsDuplicate checks if contact matches any indexed contact
func (idx *DedupIndex) IsDuplicate(c *Contact) bool {
	return len(idx.FindDuplicates(c)) > 0
}

// NormalizePhoneForDedup aggressively normalizes phone for comparison.
// Uses last 9 digits to handle country code variations (+1, +34, etc.)
func NormalizePhoneForDedup(phone string) string {
	// Extract only digits
	var digits strings.Builder
	for _, r := range phone {
		if r >= '0' && r <= '9' {
			digits.WriteRune(r)
		}
	}

	d := digits.String()

	// Use last 9 digits as canonical form
	// This handles: +1-555-123-4567, 555-123-4567, 5551234567
	// All normalize to: 551234567
	const suffixLen = 9
	if len(d) >= suffixLen {
		return d[len(d)-suffixLen:]
	}

	// Short numbers kept as-is (local/extension numbers)
	if len(d) >= 6 {
		return d
	}

	return ""
}

// NormalizeEmailForDedup normalizes email for comparison.
// Handles: case, plus-addressing (user+tag@), googlemail vs gmail
func NormalizeEmailForDedup(email string) string {
	email = strings.ToLower(strings.TrimSpace(email))

	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 {
		return email
	}

	local, domain := parts[0], parts[1]

	// Strip plus-addressing (user+anything@domain → user@domain)
	if idx := strings.Index(local, "+"); idx != -1 {
		local = local[:idx]
	}

	// Normalize gmail variants
	if domain == "googlemail.com" {
		domain = "gmail.com"
	}

	// Gmail ignores dots in local part
	if domain == "gmail.com" {
		local = strings.ReplaceAll(local, ".", "")
	}

	return local + "@" + domain
}

// NormalizeNameForDedup normalizes name for comparison.
// Handles: case, accents, extra whitespace, common prefixes
func NormalizeNameForDedup(name string) string {
	// Lowercase
	name = strings.ToLower(name)

	// Remove accents (é → e, ñ → n, etc.)
	name = removeAccents(name)

	// Collapse whitespace
	name = strings.Join(strings.Fields(name), " ")

	// Remove common prefixes/suffixes that vary
	prefixes := []string{"dr ", "dr. ", "mr ", "mr. ", "mrs ", "mrs. ", "ms ", "ms. ", "prof ", "prof. "}
	for _, p := range prefixes {
		name = strings.TrimPrefix(name, p)
	}

	suffixes := []string{" jr", " jr.", " sr", " sr.", " ii", " iii", " iv", " phd", " md"}
	for _, s := range suffixes {
		name = strings.TrimSuffix(name, s)
	}

	return strings.TrimSpace(name)
}

// removeAccents strips diacritical marks from unicode text
func removeAccents(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, s)
	return result
}

// hasAnyOverlap checks if two contacts share any phone or email
func hasAnyOverlap(a, b *Contact) bool {
	// Check phone overlap
	aPhones := make(map[string]struct{})
	for _, p := range a.Phones {
		aPhones[NormalizePhoneForDedup(p)] = struct{}{}
	}
	for _, p := range b.Phones {
		if _, ok := aPhones[NormalizePhoneForDedup(p)]; ok {
			return true
		}
	}

	// Check email overlap
	aEmails := make(map[string]struct{})
	for _, e := range a.Emails {
		aEmails[NormalizeEmailForDedup(e)] = struct{}{}
	}
	for _, e := range b.Emails {
		if _, ok := aEmails[NormalizeEmailForDedup(e)]; ok {
			return true
		}
	}

	return false
}

// MatchStrength indicates how confident we are in a duplicate match
type MatchStrength int

const (
	MatchNone   MatchStrength = iota
	MatchWeak                 // Name only
	MatchMedium               // Name + partial data overlap
	MatchStrong               // Phone or email match
)

// CompareContacts returns the match strength between two contacts
func CompareContacts(a, b *Contact) MatchStrength {
	// Check for phone match (strongest signal)
	for _, pa := range a.Phones {
		keyA := NormalizePhoneForDedup(pa)
		if keyA == "" {
			continue
		}
		for _, pb := range b.Phones {
			if keyA == NormalizePhoneForDedup(pb) {
				return MatchStrong
			}
		}
	}

	// Check for email match (strong signal)
	for _, ea := range a.Emails {
		keyA := NormalizeEmailForDedup(ea)
		if keyA == "" {
			continue
		}
		for _, eb := range b.Emails {
			if keyA == NormalizeEmailForDedup(eb) {
				return MatchStrong
			}
		}
	}

	// Check name match
	nameA := NormalizeNameForDedup(a.DisplayName())
	nameB := NormalizeNameForDedup(b.DisplayName())

	// Don't match unnamed/empty contacts
	if nameA == "unnamed contact" || nameB == "unnamed contact" {
		return MatchNone
	}

	if nameA != "" && nameA == nameB {
		// Same name - check for any supporting evidence
		if a.Organization != "" && a.Organization == b.Organization {
			return MatchMedium
		}
		if a.Birthday != "" && a.Birthday == b.Birthday {
			return MatchMedium
		}
		return MatchWeak
	}

	return MatchNone
}
