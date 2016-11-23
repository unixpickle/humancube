package main

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/unixpickle/humancube"
	"github.com/unixpickle/rubiksimg"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintln(os.Stderr, "Usage:", os.Args[0], "scramble solve output_dir")
		os.Exit(1)
	}
	scramble, err := humancube.CubeForMoves(os.Args[1])
	if err != nil {
		die("Scramble:", err)
	}
	outDir := os.Args[3]
	for i, x := range append(strings.Fields(os.Args[2]), "") {
		subPath := filepath.Join(outDir, fmt.Sprintf("cube%d.png", i))
		img := rubiksimg.GenerateImage(512, scramble.StickerCube())
		outFile, err := os.Create(subPath)
		if err != nil {
			die(err)
		}
		png.Encode(outFile, img)
		outFile.Close()

		humancube.Move(scramble, x)
	}
}

func die(msg ...interface{}) {
	fmt.Fprintln(os.Stderr, msg...)
	os.Exit(1)
}
