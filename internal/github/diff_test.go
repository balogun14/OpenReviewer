package github

import "testing"

func TestChangedLines(t *testing.T) {
	patch := `@@ -10,4 +10,5 @@ func main() {
 context
-removed
+added
 same
+another
}`

	lines := ChangedLines(patch)
	for _, line := range []int{11, 13} {
		if _, ok := lines[line]; !ok {
			t.Fatalf("expected changed line %d in %+v", line, lines)
		}
	}
	if _, ok := lines[12]; ok {
		t.Fatalf("did not expect unchanged line 12")
	}
}

func TestReviewableFileFiltersGeneratedAndUnsupportedFiles(t *testing.T) {
	cases := []struct {
		name string
		file PullRequestFile
		want bool
	}{
		{
			name: "go",
			file: PullRequestFile{Filename: "main.go", Status: "modified", Patch: "+package main"},
			want: true,
		},
		{
			name: "generated",
			file: PullRequestFile{Filename: "api.pb.go", Status: "modified", Patch: "+package api"},
			want: false,
		},
		{
			name: "lockfile",
			file: PullRequestFile{Filename: "package-lock.lock", Status: "modified", Patch: "+lock"},
			want: false,
		},
		{
			name: "removed",
			file: PullRequestFile{Filename: "main.go", Status: "removed", Patch: "-package main"},
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ReviewableFile(tc.file); got != tc.want {
				t.Fatalf("ReviewableFile() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestBuildDiffSkipsUnreviewableFiles(t *testing.T) {
	diff := BuildDiff([]PullRequestFile{
		{Filename: "main.go", Status: "modified", Patch: "@@ -1 +1 @@\n+package main"},
		{Filename: "api.pb.go", Status: "modified", Patch: "@@ -1 +1 @@\n+// Code generated"},
	})

	if diff == "" {
		t.Fatal("expected diff")
	}
	if contains(diff, "api.pb.go") {
		t.Fatalf("expected generated file to be skipped: %s", diff)
	}
}

func TestBuildDiffChunks(t *testing.T) {
	chunks := BuildDiffChunks([]PullRequestFile{
		{Filename: "a.go", Status: "modified", Patch: "@@ -1 +1 @@\n+package a"},
		{Filename: "b.go", Status: "modified", Patch: "@@ -1 +1 @@\n+package b"},
	}, 60)

	if len(chunks) != 2 {
		t.Fatalf("expected two chunks, got %d: %+v", len(chunks), chunks)
	}
	if !contains(chunks[0], "OpenReview diff chunk 1/2") {
		t.Fatalf("expected chunk marker: %s", chunks[0])
	}
}

func contains(value string, needle string) bool {
	for i := 0; i+len(needle) <= len(value); i++ {
		if value[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
