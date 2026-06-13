package main

import (
	"encoding/binary"
	"image"
	"image/color"
	"log"
	"os"
)

func main() {
	const W, H = 64, 64
	img := image.NewNRGBA(image.Rect(0, 0, W, H))
	bg := color.NRGBA{0x0d, 0x11, 0x17, 0xff}  // #0d1117
	chip := color.NRGBA{0x1c, 0x21, 0x28, 0xff} // #1c2128 (chip body)
	inner := color.NRGBA{0x3f, 0xb9, 0x50, 0xff} // #3fb950 (accent)
	pin := color.NRGBA{0x30, 0x36, 0x3d, 0xff}   // #30363d (pin)

	// fill background
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			img.SetNRGBA(x, y, bg)
		}
	}

	// draw chip body (rounded-like rectangle)
	bx0, by0 := 10, 14
	bx1, by1 := 54, 50
	for y := by0; y < by1; y++ {
		for x := bx0; x < bx1; x++ {
			img.SetNRGBA(x, y, chip)
		}
	}

	// draw inner accent rectangle
	ix0, iy0 := 18, 22
	ix1, iy1 := 46, 42
	for y := iy0; y < iy1; y++ {
		for x := ix0; x < ix1; x++ {
			img.SetNRGBA(x, y, inner)
		}
	}

	// draw simple stylized 'G' by carving inner rect
	// carve a gap to resemble a 'G' shape
	for y := 26; y < 38; y++ {
		for x := 30; x < 38; x++ {
			// make a notch on right-lower to suggest G
			if !(x >= 34 && y >= 32) {
				img.SetNRGBA(x, y, chip)
			}
		}
	}

	// draw pins on left and right edges
	for i := 0; i < 6; i++ {
		py := by0 + 2 + i*6
		// left pin
		for y := py; y < py+4; y++ {
			for x := 6; x < 10; x++ {
				img.SetNRGBA(x, y, pin)
			}
		}
		// right pin
		for y := py; y < py+4; y++ {
			for x := 54; x < 58; x++ {
				img.SetNRGBA(x, y, pin)
			}
		}
	}

	// prepare ICO
	f, err := os.Create("web/favicon.ico")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	// ICONDIR
	// Reserved 0, Type 1 (icon), Count 1
	binary.Write(f, binary.LittleEndian, uint16(0))
	binary.Write(f, binary.LittleEndian, uint16(1))
	binary.Write(f, binary.LittleEndian, uint16(1))
	// Placeholder for ICONDIRENTRY (16 bytes)
	entryOffset := 6 + 16
	// We'll build image data into buffer
	// BITMAPINFOHEADER (40 bytes)
	// width: W, height: H*2 (including AND mask)
	// planes=1, bitcount=32, compression=0 (BI_RGB)
	// image size: 40 + pixel data + mask
	pixelBytes := W * H * 4
	maskRowBytes := ((W + 31) / 32) * 4
	maskBytes := maskRowBytes * H
	imgSize := 40 + pixelBytes + maskBytes
	// write ICONDIRENTRY
	// width (1), height (1), colorCount (1), reserved (1), planes (2), bitCount (2), bytesInRes (4), imageOffset (4)
	widthByte := byte(W)
	heightByte := byte(H)
	binary.Write(f, binary.LittleEndian, widthByte)
	binary.Write(f, binary.LittleEndian, heightByte)
	binary.Write(f, binary.LittleEndian, byte(0)) // color count
	binary.Write(f, binary.LittleEndian, byte(0)) // reserved
	binary.Write(f, binary.LittleEndian, uint16(1)) // planes
	binary.Write(f, binary.LittleEndian, uint16(32)) // bitcount
	binary.Write(f, binary.LittleEndian, uint32(imgSize))
	binary.Write(f, binary.LittleEndian, uint32(entryOffset))
	// Now write BITMAPINFOHEADER
	// biSize
	binary.Write(f, binary.LittleEndian, uint32(40))
	binary.Write(f, binary.LittleEndian, int32(W))
	binary.Write(f, binary.LittleEndian, int32(H*2))
	binary.Write(f, binary.LittleEndian, uint16(1)) // planes (we repeat here as part of header?) actually biPlanes is uint16
	binary.Write(f, binary.LittleEndian, uint16(32))
	binary.Write(f, binary.LittleEndian, uint32(0)) // compression
	binary.Write(f, binary.LittleEndian, uint32(uint32(pixelBytes+maskBytes)))
	binary.Write(f, binary.LittleEndian, int32(0))
	binary.Write(f, binary.LittleEndian, int32(0))
	binary.Write(f, binary.LittleEndian, uint32(0))
	binary.Write(f, binary.LittleEndian, uint32(0))
	// Pixel data: BMP stores bottom-up; each pixel as B,G,R,A
	// iterate rows from bottom to top
	for y := H - 1; y >= 0; y-- {
		for x := 0; x < W; x++ {
			c := img.NRGBAAt(x, y)
			f.Write([]byte{c.B, c.G, c.R, c.A})
		}
	}
	// AND mask: 1 bit per pixel, padded to 32-bit rows. We'll write zeros (fully opaque)
	// Each row has maskRowBytes bytes
	mask := make([]byte, maskBytes)
	f.Write(mask)
	log.Printf("wrote %d byte favicon.ico", 6+16+imgSize)
}
