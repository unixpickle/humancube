package main

import (
	"errors"
	"strings"

	"github.com/unixpickle/gocube"
	"github.com/unixpickle/num-analysis/linalg"
)

// Move applies a WCA-notation move to the cube.
func Move(c *gocube.CubieCube, s string) error {
	hybrids := map[string]string{
		"r": "L x",
		"l": "R x'",
		"u": "D y",
		"d": "U y'",
		"f": "B z",
		"b": "F z'",
		"M": "R L' x'",
		"E": "U D' y'",
		"S": "F' B z",
	}

	if len(s) == 0 {
		return errors.New("empty move is invalid")
	}

	if hybrid, ok := hybrids[s[:1]]; ok {
		if len(s) == 1 {
			moves := strings.Fields(hybrid)
			for _, move := range moves {
				Move(c, move)
				return nil
			}
		} else if len(s) == 2 {
			repeats, ok := map[byte]int{'\'': 3, '2': 2}[s[1]]
			if !ok {
				return errors.New("invalid move: " + s)
			}
			for i := 0; i < repeats; i++ {
				Move(c, s[:1])
			}
			return nil
		} else {
			return errors.New("invalid move: " + s)
		}
	}

	rot, err := gocube.ParseRotation(s)
	if err == nil {
		stickers := c.StickerCube()
		stickers.Rotate(rot)
		stickers.ReinterpretCenters()
		cube, _ := stickers.CubieCube()
		*c = *cube
		return nil
	}

	move, err := gocube.ParseMove(s)
	if err != nil {
		return err
	}
	c.Move(move)
	return nil
}

// CubeVector returns a vectorized representation of
// the stickers of a cube.
func CubeVector(c *gocube.CubieCube) linalg.Vector {
	stickerCube := c.StickerCube()
	res := make(linalg.Vector, 8*6*6)

	var stickerIdx int
	for i, sticker := range stickerCube[:] {
		if i%9 == 4 {
			continue
		}
		res[sticker-1+stickerIdx] = 1
		stickerIdx += 6
	}

	return res
}
