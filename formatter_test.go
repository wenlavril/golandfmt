package main

import (
	"strings"
	"testing"
)

func normalize(s string) string {
	return strings.TrimSpace(s) + "\n"
}

func TestVisualLen(t *testing.T) {
	tests := []struct {
		input    string
		tabWidth int
		expected int
	}{
		{"hello", 4, 5},
		{"\thello", 4, 9},        // tab(4) + "hello"(5)
		{"\t\thello", 4, 13},     // tab(4) + tab(4) + "hello"(5)
		{"\tprintln()", 4, 13},   // tab(4) + "println()"(9)
		{"no tabs here", 4, 12},
	}
	for _, tt := range tests {
		got := visualLen(tt.input, tt.tabWidth)
		if got != tt.expected {
			t.Errorf("visualLen(%q, %d) = %d, want %d", tt.input, tt.tabWidth, got, tt.expected)
		}
	}
}

func TestCallExpr_WrapIfLong(t *testing.T) {
	// 23 items (100-122). With indent "\t\t"=8 vis cols, 22 items fit (117 vis cols), 23rd breaks.
	input := `package main

func f() {
	println(100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115, 116, 117, 118, 119, 120, 121, 122)
}
`
	expected := `package main

func f() {
	println(
		100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115, 116, 117, 118, 119, 120, 121,
		122,
	)
}
`
	out, err := Format([]byte(input), 120)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}
	if normalize(string(out)) != normalize(expected) {
		t.Errorf("CallExpr wrap mismatch.\nGot:\n%s\nExpected:\n%s", string(out), expected)
	}
}

func TestCallExpr_NoWrapIfFits(t *testing.T) {
	input := `package main

func f() {
	println(100, 101, 102)
}
`
	out, err := Format([]byte(input), 120)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}
	if normalize(string(out)) != normalize(input) {
		t.Errorf("Short call should not be wrapped.\nGot:\n%s\nExpected:\n%s", string(out), input)
	}
}

func TestCompositeLit_WrapIfLong(t *testing.T) {
	// 23 items. Same packing logic: 22 fit on first wrapped line, 23rd breaks.
	input := `package main

func g() {
	_ = []int{100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115, 116, 117, 118, 119, 120, 121, 122}
}
`
	expected := `package main

func g() {
	_ = []int{
		100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115, 116, 117, 118, 119, 120, 121,
		122,
	}
}
`
	out, err := Format([]byte(input), 120)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}
	if normalize(string(out)) != normalize(expected) {
		t.Errorf("CompositeLit wrap mismatch.\nGot:\n%s\nExpected:\n%s", string(out), expected)
	}
}

func TestFuncParams_WrapIfLong(t *testing.T) {
	// "func h(" = 7 visual cols, no tab
	input := `package main

func h(p1 int, p2 int, p3 int, p4 int, p5 int, p6 int, p7 int, p8 int, p9 int, p10 int, p11 int, p12 int) {
}
`
	// 7 + items + ")" > 120: the items are ~7 chars each (with ", "), 12 items ≈ 7*12 = 84 + 7 + 1 = 92
	// Actually let me compute: "func h(p1 int, p2 int, p3 int, p4 int, p5 int, p6 int, p7 int, p8 int, p9 int, p10 int, p11 int, p12 int) {"
	// = 7 + 6*7 + 5*8 + 1 + 2 = ... let me just check if it exceeds 120.
	// The line is 107 chars. Need to check with visual len...
	// "func h(" = 7, items = ", "-separated field list, ") {" at end
	// Total: count manually... 
	// func h(p1 int, p2 int, p3 int, p4 int, p5 int, p6 int, p7 int, p8 int, p9 int, p10 int, p11 int, p12 int) {
	// Counting: that's about 109 characters. Let me use maxLen=80 to be safe.
	expected := `package main

func h(
	p1 int, p2 int, p3 int, p4 int, p5 int, p6 int, p7 int, p8 int, p9 int,
	p10 int, p11 int, p12 int,
) {
}
`
	out, err := Format([]byte(input), 80)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}
	if normalize(string(out)) != normalize(expected) {
		t.Errorf("FuncParams wrap mismatch.\nGot:\n%s\nExpected:\n%s", string(out), expected)
	}
}

func TestFuncResults_WrapIfLong(t *testing.T) {
	input := `package main

func j() (r1 int, r2 int, r3 int, r4 int, r5 int, r6 int, r7 int, r8 int, r9 int, r10 int, r11 int, r12 int) {
	return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0
}
`
	expected := `package main

func j() (
	r1 int, r2 int, r3 int, r4 int, r5 int, r6 int, r7 int, r8 int, r9 int,
	r10 int, r11 int, r12 int,
) {
	return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0
}
`
	out, err := Format([]byte(input), 80)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}
	if normalize(string(out)) != normalize(expected) {
		t.Errorf("FuncResults wrap mismatch.\nGot:\n%s\nExpected:\n%s", string(out), expected)
	}
}

func TestCollapse_MultiLineToSingleIfFits(t *testing.T) {
	input := `package main

func f() {
	println(
		1, 2, 3,
	)
}
`
	expected := `package main

func f() {
	println(1, 2, 3)
}
`
	out, err := Format([]byte(input), 120)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}
	if normalize(string(out)) != normalize(expected) {
		t.Errorf("Collapse mismatch.\nGot:\n%s\nExpected:\n%s", string(out), expected)
	}
}

func TestDemo_FullExample(t *testing.T) {
	input := `package main

func f() {
	println(100, 101, 102, 103, 104, 105)
}

func g() {
	_ = []int{100, 101, 102, 103, 104, 105}
}

func h(p1 int, p2 int, p3 int, p4 int, p5 int, p6 int) {
}

func j() (r1 int, r2 int, r3 int, r4 int, r5 int, r6 int) {
	return 0, 0, 0, 0, 0, 0
}
`
	// With maxLen=40, tab=4:
	// "\tprintln(" = 12 vis cols. Content "100, 101, 102, 103, 104, 105)" = 30 → 12+30=42 > 40 → wrap
	// Wrapped items at "\t\t" = 8 vis cols:
	//   "\t\t100, 101, 102, 103," = 8 + 20 = 28 ≤ 40 ✓, adding 104 → 28+2+3+1=34 ≤ 40 ✓ wait...
	// Let me recalculate: "100" = 3, ", 101" = 5 → cumulative after each item:
	//   100: 8+3 = 11
	//   101: 11+2+3 = 16 (with trailing comma check: 16+1=17 ≤ 40 ✓)
	//   102: 16+2+3 = 21 (21+1=22 ≤ 40 ✓)
	//   103: 21+2+3 = 26 (26+1=27 ≤ 40 ✓)
	//   104: 26+2+3 = 31 (31+1=32 ≤ 40 ✓)
	//   105: 31+2+3 = 36 (36+1=37 ≤ 40 ✓)
	// All 6 fit on one line! That means "\t\t100, 101, 102, 103, 104, 105," = 37 vis cols
	// That's only 1 line of items. But we want 2 lines like the demo...
	// The demo uses a smaller effective width. Let me use maxLen=35.
	// With maxLen=35:
	//   100: 8+3 = 11
	//   101: 11+5 = 16 (16+1=17 ≤ 35 ✓)
	//   102: 16+5 = 21 (21+1=22 ≤ 35 ✓)
	//   103: 21+5 = 26 (26+1=27 ≤ 35 ✓)
	//   104: 26+5 = 31 (31+1=32 ≤ 35 ✓)
	//   105: 31+5 = 36 (36+1=37 > 35) → new line
	// Line 1: "100, 101, 102, 103, 104," (32 vis)
	// Line 2: "105," 
	// But demo shows: "100, 101, 102, 103," and "104, 105,"
	// That's 4 items per line. Let me use maxLen=30:
	//   100: 8+3 = 11
	//   101: 11+5 = 16 (16+1=17 ≤ 30 ✓)
	//   102: 16+5 = 21 (21+1=22 ≤ 30 ✓)
	//   103: 21+5 = 26 (26+1=27 ≤ 30 ✓)
	//   104: 26+5 = 31 (31+1=32 > 30) → new line
	// Line 1: "\t\t100, 101, 102, 103,"
	// Line 2: "\t\t104, 105,"
	// That matches the demo! maxLen=30.
	//
	// For func params at maxLen=30:
	// "func h(" = 7 vis cols. Content = "p1 int, p2 int, p3 int, p4 int, p5 int, p6 int)" > 30 → wrap
	// Items at "\t" = 4 vis cols:
	//   "p1 int": 4+6 = 10
	//   "p2 int": 10+2+6 = 18 (18+1=19 ≤ 30 ✓)
	//   "p3 int": 18+2+6 = 26 (26+1=27 ≤ 30 ✓)
	//   "p4 int": 26+2+6 = 34 (34+1=35 > 30) → new line
	// Line 1: "\tp1 int, p2 int, p3 int,"
	// Line 2: "\tp4 int, p5 int, p6 int,"
	// Matches demo!
	//
	// For func results: same as params pattern.
	// "func j() (" = 10 vis cols > 30 counting content → wrap
	// Items at "\t" = 4 vis cols: same packing as params.

	expected := `package main

func f() {
	println(
		100, 101, 102, 103,
		104, 105,
	)
}

func g() {
	_ = []int{
		100, 101, 102, 103,
		104, 105,
	}
}

func h(
	p1 int, p2 int, p3 int,
	p4 int, p5 int, p6 int,
) {
}

func j() (
	r1 int, r2 int, r3 int,
	r4 int, r5 int, r6 int,
) {
	return 0, 0, 0, 0, 0, 0
}
`
	out, err := Format([]byte(input), 30)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}
	if normalize(string(out)) != normalize(expected) {
		t.Errorf("Demo mismatch.\nGot:\n%s\nExpected:\n%s", string(out), expected)
	}
}

func TestNoChange_ShortCode(t *testing.T) {
	input := `package main

func f() {
	x := 1
}
`
	out, err := Format([]byte(input), 120)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}
	if normalize(string(out)) != normalize(input) {
		t.Errorf("Short code should not change.\nGot:\n%s\nExpected:\n%s", string(out), input)
	}
}
