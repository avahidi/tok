package main

import (
	"testing"
)

func TestDatabaseAdd(t *testing.T) {
	db := &Database{}

	if len(db.Entries) != 0 {
		t.Fatalf("Empty database was not empty")
	}

	e1 := db.Add("entry 1", "secretone")
	e2 := db.Add("entry 2", "secrettwo")

	if len(db.Entries) != 2 {
		t.Fatalf("Database should have 2 entries, not %d", len(db.Entries))
	}
	if e1.Name != "entry 1" || e1.Secret != "secretone" {
		t.Errorf("Entry 1 incorrect: %v", e1)
	}
	if e2.Name != "entry 2" || e2.Secret != "secrettwo" {
		t.Errorf("Entry 2 incorrect: %v", e2)
	}
}

func TestDatabaseSearch(t *testing.T) {
	db := &Database{}
	e1 := db.Add("entry 1", "secretone")
	e2 := db.Add("entry 2", "secrettwo")

	if s1 := db.Find("entry 1"); s1 != e1 {
		t.Errorf("Find Failed: Wanted %v got %v", e1, s1)
	}
	if s2 := db.Find("entry 2"); s2 != e2 {
		t.Errorf("Find Failed: Wanted %v got %v", e2, s2)
	}

	if sx := db.Find("entry"); sx != e1 && sx != e2 {
		t.Errorf("Find Failed: Wanted %v or %v got %v", e1, e2, sx)
	}

	if s0 := db.Find("not here"); s0 != nil {
		t.Errorf("Find returned non-existing entry")
	}
}
