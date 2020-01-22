package client

import (
	"bytes"
	"io"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProgressReader_PercentageIsCorrect(t *testing.T) {
	byteArr := []byte{0x00, 0x01, 0x02, 0x03, 0x04}
	reader := bytes.NewReader(byteArr)
	var percentage float32
	progressReader := ProgressReader{
		Reader: reader,
		Size:   int64(len(byteArr)),
		UpdateFunc: func(p float32) {
			percentage = p
		},
	}
	buffer := make([]byte, 2)
	progressReader.Read(buffer)
	// Only read 40% until now
	percentageEquals(t, 40.0, percentage)
}

func TestProgressReader_DelayWorks(t *testing.T) {
	reader := getByteArrayReader(5)
	var percentage float32
	progressReader := ProgressReader{
		Reader: reader,
		Size:   5,
		Delay:  300 * time.Millisecond,
		UpdateFunc: func(p float32) {
			percentage = p
		},
	}
	buffer := make([]byte, 1)
	progressReader.Read(buffer)
	percentageEquals(t, 20.0, percentage)
	progressReader.Read(buffer) // Read again... update shouldnt be triggered so fast again
	percentageEquals(t, 20.0, percentage)
	// Now wait a second & read again. Percentage should be updated now
	time.Sleep(301 * time.Millisecond)
	progressReader.Read(buffer)
	percentageEquals(t, 60.0, percentage)
}

func TestProgressReader_WorksWithoutDelay(t *testing.T) {
	reader := getByteArrayReader(5)
	var counter int
	progressReader := ProgressReader{
		Reader: reader,
		Size:   5,
		UpdateFunc: func(p float32) {
			counter++
		},
	}
	buffer := make([]byte, 1)
	progressReader.Read(buffer)
	progressReader.Read(buffer)
	assert.Equal(t, 2, counter)
}

func getByteArrayReader(size int) io.Reader {
	byteArr := []byte{}
	for i := 0; i < size; i++ {
		byteArr = append(byteArr[:], 0x00)
	}
	return bytes.NewReader(byteArr)
}

func percentageEquals(t *testing.T, expected float32, actual float32) {
	if equal(actual, expected) == false {
		t.Errorf("Percentage should STILL be %f and not %f", expected, actual)
	}
}

// Compares equality of float32s since there could be a minimal error
func equal(x, y float32) bool {
	return math.Nextafter32(x, y) == y
}
