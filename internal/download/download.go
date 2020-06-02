package download

import (
	"fmt"
	"net/http"
	"time"

	"github.com/cavaliercoder/grab"
	"github.com/cheggaaa/pb"
)

func Download(dst string, url string) error {
	client := grab.NewClient()
	req, err := grab.NewRequest(dst, url)
	if err != nil {
		return err
	}
	resp := client.Do(req)

	t := time.NewTicker(time.Millisecond)
	defer t.Stop()

	progress := &ProgressBar{}
	progress.Start(url, resp.Size)

L:
	for {
		select {
		case <-t.C:
			progress.SetCurrent(resp.BytesComplete())
		case <-resp.Done:
			progress.Finish()
			break L
		}
	}

	if err := resp.Err(); err != nil {
		if grab.IsStatusCodeError(err) {
			code := err.(grab.StatusCodeError)
			if int(code) == http.StatusNotFound {
				return fmt.Errorf("resource not found")
			}
		}
		return err
	}

	return nil
}

type ProgressBar struct {
	bar *pb.ProgressBar
}

// Start implement the DownloadProgress interface
func (p *ProgressBar) Start(url string, size int64) {
	p.bar = pb.Start64(size)
	p.bar.Set(pb.Bytes, true)
	p.bar.SetTemplateString(fmt.Sprintf(`download %s {{counters . }} {{percent . }} {{speed . }}`, url))
}

// SetCurrent implement the DownloadProgress interface
func (p *ProgressBar) SetCurrent(size int64) {
	p.bar.SetCurrent(size)
}

// Finish implement the DownloadProgress interface
func (p *ProgressBar) Finish() {
	p.bar.Finish()
}
