package vcard

import (
	"testing"
)

// =============================================================================
// Phone Normalization Tests
// =============================================================================

func TestNormalizePhoneForDedup(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Country code variations
		{"US with +1", "+1-555-123-4567", "551234567"},
		{"US with 1 no plus", "1-555-123-4567", "551234567"},
		{"US without country code", "555-123-4567", "551234567"},
		{"US bare digits with 1", "15551234567", "551234567"},
		{"US bare digits without 1", "5551234567", "551234567"},

		// International formats
		{"Spain +34", "+34 612 345 678", "612345678"},
		{"Spain 0034", "0034 612 345 678", "612345678"},
		{"UK +44", "+44 20 7123 4567", "071234567"},      // last 9 of 442071234567
		{"Germany +49", "+49 30 12345678", "012345678"},  // last 9 of 493012345678
		{"France +33", "+33 1 23 45 67 89", "123456789"},

		// Format variations
		{"With parentheses", "(555) 123-4567", "551234567"},
		{"With dots", "555.123.4567", "551234567"},
		{"With spaces", "555 123 4567", "551234567"},
		{"Mixed separators", "+1 (555) 123-4567", "551234567"},
		{"No separators", "5551234567", "551234567"},

		// Edge cases
		{"Short local number", "123456", "123456"},
		{"Extension style", "1234567", "1234567"},
		{"Too short - 5 digits", "12345", ""},
		{"Too short - 4 digits", "1234", ""},
		{"Empty", "", ""},
		{"Only separators", "---", ""},
		{"Letters mixed in", "555-ABC-4567", "5554567"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizePhoneForDedup(tt.input)
			if got != tt.expected {
				t.Errorf("NormalizePhoneForDedup(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestPhoneMatchingAcrossFormats(t *testing.T) {
	// All these should normalize to the same value
	equivalentPhones := []string{
		"+1-555-123-4567",
		"1-555-123-4567",
		"555-123-4567",
		"(555) 123-4567",
		"555.123.4567",
		"5551234567",
		"+1 555 123 4567",
	}

	normalized := NormalizePhoneForDedup(equivalentPhones[0])
	for _, phone := range equivalentPhones[1:] {
		got := NormalizePhoneForDedup(phone)
		if got != normalized {
			t.Errorf("Phone %q normalized to %q, expected %q (same as %q)",
				phone, got, normalized, equivalentPhones[0])
		}
	}
}

// =============================================================================
// Email Normalization Tests
// =============================================================================

func TestNormalizeEmailForDedup(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Basic normalization
		{"lowercase", "John.Doe@Example.COM", "john.doe@example.com"},
		{"trim whitespace", "  john@example.com  ", "john@example.com"},
		{"already normalized", "john@example.com", "john@example.com"},

		// Plus addressing
		{"plus addressing simple", "john+newsletter@example.com", "john@example.com"},
		{"plus addressing complex", "john+work+2024@example.com", "john@example.com"},
		{"plus at start", "+test@example.com", "@example.com"},

		// Gmail-specific handling
		{"gmail dots", "j.o.h.n@gmail.com", "john@gmail.com"},
		{"gmail plus and dots", "j.o.h.n+spam@gmail.com", "john@gmail.com"},
		{"googlemail to gmail", "john@googlemail.com", "john@gmail.com"},
		{"googlemail with dots", "j.o.h.n@googlemail.com", "john@gmail.com"},

		// Non-gmail should preserve dots
		{"non-gmail dots preserved", "j.o.h.n@example.com", "j.o.h.n@example.com"},
		{"yahoo dots preserved", "j.o.h.n@yahoo.com", "j.o.h.n@yahoo.com"},
		{"outlook dots preserved", "j.o.h.n@outlook.com", "j.o.h.n@outlook.com"},

		// Edge cases
		{"no @ symbol", "notanemail", "notanemail"},
		{"empty local part", "@example.com", "@example.com"},
		{"empty domain", "john@", "john@"},
		{"multiple @", "john@doe@example.com", "john@doe@example.com"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeEmailForDedup(tt.input)
			if got != tt.expected {
				t.Errorf("NormalizeEmailForDedup(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestEmailMatchingAcrossVariations(t *testing.T) {
	// All these Gmail addresses should be equivalent
	gmailVariants := []string{
		"johndoe@gmail.com",
		"john.doe@gmail.com",
		"j.o.h.n.d.o.e@gmail.com",
		"johndoe+work@gmail.com",
		"john.doe+newsletter@gmail.com",
		"JOHNDOE@GMAIL.COM",
		"johndoe@googlemail.com",
	}

	normalized := NormalizeEmailForDedup(gmailVariants[0])
	for _, email := range gmailVariants[1:] {
		got := NormalizeEmailForDedup(email)
		if got != normalized {
			t.Errorf("Email %q normalized to %q, expected %q (same as %q)",
				email, got, normalized, gmailVariants[0])
		}
	}
}

// =============================================================================
// Name Normalization Tests
// =============================================================================

func TestNormalizeNameForDedup(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Basic normalization
		{"lowercase", "John Doe", "john doe"},
		{"uppercase", "JOHN DOE", "john doe"},
		{"mixed case", "JoHn DoE", "john doe"},

		// Whitespace handling
		{"extra whitespace", "  John   Doe  ", "john doe"},
		{"tabs", "John\tDoe", "john doe"},
		{"newlines", "John\nDoe", "john doe"},
		{"multiple spaces between", "John    Doe", "john doe"},

		// Prefix removal
		{"Dr.", "Dr. John Doe", "john doe"},
		{"Dr no dot", "Dr John Doe", "john doe"},
		{"Mr.", "Mr. John Doe", "john doe"},
		{"Mrs.", "Mrs. Jane Doe", "jane doe"},
		{"Ms.", "Ms. Jane Doe", "jane doe"},
		{"Prof.", "Prof. John Doe", "john doe"},
		{"Professor", "Professor John Doe", "professor john doe"}, // not stripped - too long

		// Suffix removal
		{"Jr.", "John Doe Jr.", "john doe"},
		{"Jr no dot", "John Doe Jr", "john doe"},
		{"Sr.", "John Doe Sr.", "john doe"},
		{"II", "John Doe II", "john doe"},
		{"III", "John Doe III", "john doe"},
		{"IV", "John Doe IV", "john doe"},
		{"PhD", "John Doe PhD", "john doe"},
		{"MD", "John Doe MD", "john doe"},

		// Combined prefix and suffix
		{"Dr and PhD", "Dr. John Doe PhD", "john doe"},
		{"Mr and Jr", "Mr. John Doe Jr.", "john doe"},

		// Accented characters (diacritics)
		{"Spanish ñ", "José García", "jose garcia"},
		{"German umlaut", "Müller", "muller"},
		{"French accent", "François", "francois"},
		{"Czech háček", "Dvořák", "dvorak"},
		{"Nordic ø", "Søren", "søren"},  // ø is not a combining character, kept as-is
		{"Portuguese ã", "João", "joao"},
		{"Multiple accents", "Ñoño Müller-García", "nono muller-garcia"},

		// Edge cases
		{"empty", "", ""},
		{"only whitespace", "   ", ""},
		{"only prefix", "Dr.", "dr."},  // prefix stripping only works with space after
		{"single name", "John", "john"},
		{"hyphenated", "Mary-Jane Watson", "mary-jane watson"},
		{"apostrophe", "O'Connor", "o'connor"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeNameForDedup(tt.input)
			if got != tt.expected {
				t.Errorf("NormalizeNameForDedup(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestNameMatchingAcrossVariations(t *testing.T) {
	// All these should be considered the same person
	nameVariants := []string{
		"John Doe",
		"john doe",
		"JOHN DOE",
		"  John   Doe  ",
		"Dr. John Doe",
		"John Doe Jr.",
		"Dr. John Doe PhD",
	}

	normalized := NormalizeNameForDedup(nameVariants[0])
	for _, name := range nameVariants[1:] {
		got := NormalizeNameForDedup(name)
		if got != normalized {
			t.Errorf("Name %q normalized to %q, expected %q (same as %q)",
				name, got, normalized, nameVariants[0])
		}
	}
}

// =============================================================================
// DedupIndex - Phone Matching Tests
// =============================================================================

func TestDedupIndex_PhoneMatch(t *testing.T) {
	existing := []*Contact{
		{FormattedName: "John Doe", Phones: []string{"+1-555-123-4567"}},
	}
	idx := NewDedupIndex(existing)

	// Same phone, different format
	newContact := &Contact{
		FormattedName: "Johnny Doe",
		Phones:        []string{"555-123-4567"},
	}

	if !idx.IsDuplicate(newContact) {
		t.Error("Expected phone match to be detected as duplicate")
	}
}

func TestDedupIndex_PhoneMatchWithCountryCode(t *testing.T) {
	tests := []struct {
		name         string
		existingPhone string
		newPhone     string
		shouldMatch  bool
	}{
		{"US +1 vs bare", "+1-555-123-4567", "555-123-4567", true},
		{"bare vs US +1", "555-123-4567", "+1-555-123-4567", true},
		{"Spain +34 vs bare", "+34 612 345 678", "612 345 678", true},
		{"Different numbers", "+1-555-123-4567", "555-987-6543", false},
		{"Partial overlap", "+1-555-123-4567", "123-4567", false}, // too short
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			existing := []*Contact{{Phones: []string{tt.existingPhone}}}
			idx := NewDedupIndex(existing)

			newContact := &Contact{Phones: []string{tt.newPhone}}
			got := idx.IsDuplicate(newContact)

			if got != tt.shouldMatch {
				t.Errorf("Phone %q vs %q: got duplicate=%v, want %v",
					tt.existingPhone, tt.newPhone, got, tt.shouldMatch)
			}
		})
	}
}

func TestDedupIndex_MultiplePhones(t *testing.T) {
	// Contact has multiple phones, match on any
	existing := []*Contact{
		{
			FormattedName: "John Doe",
			Phones:        []string{"+1-555-111-1111", "+1-555-222-2222", "+1-555-333-3333"},
		},
	}
	idx := NewDedupIndex(existing)

	tests := []struct {
		name        string
		phones      []string
		shouldMatch bool
	}{
		{"match first", []string{"555-111-1111"}, true},
		{"match second", []string{"555-222-2222"}, true},
		{"match third", []string{"555-333-3333"}, true},
		{"match any of multiple", []string{"555-999-9999", "555-222-2222"}, true},
		{"no match", []string{"555-444-4444"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newContact := &Contact{Phones: tt.phones}
			got := idx.IsDuplicate(newContact)
			if got != tt.shouldMatch {
				t.Errorf("got duplicate=%v, want %v", got, tt.shouldMatch)
			}
		})
	}
}

// =============================================================================
// DedupIndex - Email Matching Tests
// =============================================================================

func TestDedupIndex_EmailMatch(t *testing.T) {
	existing := []*Contact{
		{FormattedName: "John Doe", Emails: []string{"john.doe@gmail.com"}},
	}
	idx := NewDedupIndex(existing)

	// Same email with plus addressing and dots
	newContact := &Contact{
		FormattedName: "J Doe",
		Emails:        []string{"johndoe+work@gmail.com"},
	}

	if !idx.IsDuplicate(newContact) {
		t.Error("Expected email match to be detected as duplicate")
	}
}

func TestDedupIndex_EmailMatchVariations(t *testing.T) {
	tests := []struct {
		name          string
		existingEmail string
		newEmail      string
		shouldMatch   bool
	}{
		// Gmail variations
		{"gmail dots", "johndoe@gmail.com", "john.doe@gmail.com", true},
		{"gmail plus", "johndoe@gmail.com", "johndoe+work@gmail.com", true},
		{"gmail dots and plus", "johndoe@gmail.com", "j.o.h.n.d.o.e+spam@gmail.com", true},
		{"googlemail", "johndoe@gmail.com", "johndoe@googlemail.com", true},

		// Case insensitivity
		{"case insensitive", "John.Doe@Example.com", "john.doe@example.com", true},
		{"all caps", "JOHN@EXAMPLE.COM", "john@example.com", true},

		// Different emails
		{"different local", "john@example.com", "jane@example.com", false},
		{"different domain", "john@example.com", "john@other.com", false},

		// Non-gmail should NOT strip dots
		{"non-gmail dots differ", "j.doe@example.com", "jdoe@example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			existing := []*Contact{{Emails: []string{tt.existingEmail}}}
			idx := NewDedupIndex(existing)

			newContact := &Contact{Emails: []string{tt.newEmail}}
			got := idx.IsDuplicate(newContact)

			if got != tt.shouldMatch {
				t.Errorf("Email %q vs %q: got duplicate=%v, want %v",
					tt.existingEmail, tt.newEmail, got, tt.shouldMatch)
			}
		})
	}
}

func TestDedupIndex_MultipleEmails(t *testing.T) {
	existing := []*Contact{
		{
			FormattedName: "John Doe",
			Emails:        []string{"john@work.com", "john@personal.com", "johndoe@gmail.com"},
		},
	}
	idx := NewDedupIndex(existing)

	tests := []struct {
		name        string
		emails      []string
		shouldMatch bool
	}{
		{"match first", []string{"john@work.com"}, true},
		{"match second", []string{"john@personal.com"}, true},
		{"match gmail with dots", []string{"john.doe@gmail.com"}, true},
		{"match any of multiple", []string{"other@example.com", "john@work.com"}, true},
		{"no match", []string{"other@example.com"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newContact := &Contact{Emails: tt.emails}
			got := idx.IsDuplicate(newContact)
			if got != tt.shouldMatch {
				t.Errorf("got duplicate=%v, want %v", got, tt.shouldMatch)
			}
		})
	}
}

// =============================================================================
// DedupIndex - Name Matching Tests
// =============================================================================

func TestDedupIndex_NameOnlyNoMatch(t *testing.T) {
	existing := []*Contact{
		{FormattedName: "John Doe", Phones: []string{"111-111-1111"}},
	}
	idx := NewDedupIndex(existing)

	// Same name but different phone/email - should NOT match
	newContact := &Contact{
		FormattedName: "John Doe",
		Phones:        []string{"222-222-2222"},
	}

	if idx.IsDuplicate(newContact) {
		t.Error("Name-only match without data overlap should not be duplicate")
	}
}

func TestDedupIndex_NameWithOverlap(t *testing.T) {
	existing := []*Contact{
		{FormattedName: "John Doe", Phones: []string{"555-123-4567"}, Emails: []string{"john@example.com"}},
	}
	idx := NewDedupIndex(existing)

	// Same name AND same phone = duplicate
	newContact := &Contact{
		FormattedName: "John Doe",
		Phones:        []string{"555-123-4567"},
	}

	if !idx.IsDuplicate(newContact) {
		t.Error("Name match with phone overlap should be duplicate")
	}
}

func TestDedupIndex_PartialNames(t *testing.T) {
	// These are tricky cases - partial name matches
	tests := []struct {
		name          string
		existingName  string
		newName       string
		existingPhone string
		newPhone      string
		shouldMatch   bool
	}{
		// Same phone should match regardless of name
		{"different names same phone", "John Doe", "Johnny D", "555-123-4567", "555-123-4567", true},
		{"nickname same phone", "Robert Smith", "Bob Smith", "555-123-4567", "555-123-4567", true},
		{"formal vs casual same phone", "Dr. John Doe PhD", "John", "555-123-4567", "555-123-4567", true},

		// Different phone, name variations - should NOT match (too risky)
		{"similar names diff phone", "John Doe", "John D", "555-111-1111", "555-222-2222", false},
		{"first name only diff phone", "John Doe", "John", "555-111-1111", "555-222-2222", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			existing := []*Contact{{FormattedName: tt.existingName, Phones: []string{tt.existingPhone}}}
			idx := NewDedupIndex(existing)

			newContact := &Contact{FormattedName: tt.newName, Phones: []string{tt.newPhone}}
			got := idx.IsDuplicate(newContact)

			if got != tt.shouldMatch {
				t.Errorf("got duplicate=%v, want %v", got, tt.shouldMatch)
			}
		})
	}
}

func TestDedupIndex_NameNormalization(t *testing.T) {
	existing := []*Contact{
		{FormattedName: "Dr. José García PhD", Phones: []string{"555-123-4567"}},
	}
	idx := NewDedupIndex(existing)

	tests := []struct {
		name        string
		newName     string
		samePhone   bool
		shouldMatch bool
	}{
		// Same normalized name with same phone = match
		{"lowercase same phone", "jose garcia", true, true},
		{"no titles same phone", "José García", true, true},
		{"no accents same phone", "Jose Garcia", true, true},
		{"uppercase same phone", "JOSE GARCIA", true, true},

		// Same normalized name, different phone = no match (name only not enough)
		{"lowercase diff phone", "jose garcia", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			phone := "555-123-4567"
			if !tt.samePhone {
				phone = "555-999-9999"
			}
			newContact := &Contact{FormattedName: tt.newName, Phones: []string{phone}}
			got := idx.IsDuplicate(newContact)

			if got != tt.shouldMatch {
				t.Errorf("got duplicate=%v, want %v", got, tt.shouldMatch)
			}
		})
	}
}

// =============================================================================
// DedupIndex - Batch/Import Deduplication Tests
// =============================================================================

func TestDedupIndex_BatchDedup(t *testing.T) {
	idx := NewDedupIndex(nil)

	contacts := []Contact{
		{FormattedName: "John Doe", Phones: []string{"555-123-4567"}},
		{FormattedName: "Johnny Doe", Phones: []string{"+1-555-123-4567"}}, // duplicate
		{FormattedName: "Jane Smith", Emails: []string{"jane@example.com"}},
		{FormattedName: "J Smith", Emails: []string{"jane+work@example.com"}}, // duplicate
	}

	var imported []string
	for i := range contacts {
		c := &contacts[i]
		if !idx.IsDuplicate(c) {
			imported = append(imported, c.FormattedName)
			idx.Add(c)
		}
	}

	if len(imported) != 2 {
		t.Errorf("Expected 2 unique contacts, got %d: %v", len(imported), imported)
	}
}

func TestDedupIndex_BatchDedupComplex(t *testing.T) {
	// Simulate importing a messy contact list with duplicates
	idx := NewDedupIndex(nil)

	contacts := []Contact{
		// First person - multiple variations
		{FormattedName: "John Doe", Phones: []string{"+1-555-111-1111"}, Emails: []string{"john@example.com"}},
		{FormattedName: "Johnny Doe", Phones: []string{"555-111-1111"}},                            // dup: same phone
		{FormattedName: "J. Doe", Emails: []string{"john+work@example.com"}},                       // dup: same email (plus addr)
		{FormattedName: "Dr. John Doe", Phones: []string{"555-111-1111"}, Emails: []string{"john@other.com"}}, // dup: same phone

		// Second person
		{FormattedName: "Jane Smith", Phones: []string{"+44 20 7123 4567"}, Emails: []string{"jane@gmail.com"}},
		{FormattedName: "Jane Smith", Phones: []string{"020 7123 4567"}},    // dup: same phone (UK format)
		{FormattedName: "J Smith", Emails: []string{"j.a.n.e@gmail.com"}},   // dup: same email (gmail dots)

		// Third person - actually unique
		{FormattedName: "Bob Johnson", Phones: []string{"555-333-3333"}, Emails: []string{"bob@example.com"}},

		// Fourth person - edge case: same name as first but different everything else
		{FormattedName: "John Doe", Phones: []string{"555-999-9999"}, Emails: []string{"different.john@other.com"}},
	}

	var imported []string
	for i := range contacts {
		c := &contacts[i]
		if !idx.IsDuplicate(c) {
			imported = append(imported, c.FormattedName)
			idx.Add(c)
		}
	}

	// Should have: John Doe, Jane Smith, Bob Johnson, John Doe (the second one)
	if len(imported) != 4 {
		t.Errorf("Expected 4 unique contacts, got %d: %v", len(imported), imported)
	}
}

func TestDedupIndex_EmptyContacts(t *testing.T) {
	idx := NewDedupIndex(nil)

	// Contact with no identifying info
	emptyContact := &Contact{FormattedName: ""}
	if idx.IsDuplicate(emptyContact) {
		t.Error("Empty contact should not be flagged as duplicate")
	}

	// Contact with only name, no phone/email
	nameOnly := &Contact{FormattedName: "John Doe"}
	if idx.IsDuplicate(nameOnly) {
		t.Error("Name-only contact with empty index should not be duplicate")
	}

	// Add it and check another name-only doesn't match
	idx.Add(nameOnly)
	anotherNameOnly := &Contact{FormattedName: "John Doe"}
	if idx.IsDuplicate(anotherNameOnly) {
		t.Error("Two name-only contacts should not be considered duplicates")
	}
}

func TestDedupIndex_CrossFieldMatching(t *testing.T) {
	// Test that phone match works even when emails differ, and vice versa
	existing := []*Contact{
		{
			FormattedName: "John Doe",
			Phones:        []string{"555-123-4567"},
			Emails:        []string{"john@example.com"},
		},
	}
	idx := NewDedupIndex(existing)

	tests := []struct {
		name        string
		contact     *Contact
		shouldMatch bool
	}{
		{
			name: "same phone different email",
			contact: &Contact{
				FormattedName: "John D",
				Phones:        []string{"555-123-4567"},
				Emails:        []string{"different@example.com"},
			},
			shouldMatch: true,
		},
		{
			name: "different phone same email",
			contact: &Contact{
				FormattedName: "J Doe",
				Phones:        []string{"555-999-9999"},
				Emails:        []string{"john@example.com"},
			},
			shouldMatch: true,
		},
		{
			name: "different phone different email",
			contact: &Contact{
				FormattedName: "John Doe",
				Phones:        []string{"555-999-9999"},
				Emails:        []string{"different@example.com"},
			},
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := idx.IsDuplicate(tt.contact)
			if got != tt.shouldMatch {
				t.Errorf("got duplicate=%v, want %v", got, tt.shouldMatch)
			}
		})
	}
}

// =============================================================================
// CompareContacts - Match Strength Tests
// =============================================================================

func TestCompareContacts(t *testing.T) {
	tests := []struct {
		name     string
		a, b     *Contact
		expected MatchStrength
	}{
		{
			name:     "phone match",
			a:        &Contact{Phones: []string{"+1-555-123-4567"}},
			b:        &Contact{Phones: []string{"555-123-4567"}},
			expected: MatchStrong,
		},
		{
			name:     "email match",
			a:        &Contact{Emails: []string{"john@gmail.com"}},
			b:        &Contact{Emails: []string{"j.o.h.n@gmail.com"}},
			expected: MatchStrong,
		},
		{
			name:     "name only",
			a:        &Contact{FormattedName: "John Doe"},
			b:        &Contact{FormattedName: "john doe"},
			expected: MatchWeak,
		},
		{
			name:     "name with org",
			a:        &Contact{FormattedName: "John Doe", Organization: "Acme"},
			b:        &Contact{FormattedName: "John Doe", Organization: "Acme"},
			expected: MatchMedium,
		},
		{
			name:     "no match",
			a:        &Contact{FormattedName: "John Doe"},
			b:        &Contact{FormattedName: "Jane Smith"},
			expected: MatchNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CompareContacts(tt.a, tt.b)
			if got != tt.expected {
				t.Errorf("CompareContacts() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCompareContacts_DetailedStrength(t *testing.T) {
	tests := []struct {
		name     string
		a, b     *Contact
		expected MatchStrength
	}{
		// Strong matches - phone or email
		{
			name:     "phone match different names",
			a:        &Contact{FormattedName: "John", Phones: []string{"555-123-4567"}},
			b:        &Contact{FormattedName: "Johnny", Phones: []string{"+1-555-123-4567"}},
			expected: MatchStrong,
		},
		{
			name:     "email match different names",
			a:        &Contact{FormattedName: "John", Emails: []string{"john@example.com"}},
			b:        &Contact{FormattedName: "J Doe", Emails: []string{"JOHN@EXAMPLE.COM"}},
			expected: MatchStrong,
		},
		{
			name:     "both phone and email match",
			a:        &Contact{Phones: []string{"555-123-4567"}, Emails: []string{"john@example.com"}},
			b:        &Contact{Phones: []string{"555-123-4567"}, Emails: []string{"john@example.com"}},
			expected: MatchStrong,
		},

		// Medium matches - name + supporting evidence
		{
			name:     "name and organization match",
			a:        &Contact{FormattedName: "John Doe", Organization: "Acme Corp"},
			b:        &Contact{FormattedName: "john doe", Organization: "Acme Corp"},
			expected: MatchMedium,
		},
		{
			name:     "name and birthday match",
			a:        &Contact{FormattedName: "John Doe", Birthday: "1990-01-15"},
			b:        &Contact{FormattedName: "John Doe", Birthday: "1990-01-15"},
			expected: MatchMedium,
		},

		// Weak matches - name only
		{
			name:     "name only exact",
			a:        &Contact{FormattedName: "John Doe"},
			b:        &Contact{FormattedName: "John Doe"},
			expected: MatchWeak,
		},
		{
			name:     "name only with normalization",
			a:        &Contact{FormattedName: "Dr. John Doe PhD"},
			b:        &Contact{FormattedName: "john doe"},
			expected: MatchWeak,
		},
		{
			name:     "name only with accents",
			a:        &Contact{FormattedName: "José García"},
			b:        &Contact{FormattedName: "Jose Garcia"},
			expected: MatchWeak,
		},

		// No match
		{
			name:     "different names",
			a:        &Contact{FormattedName: "John Doe"},
			b:        &Contact{FormattedName: "Jane Smith"},
			expected: MatchNone,
		},
		{
			name:     "empty contacts",
			a:        &Contact{},
			b:        &Contact{},
			expected: MatchNone,
		},
		{
			name:     "same org different name",
			a:        &Contact{FormattedName: "John Doe", Organization: "Acme"},
			b:        &Contact{FormattedName: "Jane Smith", Organization: "Acme"},
			expected: MatchNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CompareContacts(tt.a, tt.b)
			if got != tt.expected {
				t.Errorf("CompareContacts() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// =============================================================================
// Real-World Scenarios
// =============================================================================

func TestDedupIndex_RealWorldScenarios(t *testing.T) {
	t.Run("exported contacts from different sources", func(t *testing.T) {
		// Simulates merging contacts from Google, iCloud, and Outlook
		existing := []*Contact{
			// From Google (has gmail-specific formatting)
			{FormattedName: "John Doe", Phones: []string{"+1-555-123-4567"}, Emails: []string{"john.doe@gmail.com"}},
			// From iCloud
			{FormattedName: "Jane Smith", Phones: []string{"(555) 234-5678"}, Emails: []string{"jane@icloud.com"}},
		}
		idx := NewDedupIndex(existing)

		// From Outlook (different formatting)
		outlookJohn := &Contact{
			FormattedName: "Doe, John", // Outlook often uses Last, First
			Phones:        []string{"555.123.4567"},
			Emails:        []string{"johndoe@gmail.com"}, // no dots
		}
		if !idx.IsDuplicate(outlookJohn) {
			t.Error("Should detect John from Outlook as duplicate (same phone)")
		}

		// LinkedIn export
		linkedInJane := &Contact{
			FormattedName: "Jane Smith",
			Phones:        []string{"5552345678"},
			Emails:        []string{"jane.smith@company.com"},
		}
		if !idx.IsDuplicate(linkedInJane) {
			t.Error("Should detect Jane from LinkedIn as duplicate (same phone)")
		}
	})

	t.Run("family members with same last name", func(t *testing.T) {
		existing := []*Contact{
			{FormattedName: "John Smith", Phones: []string{"555-111-1111"}, Emails: []string{"john@smith.com"}},
		}
		idx := NewDedupIndex(existing)

		// Wife - different person, should NOT be duplicate
		wife := &Contact{
			FormattedName: "Jane Smith",
			Phones:        []string{"555-222-2222"},
			Emails:        []string{"jane@smith.com"},
		}
		if idx.IsDuplicate(wife) {
			t.Error("Wife should not be flagged as duplicate of husband")
		}

		// Son - different person, should NOT be duplicate
		son := &Contact{
			FormattedName: "John Smith Jr.",
			Phones:        []string{"555-333-3333"},
			Emails:        []string{"junior@smith.com"},
		}
		if idx.IsDuplicate(son) {
			t.Error("Son should not be flagged as duplicate of father")
		}
	})

	t.Run("person changed phone number", func(t *testing.T) {
		existing := []*Contact{
			{FormattedName: "John Doe", Phones: []string{"555-OLD-NUMB"}, Emails: []string{"john@example.com"}},
		}
		idx := NewDedupIndex(existing)

		// Same person, new phone, same email
		newPhone := &Contact{
			FormattedName: "John Doe",
			Phones:        []string{"555-NEW-NUMB"},
			Emails:        []string{"john@example.com"},
		}
		if !idx.IsDuplicate(newPhone) {
			t.Error("Should detect as duplicate via email match")
		}
	})

	t.Run("corporate contacts", func(t *testing.T) {
		existing := []*Contact{
			{FormattedName: "Support", Phones: []string{"1-800-555-1234"}, Emails: []string{"support@company.com"}, Organization: "Acme Corp"},
		}
		idx := NewDedupIndex(existing)

		// Same support line, different label
		support2 := &Contact{
			FormattedName: "Acme Support",
			Phones:        []string{"800-555-1234"}, // no 1- prefix
			Organization:  "Acme Corporation",
		}
		if !idx.IsDuplicate(support2) {
			t.Error("Should detect as duplicate via phone match")
		}

		// Different department - should NOT match
		sales := &Contact{
			FormattedName: "Sales",
			Phones:        []string{"1-800-555-5678"},
			Emails:        []string{"sales@company.com"},
			Organization:  "Acme Corp",
		}
		if idx.IsDuplicate(sales) {
			t.Error("Different department should not be duplicate")
		}
	})
}

func TestDedupIndex_InternationalContacts(t *testing.T) {
	existing := []*Contact{
		// Spanish contact
		{FormattedName: "José García", Phones: []string{"+34 612 345 678"}, Emails: []string{"jose@example.es"}},
		// German contact
		{FormattedName: "Hans Müller", Phones: []string{"+49 30 12345678"}, Emails: []string{"hans@example.de"}},
		// Japanese contact (romanized)
		{FormattedName: "Tanaka Yuki", Phones: []string{"+81 3 1234 5678"}, Emails: []string{"tanaka@example.jp"}},
	}
	idx := NewDedupIndex(existing)

	tests := []struct {
		name        string
		contact     *Contact
		shouldMatch bool
	}{
		{
			name: "Spanish without country code",
			contact: &Contact{
				FormattedName: "Jose Garcia", // no accents
				Phones:        []string{"612 345 678"},
			},
			shouldMatch: true,
		},
		{
			name: "German with different umlaut representation",
			contact: &Contact{
				FormattedName: "Hans Mueller", // ue instead of ü
				Phones:        []string{"030-12345678"},
			},
			shouldMatch: true,
		},
		{
			name: "Japanese with reversed name order",
			contact: &Contact{
				FormattedName: "Yuki Tanaka", // Western order
				Phones:        []string{"03-1234-5678"},
			},
			shouldMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := idx.IsDuplicate(tt.contact)
			if got != tt.shouldMatch {
				t.Errorf("got duplicate=%v, want %v", got, tt.shouldMatch)
			}
		})
	}
}

// =============================================================================
// MergeContacts Tests
// =============================================================================

func TestMergeContacts_BasicFields(t *testing.T) {
	tests := []struct {
		name       string
		dst        *Contact
		src        *Contact
		wantMerged bool
		checkFunc  func(t *testing.T, dst *Contact)
	}{
		{
			name:       "merge missing formatted name",
			dst:        &Contact{},
			src:        &Contact{FormattedName: "John Doe"},
			wantMerged: true,
			checkFunc: func(t *testing.T, dst *Contact) {
				if dst.FormattedName != "John Doe" {
					t.Errorf("FormattedName = %q, want %q", dst.FormattedName, "John Doe")
				}
			},
		},
		{
			name:       "preserve existing formatted name",
			dst:        &Contact{FormattedName: "Existing Name"},
			src:        &Contact{FormattedName: "New Name"},
			wantMerged: false,
			checkFunc: func(t *testing.T, dst *Contact) {
				if dst.FormattedName != "Existing Name" {
					t.Errorf("FormattedName = %q, want %q", dst.FormattedName, "Existing Name")
				}
			},
		},
		{
			name:       "merge all name parts",
			dst:        &Contact{GivenName: "John"},
			src:        &Contact{GivenName: "Johnny", FamilyName: "Doe", MiddleName: "Q", Prefix: "Dr.", Suffix: "Jr."},
			wantMerged: true,
			checkFunc: func(t *testing.T, dst *Contact) {
				if dst.GivenName != "John" { // preserved
					t.Errorf("GivenName = %q, want %q", dst.GivenName, "John")
				}
				if dst.FamilyName != "Doe" { // merged
					t.Errorf("FamilyName = %q, want %q", dst.FamilyName, "Doe")
				}
				if dst.MiddleName != "Q" {
					t.Errorf("MiddleName = %q, want %q", dst.MiddleName, "Q")
				}
				if dst.Prefix != "Dr." {
					t.Errorf("Prefix = %q, want %q", dst.Prefix, "Dr.")
				}
				if dst.Suffix != "Jr." {
					t.Errorf("Suffix = %q, want %q", dst.Suffix, "Jr.")
				}
			},
		},
		{
			name:       "merge organization and title",
			dst:        &Contact{Organization: "Acme"},
			src:        &Contact{Organization: "Other Corp", Title: "Engineer"},
			wantMerged: true,
			checkFunc: func(t *testing.T, dst *Contact) {
				if dst.Organization != "Acme" { // preserved
					t.Errorf("Organization = %q, want %q", dst.Organization, "Acme")
				}
				if dst.Title != "Engineer" { // merged
					t.Errorf("Title = %q, want %q", dst.Title, "Engineer")
				}
			},
		},
		{
			name:       "merge birthday",
			dst:        &Contact{},
			src:        &Contact{Birthday: "1990-01-15"},
			wantMerged: true,
			checkFunc: func(t *testing.T, dst *Contact) {
				if dst.Birthday != "1990-01-15" {
					t.Errorf("Birthday = %q, want %q", dst.Birthday, "1990-01-15")
				}
			},
		},
		{
			name:       "preserve existing birthday",
			dst:        &Contact{Birthday: "1990-01-15"},
			src:        &Contact{Birthday: "1991-02-20"},
			wantMerged: false,
			checkFunc: func(t *testing.T, dst *Contact) {
				if dst.Birthday != "1990-01-15" {
					t.Errorf("Birthday = %q, want %q", dst.Birthday, "1990-01-15")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeContacts(tt.dst, tt.src)
			if got != tt.wantMerged {
				t.Errorf("MergeContacts() = %v, want %v", got, tt.wantMerged)
			}
			tt.checkFunc(t, tt.dst)
		})
	}
}

func TestMergeContacts_Emails(t *testing.T) {
	tests := []struct {
		name       string
		dstEmails  []string
		srcEmails  []string
		wantEmails []string
		wantMerged bool
	}{
		{
			name:       "add new emails",
			dstEmails:  []string{"existing@example.com"},
			srcEmails:  []string{"new@example.com"},
			wantEmails: []string{"existing@example.com", "new@example.com"},
			wantMerged: true,
		},
		{
			name:       "skip duplicate emails",
			dstEmails:  []string{"john@example.com"},
			srcEmails:  []string{"JOHN@EXAMPLE.COM"}, // same after normalization
			wantEmails: []string{"john@example.com"},
			wantMerged: false,
		},
		{
			name:       "skip gmail variations",
			dstEmails:  []string{"johndoe@gmail.com"},
			srcEmails:  []string{"john.doe@gmail.com", "johndoe+work@gmail.com"},
			wantEmails: []string{"johndoe@gmail.com"},
			wantMerged: false,
		},
		{
			name:       "merge multiple new emails",
			dstEmails:  []string{},
			srcEmails:  []string{"a@example.com", "b@example.com"},
			wantEmails: []string{"a@example.com", "b@example.com"},
			wantMerged: true,
		},
		{
			name:       "partial merge",
			dstEmails:  []string{"existing@example.com"},
			srcEmails:  []string{"existing@example.com", "new@example.com"},
			wantEmails: []string{"existing@example.com", "new@example.com"},
			wantMerged: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := &Contact{Emails: tt.dstEmails}
			src := &Contact{Emails: tt.srcEmails}

			got := MergeContacts(dst, src)
			if got != tt.wantMerged {
				t.Errorf("MergeContacts() = %v, want %v", got, tt.wantMerged)
			}
			if len(dst.Emails) != len(tt.wantEmails) {
				t.Errorf("len(Emails) = %d, want %d; Emails = %v", len(dst.Emails), len(tt.wantEmails), dst.Emails)
			}
			for i, want := range tt.wantEmails {
				if i < len(dst.Emails) && dst.Emails[i] != want {
					t.Errorf("Emails[%d] = %q, want %q", i, dst.Emails[i], want)
				}
			}
		})
	}
}

func TestMergeContacts_Phones(t *testing.T) {
	tests := []struct {
		name       string
		dstPhones  []string
		srcPhones  []string
		wantPhones []string
		wantMerged bool
	}{
		{
			name:       "add new phones",
			dstPhones:  []string{"555-111-1111"},
			srcPhones:  []string{"555-222-2222"},
			wantPhones: []string{"555-111-1111", "555-222-2222"},
			wantMerged: true,
		},
		{
			name:       "skip duplicate phones with different formats",
			dstPhones:  []string{"+1-555-123-4567"},
			srcPhones:  []string{"555-123-4567", "(555) 123-4567"},
			wantPhones: []string{"+1-555-123-4567"},
			wantMerged: false,
		},
		{
			name:       "skip duplicate with country code variation",
			dstPhones:  []string{"555-123-4567"},
			srcPhones:  []string{"+1-555-123-4567"},
			wantPhones: []string{"555-123-4567"},
			wantMerged: false,
		},
		{
			name:       "merge multiple new phones",
			dstPhones:  []string{},
			srcPhones:  []string{"555-111-1111", "555-222-2222"},
			wantPhones: []string{"555-111-1111", "555-222-2222"},
			wantMerged: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := &Contact{Phones: tt.dstPhones}
			src := &Contact{Phones: tt.srcPhones}

			got := MergeContacts(dst, src)
			if got != tt.wantMerged {
				t.Errorf("MergeContacts() = %v, want %v", got, tt.wantMerged)
			}
			if len(dst.Phones) != len(tt.wantPhones) {
				t.Errorf("len(Phones) = %d, want %d; Phones = %v", len(dst.Phones), len(tt.wantPhones), dst.Phones)
			}
		})
	}
}

func TestMergeContacts_Addresses(t *testing.T) {
	addr1 := Address{Street: "123 Main St", City: "Springfield", PostalCode: "12345"}
	addr2 := Address{Street: "456 Oak Ave", City: "Shelbyville", PostalCode: "67890"}
	addr1Dup := Address{Street: "123 main st", City: "SPRINGFIELD", PostalCode: "12345"} // same after normalization

	tests := []struct {
		name         string
		dstAddresses []Address
		srcAddresses []Address
		wantCount    int
		wantMerged   bool
	}{
		{
			name:         "add new address",
			dstAddresses: []Address{addr1},
			srcAddresses: []Address{addr2},
			wantCount:    2,
			wantMerged:   true,
		},
		{
			name:         "skip duplicate address",
			dstAddresses: []Address{addr1},
			srcAddresses: []Address{addr1Dup},
			wantCount:    1,
			wantMerged:   false,
		},
		{
			name:         "merge to empty",
			dstAddresses: []Address{},
			srcAddresses: []Address{addr1, addr2},
			wantCount:    2,
			wantMerged:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := &Contact{Addresses: tt.dstAddresses}
			src := &Contact{Addresses: tt.srcAddresses}

			got := MergeContacts(dst, src)
			if got != tt.wantMerged {
				t.Errorf("MergeContacts() = %v, want %v", got, tt.wantMerged)
			}
			if len(dst.Addresses) != tt.wantCount {
				t.Errorf("len(Addresses) = %d, want %d", len(dst.Addresses), tt.wantCount)
			}
		})
	}
}

func TestMergeContacts_URLs(t *testing.T) {
	tests := []struct {
		name       string
		dstURLs    []string
		srcURLs    []string
		wantURLs   []string
		wantMerged bool
	}{
		{
			name:       "add new URLs",
			dstURLs:    []string{"https://example.com"},
			srcURLs:    []string{"https://other.com"},
			wantURLs:   []string{"https://example.com", "https://other.com"},
			wantMerged: true,
		},
		{
			name:       "skip duplicate URLs (case insensitive)",
			dstURLs:    []string{"https://Example.com"},
			srcURLs:    []string{"https://example.com"},
			wantURLs:   []string{"https://Example.com"},
			wantMerged: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := &Contact{URLs: tt.dstURLs}
			src := &Contact{URLs: tt.srcURLs}

			got := MergeContacts(dst, src)
			if got != tt.wantMerged {
				t.Errorf("MergeContacts() = %v, want %v", got, tt.wantMerged)
			}
			if len(dst.URLs) != len(tt.wantURLs) {
				t.Errorf("len(URLs) = %d, want %d", len(dst.URLs), len(tt.wantURLs))
			}
		})
	}
}

func TestMergeContacts_Notes(t *testing.T) {
	tests := []struct {
		name       string
		dstNote    string
		srcNote    string
		wantNote   string
		wantMerged bool
	}{
		{
			name:       "merge to empty",
			dstNote:    "",
			srcNote:    "New note",
			wantNote:   "New note",
			wantMerged: true,
		},
		{
			name:       "append different note",
			dstNote:    "Existing note",
			srcNote:    "Additional info",
			wantNote:   "Existing note\n\n---\n\nAdditional info",
			wantMerged: true,
		},
		{
			name:       "skip identical note",
			dstNote:    "Same note",
			srcNote:    "Same note",
			wantNote:   "Same note",
			wantMerged: false,
		},
		{
			name:       "skip empty src note",
			dstNote:    "Existing",
			srcNote:    "",
			wantNote:   "Existing",
			wantMerged: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := &Contact{Note: tt.dstNote}
			src := &Contact{Note: tt.srcNote}

			got := MergeContacts(dst, src)
			if got != tt.wantMerged {
				t.Errorf("MergeContacts() = %v, want %v", got, tt.wantMerged)
			}
			if dst.Note != tt.wantNote {
				t.Errorf("Note = %q, want %q", dst.Note, tt.wantNote)
			}
		})
	}
}

func TestMergeContacts_RealWorldScenario(t *testing.T) {
	// Simulate merging contacts from different sources
	t.Run("Google meets iCloud contact", func(t *testing.T) {
		// Existing contact from Google
		existing := &Contact{
			FormattedName: "John Doe",
			Emails:        []string{"john.doe@gmail.com"},
			Phones:        []string{"+1-555-123-4567"},
			Organization:  "Acme Corp",
		}

		// Same person from iCloud export
		incoming := &Contact{
			FormattedName: "John Doe",
			GivenName:     "John",
			FamilyName:    "Doe",
			Emails:        []string{"johndoe@gmail.com", "john@work.com"}, // gmail variant + work email
			Phones:        []string{"555-123-4567", "555-999-8888"},       // same + home
			Birthday:      "1990-05-15",
			Note:          "Met at conference 2023",
		}

		merged := MergeContacts(existing, incoming)
		if !merged {
			t.Error("Expected merge to occur")
		}

		// Check merged fields
		if existing.GivenName != "John" {
			t.Errorf("GivenName not merged: %q", existing.GivenName)
		}
		if existing.FamilyName != "Doe" {
			t.Errorf("FamilyName not merged: %q", existing.FamilyName)
		}
		if existing.Birthday != "1990-05-15" {
			t.Errorf("Birthday not merged: %q", existing.Birthday)
		}

		// Should have original email + work email (gmail variant skipped)
		if len(existing.Emails) != 2 {
			t.Errorf("Expected 2 emails, got %d: %v", len(existing.Emails), existing.Emails)
		}

		// Should have original phone + home phone (duplicate skipped)
		if len(existing.Phones) != 2 {
			t.Errorf("Expected 2 phones, got %d: %v", len(existing.Phones), existing.Phones)
		}
	})

	t.Run("Sparse contact gets enriched", func(t *testing.T) {
		// Existing sparse contact (just a name and phone)
		existing := &Contact{
			FormattedName: "Jane",
			Phones:        []string{"555-111-1111"},
		}

		// Rich contact from another source
		incoming := &Contact{
			FormattedName: "Jane Smith",
			GivenName:     "Jane",
			FamilyName:    "Smith",
			Emails:        []string{"jane@example.com"},
			Phones:        []string{"555-111-1111"},
			Organization:  "Tech Inc",
			Title:         "Engineer",
			Birthday:      "1985-03-20",
			URLs:          []string{"https://jane.dev"},
			Note:          "Works on backend",
		}

		merged := MergeContacts(existing, incoming)
		if !merged {
			t.Error("Expected merge to occur")
		}

		// Original name preserved
		if existing.FormattedName != "Jane" {
			t.Errorf("FormattedName changed: %q", existing.FormattedName)
		}

		// New fields merged
		if existing.GivenName != "Jane" {
			t.Errorf("GivenName not merged")
		}
		if existing.FamilyName != "Smith" {
			t.Errorf("FamilyName not merged")
		}
		if len(existing.Emails) != 1 {
			t.Errorf("Email not merged")
		}
		if existing.Organization != "Tech Inc" {
			t.Errorf("Organization not merged")
		}
		if existing.Title != "Engineer" {
			t.Errorf("Title not merged")
		}
		if len(existing.URLs) != 1 {
			t.Errorf("URL not merged")
		}
	})
}
