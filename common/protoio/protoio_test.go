package protoio

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/cosmos/gogoproto/proto"
	"github.com/cosmos/gogoproto/test"
	"github.com/stretchr/testify/assert"
	"go.uber.org/multierr"
	"math/rand"
	"testing"
	"time"
)

func run(w Writer, r Reader) (errs error) {
	defer func() {
		_ = w.Close()
		_ = r.Close()
	}()
	size := 1000
	msgs := make([]*test.NinOptNative, size)
	lens := make([]int, size)
	randomness := rand.New(rand.NewSource(time.Now().Unix()))
	for i := range msgs {
		if i%3 == 0 {
			msgs[i] = test.NewPopulatedNinOptNative(randomness, true)
		} else if i%2 == 0 {
			msgs[i] = test.NewPopulatedNinOptNative(randomness, false)
		} else {
			msgs[i] = &test.NinOptNative{}
		}
		bz, err := proto.Marshal(msgs[i])
		errs = multierr.Append(err, err)
		length := len(bz)
		buf := make([]byte, binary.MaxVarintLen64)
		lens[i] = length + binary.PutUvarint(buf, uint64(length))
		n, err := w.WriteMsg(msgs[i])
		if err != nil {
			errs = multierr.Append(errs, err)
		}
		if n != lens[i] {
			errs = multierr.Append(errs, fmt.Errorf("writer: expected write %d bytes, actually write %d bytes", lens[i], n))
		}
	}

	for i := range msgs {
		msg := &test.NinOptNative{}
		n, err := r.ReadMsg(msg)
		if err != nil {
			errs = multierr.Append(errs, err)
			continue
		}
		if n != lens[i] {
			errs = multierr.Append(errs, fmt.Errorf("reader: expected read %d bytes, actually read %d bytes", lens[i], n))
			continue
		}
		if err = msg.VerboseEqual(msgs[i]); err != nil {
			errs = multierr.Append(errs, err)
			continue
		}
	}

	return errs
}

func TestMaxSizeNotExceeds(t *testing.T) {
	buf := &bytes.Buffer{}
	w := NewDelimitedWriter(buf)
	r := NewDelimitedReader(buf, 1024*1024)
	err := run(w, r)
	assert.Nil(t, err)
}

func TestMaxSizeExceeds(t *testing.T) {
	buf := &bytes.Buffer{}
	w := NewDelimitedWriter(buf)
	r := NewDelimitedReader(buf, 20)
	err := run(w, r)
	assert.Error(t, err)
	t.Log(err)
}

func TestVarintTruncated(t *testing.T) {
	buf := &bytes.Buffer{}
	buf.Write([]byte{0xff, 0xff})
	reader := NewDelimitedReader(buf, 1024*1024)
	msg := &test.NinOptNative{}
	n, err := reader.ReadMsg(msg)
	assert.Error(t, err)
	assert.Equal(t, 2, n)
}
