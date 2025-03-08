package main

import (
	"log"
	"testing"
)

func add(db *Database, name, secret string) *Entry {
	e, err := NewEntry(name, secret, DEFAULT_HASH, "", DEFAULT_PERIOD, DEFAULT_DIGITS)
	if err != nil {
		log.Fatalf("Internal error: '%v'", err)
	}
	db.Add(e)
	return e
}

func TestDatabaseAdd(t *testing.T) {
	db := &Database{}

	if len(db.Entries) != 0 {
		t.Fatalf("Empty database was not empty")
	}

	e1 := add(db, "entry 1", "NZSXMZLSEBTW63TOME")
	e2 := add(db, "entry 2", "M5UXMZJAPFXXKIDVOA")

	if len(db.Entries) != 2 {
		t.Fatalf("Database should have 2 entries, not %d", len(db.Entries))
	}
	if e1.Name != "entry 1" || e1.Secret != "NZSXMZLSEBTW63TOME" {
		t.Errorf("Entry 1 incorrect: %v", e1)
	}
	if e2.Name != "entry 2" || e2.Secret != "M5UXMZJAPFXXKIDVOA" {
		t.Errorf("Entry 2 incorrect: %v", e2)
	}
}

func TestDatabaseSearch(t *testing.T) {
	db := &Database{}
	e1 := add(db, "entry 1", "NZSXMZLSEBTW63TOME")
	e2 := add(db, "entry 2", "M5UXMZJAPFXXKIDVOA")

	// check the internal find functions first
	if s1, err := db.findIndex("#1"); err != nil || s1 != e1 {
		t.Errorf("findIndex failed: Wanted %v got %v", e1, s1)
	}

	if s2, err := db.findIndex("#2"); err != nil || s2 != e2 {
		t.Errorf("findIndex failed: Wanted %v got %v", e2, s2)
	}

	if s1 := db.findExact("entry 1"); s1 != e1 {
		t.Errorf("findExact Failed: Wanted %v got %v", e1, s1)
	}
	if s2 := db.findExact("entry 2"); s2 != e2 {
		t.Errorf("Find Failed: Wanted %v got %v", e2, s2)
	}

	if sx := db.findFuzzy("entry"); len(sx) != 2 {
		t.Errorf("Find Failed: Wanted %v or %v got %v", e1, e2, sx)
	}

	if s0 := db.findFuzzy("not here"); len(s0) != 0 {
		t.Errorf("Find returned non-existing entry")
	}

	// check the public Find() now
	if ss, err := db.Find("#1"); err != nil || len(ss) != 1 || ss[0] != e1 {
		t.Errorf("Find failed: Wanted %v got %v", e1, ss)
	}

	if ss, err := db.Find("entry 2"); err != nil || len(ss) != 1 || ss[0] != e2 {
		t.Errorf("Find failed: Wanted %v got %v", e2, ss)
	}

	if ss, err := db.Find("entry"); err != nil || len(ss) != 2 || ss[0] != e1 || ss[1] != e2 {
		t.Errorf("Find failed: got %v %v", err, ss)
	}

}
