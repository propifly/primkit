package docgen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sampleDoc = `# Reference

## taskprim

### Commands

<!-- docgen:start:taskprim:commands -->
| Old | Table |
|-----|-------|
| old | data  |
<!-- docgen:end:taskprim:commands -->

### Schemas

Some hand-written content here.

## queueprim

<!-- docgen:start:queueprim:commands -->
| Old | Table |
|-----|-------|
<!-- docgen:end:queueprim:commands -->
`

func TestReplaceAnchored_basic(t *testing.T) {
	newContent := "| Command | Synopsis | Flags |\n|---------|----------|-------|"
	got, err := ReplaceAnchored(sampleDoc, "taskprim", newContent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got, "<!-- docgen:start:taskprim:commands -->") {
		t.Error("start anchor missing from output")
	}
	if !strings.Contains(got, "<!-- docgen:end:taskprim:commands -->") {
		t.Error("end anchor missing from output")
	}
	if strings.Contains(got, "| old | data  |") {
		t.Error("old content should have been replaced")
	}
	if !strings.Contains(got, "| Command | Synopsis | Flags |") {
		t.Error("new content should be present")
	}
	// Hand-written content must be preserved
	if !strings.Contains(got, "Some hand-written content here.") {
		t.Error("hand-written content outside anchors must be preserved")
	}
}

func TestReplaceAnchored_missingAnchor(t *testing.T) {
	_, err := ReplaceAnchored(sampleDoc, "nonexistent", "new content")
	if err == nil {
		t.Error("expected error for missing anchor, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %v", err)
	}
}

func TestReplaceAnchored_multipleAnchors(t *testing.T) {
	// Both anchors should be independently replaceable.
	step1, err := ReplaceAnchored(sampleDoc, "taskprim", "new-taskprim-table")
	if err != nil {
		t.Fatalf("step1: %v", err)
	}
	step2, err := ReplaceAnchored(step1, "queueprim", "new-queueprim-table")
	if err != nil {
		t.Fatalf("step2: %v", err)
	}
	if !strings.Contains(step2, "new-taskprim-table") {
		t.Error("taskprim replacement missing after step 2")
	}
	if !strings.Contains(step2, "new-queueprim-table") {
		t.Error("queueprim replacement missing after step 2")
	}
}

func TestReplaceAnchored_idempotent(t *testing.T) {
	newContent := "| Command | Synopsis | Flags |\n|---------|----------|-------|"
	first, err := ReplaceAnchored(sampleDoc, "taskprim", newContent)
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	second, err := ReplaceAnchored(first, "taskprim", newContent)
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	if first != second {
		t.Error("ReplaceAnchored should be idempotent with same content")
	}
}

func TestUpdateDoc_check_upToDate(t *testing.T) {
	dir := t.TempDir()
	docPath := filepath.Join(dir, "ref.md")

	// Build a doc that already has the correct generated content.
	meta := PrimMeta{
		Name: "testprim",
		Commands: []CmdMeta{
			{Name: "get", Synopsis: "get <id>", Flags: nil},
		},
	}
	table := RenderCommandTable(meta)
	doc := "<!-- docgen:start:testprim:commands -->\n" + table + "\n<!-- docgen:end:testprim:commands -->\n"

	if err := os.WriteFile(docPath, []byte(doc), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := UpdateDoc(docPath, []PrimMeta{meta}, true); err != nil {
		t.Errorf("expected no drift, got: %v", err)
	}
}

func TestUpdateDoc_check_drift(t *testing.T) {
	dir := t.TempDir()
	docPath := filepath.Join(dir, "ref.md")

	meta := PrimMeta{
		Name: "testprim",
		Commands: []CmdMeta{
			{Name: "get", Synopsis: "get <id>", Flags: nil},
		},
	}

	// Write a doc with stale content.
	doc := "<!-- docgen:start:testprim:commands -->\n| Stale | Table |\n|-------|-------|\n<!-- docgen:end:testprim:commands -->\n"
	if err := os.WriteFile(docPath, []byte(doc), 0o644); err != nil {
		t.Fatal(err)
	}

	err := UpdateDoc(docPath, []PrimMeta{meta}, true)
	if err == nil {
		t.Error("expected drift error, got nil")
	}
	if !strings.Contains(err.Error(), "out of date") {
		t.Errorf("expected 'out of date' in error, got: %v", err)
	}
	// File should be unchanged in check mode.
	contents, _ := os.ReadFile(docPath)
	if !strings.Contains(string(contents), "Stale") {
		t.Error("check mode should not modify the file")
	}
}

func TestUpdateDoc_write(t *testing.T) {
	dir := t.TempDir()
	docPath := filepath.Join(dir, "ref.md")

	meta := PrimMeta{
		Name: "testprim",
		Commands: []CmdMeta{
			{Name: "list", Synopsis: "list", Flags: []FlagMeta{
				{Name: "status", Usage: "filter by status", Default: ""},
			}},
		},
	}

	doc := "<!-- docgen:start:testprim:commands -->\n| Old |\n|-----|\n<!-- docgen:end:testprim:commands -->\n"
	if err := os.WriteFile(docPath, []byte(doc), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := UpdateDoc(docPath, []PrimMeta{meta}, false); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	contents, _ := os.ReadFile(docPath)
	if strings.Contains(string(contents), "| Old |") {
		t.Error("old content should have been replaced")
	}
	if !strings.Contains(string(contents), "`list`") {
		t.Error("new content should be present")
	}
}
