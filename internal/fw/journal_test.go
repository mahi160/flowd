package fw

import (
	"strings"
	"testing"
	"time"
)

// jan5 and jan6 are fixed test dates in January 2026.
var (
	jan5 = time.Date(2026, 1, 5, 14, 0, 0, 0, time.UTC)
	jan6 = time.Date(2026, 1, 6, 9, 0, 0, 0, time.UTC)
	jan7 = time.Date(2026, 1, 7, 16, 0, 0, 0, time.UTC)
)

func fakeBlock(start time.Time, repo string, focusMin int) Block {
	return Block{
		StartTS:    start,
		EndTS:      start.Add(time.Duration(focusMin) * time.Minute),
		FocusedMin: focusMin,
		Repo:       repo,
		Summary:    "## " + start.Format("15:04") + " – " + start.Add(time.Duration(focusMin)*time.Minute).Format("15:04") + "\n\nsome work\n",
	}
}

// ── rebuildMonthFile ─────────────────────────────────────────────────────────

func TestRebuildMonthFile_EmptyFile(t *testing.T) {
	blocks := []Block{fakeBlock(jan5, "flowd", 30)}
	got := rebuildMonthFile("", jan5, blocks, "")

	if !strings.HasPrefix(got, "# January 2026\n\n") {
		t.Errorf("expected month header; got:\n%s", got)
	}
	if !strings.Contains(got, "### Monday, 05 Jan") {
		t.Errorf("expected day heading; got:\n%s", got)
	}
}

func TestRebuildMonthFile_RewriteExistingDay(t *testing.T) {
	// Existing file has one day section for Jan 5.
	existing := "# January 2026\n\n### Monday, 05 Jan\n\n**Total:** 30 min · 1 blocks\n\nold content\n"
	blocks := []Block{
		fakeBlock(jan5, "flowd", 30),
		fakeBlock(jan5.Add(time.Hour), "flowd", 45),
	}
	got := rebuildMonthFile(existing, jan5, blocks, "")

	// Month header preserved exactly once.
	if count := strings.Count(got, "# January 2026"); count != 1 {
		t.Errorf("month header appears %d times; want 1", count)
	}
	// Day heading appears exactly once (not duplicated).
	if count := strings.Count(got, "### Monday, 05 Jan"); count != 1 {
		t.Errorf("day heading appears %d times; want 1", count)
	}
	// Old content replaced with new roll-up (2 blocks).
	if strings.Contains(got, "old content") {
		t.Error("old content should have been replaced")
	}
	if !strings.Contains(got, "2 blocks") {
		t.Errorf("expected updated roll-up with 2 blocks; got:\n%s", got)
	}
}

func TestRebuildMonthFile_InsertNewDayAtEnd(t *testing.T) {
	// Existing file has Jan 5; writing Jan 6 should append after it.
	existing := "# January 2026\n\n### Monday, 05 Jan\n\n**Total:** 30 min · 1 blocks\n\nsome work\n"
	blocks := []Block{fakeBlock(jan6, "flowd", 60)}
	got := rebuildMonthFile(existing, jan6, blocks, "")

	idxMon := strings.Index(got, "### Monday, 05 Jan")
	idxTue := strings.Index(got, "### Tuesday, 06 Jan")
	if idxMon < 0 {
		t.Error("Monday section missing")
	}
	if idxTue < 0 {
		t.Error("Tuesday section missing")
	}
	if idxMon >= idxTue {
		t.Error("Monday should appear before Tuesday")
	}
}

func TestRebuildMonthFile_InsertNewDayInMiddle(t *testing.T) {
	// Existing file has Jan 5 and Jan 7; writing Jan 6 should go between them.
	existing := "# January 2026\n\n" +
		"### Monday, 05 Jan\n\nday 5 content\n\n" +
		"### Wednesday, 07 Jan\n\nday 7 content\n"
	blocks := []Block{fakeBlock(jan6, "flowd", 45)}
	got := rebuildMonthFile(existing, jan6, blocks, "")

	idxMon := strings.Index(got, "### Monday, 05 Jan")
	idxTue := strings.Index(got, "### Tuesday, 06 Jan")
	idxWed := strings.Index(got, "### Wednesday, 07 Jan")

	if idxMon < 0 || idxTue < 0 || idxWed < 0 {
		t.Fatalf("missing section in:\n%s", got)
	}
	if !(idxMon < idxTue && idxTue < idxWed) {
		t.Errorf("wrong order: Mon=%d Tue=%d Wed=%d\n%s", idxMon, idxTue, idxWed, got)
	}
}

func TestRebuildMonthFile_StandupIncluded(t *testing.T) {
	blocks := []Block{fakeBlock(jan5, "flowd", 30)}
	standup := "• worked on flowd\n• fixed the journal"
	got := rebuildMonthFile("", jan5, blocks, standup)

	if !strings.Contains(got, "> • worked on flowd") {
		t.Errorf("standup not rendered as blockquote; got:\n%s", got)
	}
}

func TestRebuildMonthFile_SectionsSeparatedByBlankLine(t *testing.T) {
	// After a round-trip rewrite, consecutive day sections must be separated
	// by a blank line (no sections running together).
	existing := "# January 2026\n\n" +
		"### Monday, 05 Jan\n\nday 5 content\n\n" +
		"### Wednesday, 07 Jan\n\nday 7 content\n"
	blocks := []Block{fakeBlock(jan6, "flowd", 45)}
	got := rebuildMonthFile(existing, jan6, blocks, "")

	// Every "### " heading (after the first) must be preceded by a blank line.
	lines := strings.Split(got, "\n")
	for i := 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "### ") && lines[i-1] != "" {
			t.Errorf("line %d %q not preceded by blank line; prev=%q\nfull output:\n%s",
				i, lines[i], lines[i-1], got)
		}
	}
}

// ── splitIntoSections / dayFromHeading ───────────────────────────────────────

func TestSplitIntoSections(t *testing.T) {
	content := "# January 2026\n\n### Monday, 05 Jan\n\nblock a\n\n### Tuesday, 06 Jan\n\nblock b\n"
	sections := splitIntoSections(content)

	if len(sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(sections))
	}
	if !strings.Contains(sections["### Monday, 05 Jan"], "block a") {
		t.Error("Monday section missing content")
	}
	if !strings.Contains(sections["### Tuesday, 06 Jan"], "block b") {
		t.Error("Tuesday section missing content")
	}
}

func TestDayFromHeading_Order(t *testing.T) {
	earlier := dayFromHeading("### Monday, 05 Jan")
	later := dayFromHeading("### Tuesday, 06 Jan")
	if !earlier.Before(later) {
		t.Errorf("05 Jan should be before 06 Jan")
	}
}
