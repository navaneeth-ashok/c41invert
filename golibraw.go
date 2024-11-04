package main

// #cgo LDFLAGS: -lraw
// #include "libraw/libraw.h"
import "C"
import (
	"bytes"
	"fmt"
	"image"
	"os"
	"unsafe"

	"github.com/lmittmann/ppm"
)

// ImgMetadata contains image data read by libraw
type ImgMetadata struct {
	ScattoTimestamp int64
	ScattoDataOra   string
}

type rawImg struct {
	Height   int
	Width    int
	Bits     uint
	DataSize int
	Data     []byte
}

func (r rawImg) fullBytes() []byte {
	header := fmt.Sprintf("P6\n%d %d\n%d\n", r.Width, r.Height, (1<<r.Bits)-1)
	return append([]byte(header), r.Data...)
}

func lrInit() *C.libraw_data_t {
	librawProcessor := C.libraw_init(0)
	return librawProcessor
}

func goResult(result C.int) error {
	if int(result) == 0 {
		return nil
	}
	p := C.libraw_strerror(result)
	return fmt.Errorf("libraw error: %v", C.GoString(p))
}

// ImportRaw : Import and Decode camera RAW data using libraw
func ImportRaw(path string) (image.Image, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("input file [%v] does not exist", path)
	}

	librawProcessor := lrInit()
	defer C.libraw_recycle(librawProcessor)

	err := goResult(C.libraw_open_file(librawProcessor, C.CString(path)))
	if err != nil {
		return nil, fmt.Errorf("failed to open file [%v]", path)
	}

	err = goResult(C.libraw_unpack(librawProcessor))
	if err != nil {
		return nil, fmt.Errorf("failed to unpack file [%v]", path)
	}

	err = goResult(C.libraw_dcraw_process(librawProcessor))
	if err != nil {
		return nil, fmt.Errorf("failed to import file [%v]", path)
	}

	var result C.int

	img := C.libraw_dcraw_make_mem_image(librawProcessor, &result)
	defer C.libraw_dcraw_clear_mem(img)

	if goResult(result) != nil {
		return nil, fmt.Errorf("failed to import file [%v]", path)
	}
	dataBytes := make([]uint8, int(img.data_size))
	start := unsafe.Pointer(&img.data)
	size := unsafe.Sizeof(uint8(0))
	for i := 0; i < int(img.data_size); i++ {
		item := *(*uint8)(unsafe.Pointer(uintptr(start) + size*uintptr(i)))
		dataBytes[i] = item
	}

	rawImage := rawImg{
		Height:   int(img.height),
		Width:    int(img.width),
		DataSize: int(img.data_size),
		Bits:     uint(img.bits),
		Data:     dataBytes,
	}

	fullbytes := rawImage.fullBytes()
	return ppm.Decode(bytes.NewReader(fullbytes))
}
