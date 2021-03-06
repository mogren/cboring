package cboring

import (
	"bytes"
	"fmt"
	"io"
	"math"
)

// ReadRawBytes reads the next l bytes from r into a new byte slice.
func ReadRawBytes(l uint64, r io.Reader) (data []byte, err error) {
	if l > math.MaxInt32 {
		err = fmt.Errorf("cannot read %d raw bytes, is greater than max int32", l)
		return
	}

	// The following code is a compromise. Data up to one megabyte is cached in memory. However, larger data will be
	// copied in a temporary buffer, which is finally accessed. This is primarily a mitigation against resource
	// exhaustion attacks with constructed CBOR strings which indicate to contain a huge payload.
	if l <= 1024*1024 {
		data = make([]byte, l)
		_, err = io.ReadFull(r, data)
	} else {
		var buf bytes.Buffer
		if _, err = io.CopyN(&buf, r, int64(l)); err == nil {
			data = buf.Bytes()
		}
	}

	return
}

// ReadByteStringLen expects a byte string at the Reader's position and returns
// the byte string.
func ReadByteString(r io.Reader) (data []byte, err error) {
	n, err := ReadByteStringLen(r)
	if err != nil {
		return
	}

	return ReadRawBytes(n, r)
}

// WriteByteString writes a byte string into the Writer.
func WriteByteString(data []byte, w io.Writer) error {
	if err := WriteByteStringLen(uint64(len(data)), w); err != nil {
		return err
	}

	if n, err := w.Write(data); err != nil {
		return err
	} else if n != len(data) {
		return fmt.Errorf("WriteByteString: Wrote %d instead of %d bytes",
			n, len(data))
	}
	return nil
}

// ReadTextStringLen expects a text string at the Reader's position and returns
// the text string.
func ReadTextString(r io.Reader) (data string, err error) {
	n, err := ReadTextStringLen(r)
	if err != nil {
		return
	}

	if rdata, rerr := ReadRawBytes(n, r); rerr != nil {
		err = rerr
	} else {
		data = string(rdata)
	}
	return
}

// WriteTextString writes a byte string into the Writer.
func WriteTextString(data string, w io.Writer) error {
	if err := WriteTextStringLen(uint64(len(data)), w); err != nil {
		return err
	}

	// WriteString instead of w.Write to save a cast
	if n, err := io.WriteString(w, data); err != nil {
		return err
	} else if n != len(data) {
		return fmt.Errorf("WriteTextString: Wrote %d instead of %d bytes",
			n, len(data))
	}
	return nil
}
