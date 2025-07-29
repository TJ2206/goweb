package main

import (
	"bytes"
	"fmt"
	"image"
	"log"
	"syscall/js"

	_ "image/jpeg"
	_ "image/png"

	webp "github.com/HugoSmits86/nativewebp"
)

func main() {
	fmt.Println("Go WASM Initialized")
	// Expose the function to JavaScript
	js.Global().Set("cropAndEncode", js.FuncOf(cropAndEncode))
	
	// NEW: Signal to JavaScript that Go is ready
	js.Global().Call("onGoWasmReady")

	<-make(chan bool) // Keep Go running
}

// cropAndEncode function remains the same...
func cropAndEncode(this js.Value, args []js.Value) interface{} {
	imgBytesJS := args[0]
	x := args[1].Int()
	y := args[2].Int()
	width := args[3].Int()
	height := args[4].Int()

	imgBytes := make([]byte, imgBytesJS.Get("length").Int())
	js.CopyBytesToGo(imgBytes, imgBytesJS)

	log.Printf("Go received crop command for box at (%d, %d)", x, y)

	img, _, err := image.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		log.Printf("Error decoding image: %v", err)
		return js.Null()
	}

	cropRect := image.Rect(x, y, x+width, y+height)
	croppedImg := img.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(cropRect)

	var webpBuffer bytes.Buffer
	if err := webp.Encode(&webpBuffer, croppedImg, nil); err != nil {
		log.Printf("Error encoding to WebP: %v", err)
		return js.Null()
	}

	webpBytes := webpBuffer.Bytes()
	jsResultArray := js.Global().Get("Uint8Array").New(len(webpBytes))
	js.CopyBytesToJS(jsResultArray, webpBytes)

	return jsResultArray
}