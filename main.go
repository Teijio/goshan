package main

import (
    "flag"
    "fmt"
)

func main() {
    destDir := flag.String("dest", "./output", "destination folder")
    width := flag.Int("w", 1024, "width of the image")
    isThumb := flag.Bool("thumb", false, "create thumb")

    flag.Parse()
    // получаем список файлов
    fmt.Println("Destination folder:", *destDir)
    fmt.Println("Width:", *width)
    fmt.Println("Thumbs:", *isThumb)
	for i, v := range flag.Args() {
		fmt.Println("Image file (%d) %s:\r\n", i, v)
	}
}
