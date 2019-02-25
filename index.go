package deacheck

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path"
	"sync/atomic"
	"unsafe"

	"github.com/hashicorp/go-immutable-radix"
	"github.com/pkg/errors"
)

// Index holds lists of domains from different sources
type Index struct {
	forest map[string]*unsafe.Pointer // *iradix.Tree
}

// NewIndex creates *Index with declared sources.
// Each source is identified by key.
// Keys is a list of arbitrary strings
func NewIndex(keys ...string) *Index {
	forest := make(map[string]*unsafe.Pointer, len(keys))
	for _, src := range keys {
		p := unsafe.Pointer(iradix.New())
		forest[src] = &p
	}
	return &Index{
		forest: forest,
	}
}

// HasDomain checks if given domain is listed in the source
func (i *Index) HasDomain(source, domain string) bool {
	p, found := i.forest[source]
	if !found {
		return false
	}
	_, found = (*iradix.Tree)(*p).Get(key(domain))
	return found
}

// key produces []byte sequence of key from the string s
func key(s string) []byte {
	b := []byte(s)
	l := len(b)
	for i := 0; i < l/2; i++ {
		b[i], b[l-1-i] = b[l-1-i], b[i]
	}
	return b
}

// ReadAndUpdate opens given JSON file and updates the source from it.
func (i *Index) ReadAndUpdate(source, file string) error {
	r, err := os.Open(file)
	if err != nil {
		return errors.WithMessagef(err, `couldn't read data for "%s" from file "%s"`, source, file)
	}
	defer func() {
		_ = r.Close()
	}()
	return i.UpdateFromJSON(source, r)
}

// GetSaveAndUpdate makes a GET HTTP request to given url string, then uses response body to update the source.
// Additionally, it writes response body to the file.
func (i *Index) GetSaveAndUpdate(source, file, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return errors.WithMessagef(err, `couldn't download data for "%s" from "%s"`, source, url)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	_ = os.MkdirAll(path.Dir(file), 0755)

	w, err := os.OpenFile(file, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return errors.WithMessagef(err, `couldn't create file "%s"`, file)
	}

	updateErr := i.UpdateFromJSON(source, io.TeeReader(resp.Body, w))

	_ = w.Sync()
	if err := w.Close(); err != nil && updateErr == nil {
		return errors.WithMessagef(err, `couldn't close file "%s"`, file)
	}

	return updateErr
}

// GetAndUpdate makes a GET HTTP request to given url string, then uses response body to update the source
func (i *Index) GetAndUpdate(source, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return errors.WithMessagef(err, `couldn't download data for "%s" from "%s"`, source, url)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	return i.UpdateFromJSON(source, resp.Body)
}

// UpdateFromJSON updates given source from the io.Reader r.
// It assumes incoming JSON is an array of strings.
// The update happen atomically.
func (i *Index) UpdateFromJSON(source string, r io.Reader) error {
	if _, found := i.forest[source]; !found {
		return errors.Errorf(`undefined source "%s"`, source)
	}

	dec := json.NewDecoder(r)

	parsingError := func(err error) error {
		return errors.WithMessagef(err, `couldn't parse data for "%s"`, source)
	}

	if _, err := dec.Token(); err != nil {
		return parsingError(err)
	}

	tree := iradix.New()
	txn := tree.Txn()

	for dec.More() {
		var s string
		if err := dec.Decode(&s); err != nil {
			return parsingError(err)
		}
		txn.Insert(key(s), struct{}{})
	}

	if _, err := dec.Token(); err != nil {
		return parsingError(err)
	}

	tree = txn.Commit()

	p := i.forest[source]
	atomic.SwapPointer(p, unsafe.Pointer(tree))

	return nil
}
