package logring

import (
	dzstd "github.com/DataDog/zstd"
	"github.com/klauspost/compress/s2"
	"github.com/klauspost/compress/snappy"
	"github.com/klauspost/compress/zstd"
	"github.com/stretchr/testify/require"
	"io"
	"math/rand"
	"testing"
)

type rndGen struct {
	r *rand.Rand
}

func (r *rndGen) Read(b []byte) (n int, err error) {
	return r.r.Read(b)
}
func newRND() *rndGen {
	return &rndGen{r: rand.New(rand.NewSource(0))}

}

var compDisplayed = map[string]bool{}

func fmtPrintOnce(b *testing.B, format string, args ...interface{}) {
	if _, ok := compDisplayed[b.Name()]; !ok {
		b.Logf(format, args...)
		compDisplayed[b.Name()] = true
	}
}

func BenchmarkZSTDDatadog(b *testing.B) {
	r := logData()
	wr := NewCountingDiscardWriter()
	w := dzstd.NewWriter(wr)
	w.Write(r)
	w.Close()
	w = dzstd.NewWriter(wr)
	b.ResetTimer()
	b.ReportMetric(float64(len(r))/float64(wr.ctr), "compression")
	for i := 0; i < b.N; i++ {
		w.Write(r)
	}
	w.Close()
	b.ReportMetric((float64(b.N) * float64(len(r)) / b.Elapsed().Seconds() / 1024 / 1024), "MB/sec")
}
func BenchmarkZSTD(b *testing.B) {
	r := logData()
	wr := NewCountingDiscardWriter()
	w, err := zstd.NewWriter(wr)
	require.NoError(b, err)
	w.Write(r)
	w.Close()
	w.Reset(io.Discard)
	b.ResetTimer()
	b.ReportMetric(float64(len(r))/float64(wr.ctr), "compression")
	for i := 0; i < b.N; i++ {
		w.Write(r)
	}
	w.Close()
	b.ReportMetric((float64(b.N) * float64(len(r)) / b.Elapsed().Seconds() / 1024 / 1024), "MB/sec")
}
func BenchmarkSnappy(b *testing.B) {
	r := logData()
	wr := NewCountingDiscardWriter()
	w := snappy.NewBufferedWriter(wr)
	w.Write(r)
	w.Close()
	w.Reset(io.Discard)
	b.ResetTimer()
	b.ReportMetric(float64(len(r))/float64(wr.ctr), "compression")
	for i := 0; i < b.N; i++ {
		w.Write(r)
	}
	w.Close()
	b.ReportMetric((float64(b.N) * float64(len(r)) / b.Elapsed().Seconds() / 1024 / 1024), "MB/sec")
}

func BenchmarkS2(b *testing.B) {
	r := logData()
	wr := NewCountingDiscardWriter()
	w := s2.NewWriter(wr)
	b.Elapsed()
	w.Write(r)
	w.Close()
	w.Reset(io.Discard)
	b.ResetTimer()
	b.ReportMetric(float64(len(r))/float64(wr.ctr), "compression")
	for i := 0; i < b.N; i++ {
		w.Write(r)
	}
	w.Close()
	b.ReportMetric((float64(b.N) * float64(len(r)) / b.Elapsed().Seconds() / 1024 / 1024), "MB/sec")
}
