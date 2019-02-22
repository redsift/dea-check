package deacheck_test

import (
	"bufio"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/redsift/deacheck"
)

func TestIndex_Update(t *testing.T) {
	idx := deacheck.NewIndex("dea")

	if found := idx.HasDomain("dea", "example.com"); found {
		t.Errorf("HasDomain() return %t, want %t", found, !found)
	}

	err := idx.UpdateFromJSON("dea", strings.NewReader(`["example.com"]`))
	if err != nil {
		t.Errorf("UpdateFromJSON() err=%v, want=%v", err, nil)
	}

	if found := idx.HasDomain("dea", "example.com"); !found {
		t.Errorf("HasDomain() return %t, want %t", found, !found)
	}

	err = idx.UpdateFromJSON("dea", strings.NewReader(`["redsift.io"]`))
	if err != nil {
		t.Errorf("UpdateFromJSON() err=%v, want=%v", err, nil)
	}

	if found := idx.HasDomain("dea", "example.com"); found {
		t.Errorf("HasDomain() return %t, want %t", found, !found)
	}
}

const (
	deaFile      = "_samples/dea-20190222T1739.json"
	wildcardFile = "_samples/wildcard-20190222T1739.json"
)

func BenchmarkIndex_Update(b *testing.B) {

	dea, err := os.Open(deaFile)
	if err != nil {
		b.Fatalf(`couldn't open "%s": %v`, deaFile, err)
	}
	defer func() {
		_ = dea.Close()
	}()

	wildcard, err := os.Open(wildcardFile)
	if err != nil {
		b.Fatalf(`couldn't open "%s": %v`, wildcardFile, err)
	}
	defer func() {
		_ = wildcard.Close()
	}()

	b.ResetTimer()
	b.ReportAllocs()
	idx := deacheck.NewIndex("dea", "wildcard")
	for i := 0; i < b.N; i++ {
		if _, err := dea.Seek(0, io.SeekStart); err != nil {
			b.Fatalf(`couldn't reset "%s": %v`, deaFile, err)
		}
		if err := idx.UpdateFromJSON("dea", bufio.NewReader(dea)); err != nil {
			b.Fatalf(`couldn't update index from "%s": %v`, deaFile, err)
		}
		if _, err := wildcard.Seek(0, io.SeekStart); err != nil {
			b.Fatalf(`couldn't reset "%s": %v`, wildcardFile, err)
		}
		if err := idx.UpdateFromJSON("wildcard", bufio.NewReader(wildcard)); err != nil {
			b.Fatalf(`couldn't update index from "%s": %v`, wildcardFile, err)
		}
	}
}

func BenchmarkIndex_HasDomain(b *testing.B) {
	dea, err := os.Open(deaFile)
	if err != nil {
		b.Fatalf(`couldn't open "%s": %v`, deaFile, err)
	}
	defer func() {
		_ = dea.Close()
	}()

	idx := deacheck.NewIndex("dea")
	if err := idx.UpdateFromJSON("dea", bufio.NewReader(dea)); err != nil {
		b.Fatalf(`couldn't update index from "%s": %v`, deaFile, err)
	}

	domains := []string{"1rzpdv6y4a5cf5rcmxg.ml", "20mm.eu", "sharklasers.com"}
	domainsLen := len(domains)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		domain := domains[i%domainsLen]
		if found := idx.HasDomain("dea", domain); !found {
			b.Fatalf(`HasDomain("%s") return %t, want %t`, domain, found, !found)
		}
	}
}
