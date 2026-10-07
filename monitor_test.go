package monitor

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMonitor(t *testing.T) {
	t.Run("should provide statistics once download has completed", func (t *testing.T) {
 		trafficMonitor := NewMonitor(0, 0)

		resp, err := http.Get("https://dl-cdn.alpinelinux.org/alpine/v3.24/releases/x86_64/netboot/vmlinuz-lts")
		if err != nil {
			t.Fatalf("failed to download test document")	
		}
		inputPipe := NewReader(trafficMonitor, resp.Body)
		n, err := io.Copy(io.Discard, inputPipe)
		if err != nil {
			t.Fatalf("failed to download test document")	
		}
		inputPipe.Close()
		assert.Greater(t, trafficMonitor.Status().PeakRate, uint64(0))
		assert.Equal(t, trafficMonitor.Status().Bytes, uint64(n))
	})
}

