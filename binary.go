package parse

import (
	"encoding/binary"
	"io"
	"math"
)

// NumberSize returns the number of bytes required to store a number.
//
// Accepted types are int8, uint8, int16, uint16, int32, uint32, int64, uint64,
// float32, and float64. A pointer to any such type is also accepted. Any other
// type returns 0.
func NumberSize(data interface{}) int {
	switch data.(type) {
	case int8, uint8, *int8, *uint8:
		return 1
	case int16, uint16, *int16, *uint16:
		return 2
	case int32, uint32, float32, *int32, *uint32, *float32:
		return 4
	case int64, uint64, float64, *int64, *uint64, *float64:
		return 8
	default:
		return 0
	}
}

// BinaryReader wraps an io.Reader to provide primitive methods for reading
// binary data.
//
// Methods on BinaryReader that read return a bool indicating failure. If an
// error occurs during any such call, then subsequent calls do nothing, and
// return true. The error that occurred can be retrieved with the Err method.
type BinaryReader struct {
	r   io.Reader
	ord binary.ByteOrder
	n   int64
	err error
}

// NewBinaryReader returns a BinaryReader that reads from r, with the byte order
// set to little endian.
func NewBinaryReader(r io.Reader) *BinaryReader {
	return &BinaryReader{r: r, ord: binary.LittleEndian}
}

// N returns the number of bytes read from the underlying reader.
func (r *BinaryReader) N() (n int64) {
	return r.n
}

// Err returns the first error that occurred while reading, if any.
func (r *BinaryReader) Err() (err error) {
	return r.err
}

// End returns the number of bytes read, and the first error that occurred.
func (r *BinaryReader) End() (n int64, err error) {
	return r.n, r.err
}

// SetByteOrder sets the byte order for which numeric values are read. Panics if
// order is nil.
func (r *BinaryReader) SetByteOrder(order binary.ByteOrder) {
	if order == nil {
		panic("expected non-nil ByteOrder")
	}
	r.ord = order
}

// Add receives the results of a read and adds them to the binary reader. err is
// treated as an error that occurs on the binary reader.
func (r *BinaryReader) Add(n int64, err error) (failed bool) {
	if r.err != nil {
		return true
	}

	r.n += n
	r.err = err

	if r.err != nil {
		return true
	}
	return false
}

// Bytes reads exactly len(p) bytes into p.
func (r *BinaryReader) Bytes(p []byte) (failed bool) {
	if r.err != nil {
		return true
	}

	var n int
	n, r.err = io.ReadFull(r.r, p)
	r.n += int64(n)

	if r.err != nil {
		return true
	}
	return false
}

// Number reads a number into v. v must be a pointer to any non-pointer type
// accepted by NumberSize. Any other type panics.
func (r *BinaryReader) Number(v interface{}) (failed bool) {
	if r.err != nil {
		return true
	}

	var b []byte
	if size := NumberSize(v); size == 0 {
		panic("invalid type")
	} else {
		var a [8]byte
		b = a[:size]
	}
	if r.Bytes(b) {
		return true
	}
	switch n := v.(type) {
	case *int8:
		*n = int8(b[0])
	case *uint8:
		*n = b[0]
	case *int16:
		*n = int16(r.ord.Uint16(b))
	case *uint16:
		*n = r.ord.Uint16(b)
	case *int32:
		*n = int32(r.ord.Uint32(b))
	case *uint32:
		*n = r.ord.Uint32(b)
	case *int64:
		*n = int64(r.ord.Uint64(b))
	case *uint64:
		*n = r.ord.Uint64(b)
	case *float32:
		*n = math.Float32frombits(r.ord.Uint32(b))
	case *float64:
		*n = math.Float64frombits(r.ord.Uint64(b))
	default:
		panic("invalid type")
	}
	return false
}

// All reads all remaining bytes.
func (r *BinaryReader) All() (data []byte, failed bool) {
	if r.err != nil {
		return nil, true
	}

	data, r.err = io.ReadAll(r.r)
	r.n += int64(len(data))

	if r.err != nil {
		return nil, true
	}

	return data, false
}

// BinaryWriter wraps an io.Writer to provide primitive methods for writing
// binary data.
//
// Methods on BinaryWriter that write return a bool indicating failure. If an
// error occurs during any such call, then subsequent calls do nothing, and
// return true. The error that occurred can be retrieved with the Err method.
type BinaryWriter struct {
	w   io.Writer
	ord binary.ByteOrder
	n   int64
	err error
}

// NewBinaryWriter returns a BinaryWriter that writes to w, with the byte order
// set to little endian.
func NewBinaryWriter(w io.Writer) *BinaryWriter {
	return &BinaryWriter{w: w, ord: binary.LittleEndian}
}

// N returns the number of bytes written to the underlying writer.
func (w *BinaryWriter) N() (n int64) {
	return w.n
}

// Err returns the first error that occurred while writing, if any.
func (w *BinaryWriter) Err() (err error) {
	return w.err
}

// End returns the number of bytes written, and the first error that occurred.
func (w *BinaryWriter) End() (n int64, err error) {
	return w.n, w.err
}

// SetByteOrder sets the byte order for which numeric values are written. Panics
// if order is nil.
func (w *BinaryWriter) SetByteOrder(order binary.ByteOrder) {
	if order == nil {
		panic("expected non-nil ByteOrder")
	}
	w.ord = order
}

// Add receives the results of a write and adds them to the binary writer. err
// is treated as an error that occurs on the binary writer.
func (w *BinaryWriter) Add(n int64, err error) (failed bool) {
	if w.err != nil {
		return true
	}

	w.n += n
	w.err = err

	if w.err != nil {
		return true
	}
	return false
}

// Bytes writes the bytes in p to the writer.
func (w *BinaryWriter) Bytes(p []byte) (failed bool) {
	if w.err != nil {
		return true
	}

	var n int
	n, w.err = w.w.Write(p)
	w.n += int64(n)
	if n < len(p) {
		return true
	}

	return false
}

// Number writes v as a number. v must be any non-pointer type accepted by
// NumberSize. Any other type panics.
func (w *BinaryWriter) Number(v interface{}) (failed bool) {
	if w.err != nil {
		return true
	}

	var b []byte
	if size := NumberSize(v); size == 0 {
		panic("invalid type")
	} else {
		var a [8]byte
		b = a[:size]
	}
	switch n := v.(type) {
	case int8:
		b[0] = uint8(n)
	case uint8:
		b[0] = n
	case int16:
		w.ord.PutUint16(b, uint16(n))
	case uint16:
		w.ord.PutUint16(b, n)
	case int32:
		w.ord.PutUint32(b, uint32(n))
	case uint32:
		w.ord.PutUint32(b, n)
	case int64:
		w.ord.PutUint64(b, uint64(n))
	case uint64:
		w.ord.PutUint64(b, n)
	case float32:
		w.ord.PutUint32(b, math.Float32bits(n))
	case float64:
		w.ord.PutUint64(b, math.Float64bits(n))
	default:
		panic("invalid type")
	}
	return w.Bytes(b)
}
