package main

import (
	"testing"
)

func add(db *Database, name, secret string) *Entry {
	hash_, _ := hashFromName(DEFAULT_HASH)
	e := NewEntry(name, secret, DEFAULT_PERIOD, DEFAULT_DIGITS, hash_)
	db.Add(e)
	return e
}

func TestDatabaseAdd(t *testing.T) {
	db := &Database{}

	if len(db.Entries) != 0 {
		t.Fatalf("Empty database was not empty")
	}

	e1 := add(db, "entry 1", "secretone")
	e2 := add(db, "entry 2", "secrettwo")

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
	e1 := add(db, "entry 1", "secretone")
	e2 := add(db, "entry 2", "secrettwo")

	if s1 := db.FindExact("entry 1"); s1 != e1 {
		t.Errorf("Find Failed: Wanted %v got %v", e1, s1)
	}
	if s2 := db.FindExact("entry 2"); s2 != e2 {
		t.Errorf("Find Failed: Wanted %v got %v", e2, s2)
	}

	if sx := db.FindFuzzy("entry"); len(sx) != 2 {
		t.Errorf("Find Failed: Wanted %v or %v got %v", e1, e2, sx)
	}

	if s0 := db.FindFuzzy("not here"); len(s0) != 0 {
		t.Errorf("Find returned non-existing entry")
	}
}
