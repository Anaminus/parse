package parse

import (
	"bufio"
	"io"
	"io/ioutil"
)

// TextReader wraps an io.Reader to provide primitive methods for parsing text.
type TextReader struct {
	r   *bufio.Reader
	buf []byte
	n   int64
	err error
}

// NewTextReader returns a TextReader that reads r.
func NewTextReader(r io.Reader) *TextReader {
	br, ok := r.(*bufio.Reader)
	if !ok {
		br = bufio.NewReader(r)
	}
	return &TextReader{
		r:   br,
		buf: make([]byte, 64),
	}
}

// N returns the number of bytes read from the underlying reader.
func (r *TextReader) N() int64 {
	return r.n
}

// Err returns the first error that occurred while reading, if any.
func (r *TextReader) Err() error {
	return r.err
}

// End returns the number of bytes read, and the first error that occurred.
func (r *TextReader) End() (n int64, err error) {
	return r.n, r.err
}

// Next returns the next rune from the reader, and advances the cursor by the
// length of the rune. Returns r < 0 if an error occurred.
func (t *TextReader) Next() (r rune) {
	if t.err != nil {
		return -1
	}
	var w int
	r, w, t.err = t.r.ReadRune()
	if t.err != nil {
		return -1
	}
	t.n += int64(w)
	return r
}

// MustNext is like Next, but sets the error to io.ErrUnexpectedEOF if the
// end of the reader is reached.
func (t *TextReader) MustNext() (r rune) {
	if t.err != nil {
		return -1
	}
	if r = t.Next(); r < 0 {
		if t.err == io.EOF {
			t.err = io.ErrUnexpectedEOF
		}
	}
	return r
}

// Is compares s to the next characters in the reader. If they are equal, then
// the cursor is advanced, and true is returned. Otherwise, the cursor is not
// advanced, and false is returned.
func (t *TextReader) Is(s string) (ok bool) {
	if t.err != nil {
		return false
	}
	if s == "" {
		return true
	}
	var b []byte
	if b, t.err = t.r.Peek(len(s)); t.err != nil {
		if t.err == io.EOF {
			t.err = nil
		}
		return false
	}
	if string(b) != s {
		return false
	}
	t.r.Discard(len(s))
	t.n += int64(len(s))
	return true
}

// IsAny advances the cursor while the next character matches f, or until the
// end of the reader. Returns the characters read, and whether a non-EOF error
// occurred.
func (t *TextReader) IsAny(f func(rune) bool) (s string, ok bool) {
	if t.err != nil {
		return "", false
	}
	t.buf = t.buf[:0]
	for {
		var c rune
		var w int
		if c, w, t.err = t.r.ReadRune(); t.err != nil {
			if t.err == io.EOF {
				t.err = nil
				return string(t.buf), true
			}
			return "", false
		}
		if !f(c) {
			t.r.UnreadRune()
			return string(t.buf), true
		}
		t.buf = append(t.buf, string(c)...)
		t.n += int64(w)
	}
}

// IsEOF returns true if the cursor is at the end of the reader.
func (t *TextReader) IsEOF() (ok bool) {
	if t.err != nil {
		return t.err == io.EOF
	}
	_, err := t.r.Peek(1)
	return err == io.EOF
}

// Skip advances the cursor until a character does not match f. Returns whether
// a non-EOF error occurred.
func (t *TextReader) Skip(f func(rune) bool) (ok bool) {
	if t.err != nil {
		return false
	}
	for {
		var c rune
		var w int
		if c, w, t.err = t.r.ReadRune(); t.err != nil {
			if t.err == io.EOF {
				t.err = nil
				return true
			}
			return false
		}
		if !f(c) {
			t.r.UnreadRune()
			return true
		}
		t.n += int64(w)
	}
}

// Until advances the cursor until a character matches v. Returns the characters
// read, and whether an errored occurred.
func (t *TextReader) Until(v rune) (s string, ok bool) {
	if t.err != nil {
		return "", false
	}
	t.buf = t.buf[:0]
	for {
		var c rune
		var w int
		if c, w, t.err = t.r.ReadRune(); t.err != nil {
			if t.err == io.EOF {
				t.err = io.ErrUnexpectedEOF
			}
			return "", false
		}
		if c == v {
			return string(t.buf), true
		}
		t.buf = append(t.buf, string(c)...)
		t.n += int64(w)
	}
}

// UntilAny advances the cursor until a character matches f. Returns the
// characters read, and whether an errored occurred.
func (t *TextReader) UntilAny(f func(rune) bool) (s string, ok bool) {
	if t.err != nil {
		return "", false
	}
	t.buf = t.buf[:0]
	for {
		var c rune
		var w int
		if c, w, t.err = t.r.ReadRune(); t.err != nil {
			if t.err == io.EOF {
				t.err = io.ErrUnexpectedEOF
			}
			return "", false
		}
		if f(c) {
			return string(t.buf), true
		}
		t.buf = append(t.buf, string(c)...)
		t.n += int64(w)
	}
}

// UntilEOF reads the remaining characters in the reader.
func (t *TextReader) UntilEOF() (s string, ok bool) {
	if t.err != nil {
		return "", false
	}
	var b []byte
	if b, t.err = ioutil.ReadAll(t.r); t.err != nil {
		return "", false
	}
	t.n += int64(len(b))
	return string(b), true
}
