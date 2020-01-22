package client

import (
	"io"
	"time"
)

/**
  Allows to track the percentage of read of a Reader by wrapping around it
  The percentage is sent through UpdateFunc whenever there's a change. Filesize has to be known

  Usage:

  reader := any_io.Reader
  &ProgressReader{
      Reader: af,
      Size: filesize,
	    Delay: 100 * time.Millisecond,
      UpdateFunc: func(percentage float32) {
        fmt.Println("Percentage read: %d\%", percentage)
      },
    }
*/
type ProgressReader struct {
	Reader     io.Reader
	Size       int64         // Reader Size in bytes
	UpdateFunc func(float32) // Function called on percentage change
	Delay      time.Duration // Delay in milliseconds before sending another update. 0 if not set
	totalRead  int64
	lastUpdate time.Time
}

func (pr *ProgressReader) Read(p []byte) (read int, err error) {
	read, err = pr.Reader.Read(p)
	pr.totalRead += int64(read)
	if pr.UpdateFunc != nil { // No need to do all this if no update function is set
		var percentage float32
		if read > 0 {
			percentage = float32(pr.totalRead) / float32(pr.Size) * 100
		} else {
			percentage = 100
		}

		if time.Since(pr.lastUpdate) > pr.Delay {
			pr.UpdateFunc(percentage)
			pr.lastUpdate = time.Now()
		}
	}
	return
}
