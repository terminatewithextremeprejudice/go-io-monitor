package monitor

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	Black   = Color("\033[1;30m%s\033[0m")
	Red     = Color("\033[1;31m%s\033[0m")
	Green   = Color("\033[1;32m%s\033[0m")
	Yellow  = Color("\033[1;33m%s\033[0m")
	Purple  = Color("\033[1;34m%s\033[0m")
	Magenta = Color("\033[1;35m%s\033[0m")
	Teal    = Color("\033[1;36m%s\033[0m")
	White   = Color("\033[1;37m%s\033[0m")
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
	
	/* 
	// more of a code example that meaninful test
	t.Run("should provide real time updates", func (t *testing.T) {
		wg := sync.WaitGroup{}
		ch := make(chan bool)
		trafficMonitor := NewMonitor(0, 0)

		wg.Add(1)
		go func() {
			for {
				select {
					case <-ch:
						wg.Done()
						return
					default: {
						cmd := exec.Command("clear") //Linux example, its tested
						cmd.Stdout = os.Stdout
						cmd.Run()
						fmt.Println(Purple(fmt.Sprintf("current rate: ")), Green(fmt.Sprintf("%.2f kB/s", float64(trafficMonitor.Status().CurrentRate)/1024)), Yellow(" in"), Magenta(" / "), Red(fmt.Sprintf("%.2f kB/s", 0.0)), Yellow(" out"))
		
						time.Sleep(100 * time.Millisecond)
					}
				}
			}
		}()

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
		ch <- true
		close(ch)
		wg.Wait()
	})
	*/
}

func Color(colorString string) func(...interface{}) string {
	sprint := func(args ...interface{}) string {
		return fmt.Sprintf(colorString,
			fmt.Sprint(args...))
	}
	return sprint
}
