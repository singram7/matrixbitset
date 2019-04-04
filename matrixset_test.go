package matrixbitset

import (
	"fmt"
	"image/color"
	"image/png"
	"os"
	"sort"
	"testing"
)

func TestBounds(t *testing.T) {
	m := NewMatrixBitSet(6001, 6001)
	m.Fill(0, 0, 100, 100)
	m.Fill(100, 100, 3000, 3000)

	// Bounds tests nextSet and prevSet as well
	expectedMaxR, expectedMaxC := uint(3099), uint(3099)
	bounds, ok := m.BoundsOfSets()
	if !ok {
		t.Errorf("Got a bad response for Bounds")
	} else if bounds == nil {
		t.Error("Bounds returned nil bounds")
	} else if bounds.MinR != 0 && bounds.MinC != 0 {
		t.Errorf("Expected min [%d, %d], received [%d, %d]", 0, 0, bounds.MinR, bounds.MinC)
	} else if bounds.MaxR != expectedMaxR || bounds.MaxC != expectedMaxC {
		t.Errorf("Expected max [%d, %d], received [%d, %d]", expectedMaxR, expectedMaxC, bounds.MaxR, bounds.MaxC)
	}
}

func TestSingleBit(t *testing.T) {
	m := NewMatrixBitSet(6001, 6001)
	m.Set(3000, 3000)
	bounds, ok := m.BoundsOfSets()
	if !ok {
		t.Error("No bounds returned")
	} else {
		if bounds == nil {
			t.Error("Bounds returned nil bounds")
		}
		if bounds.MinR != 3000 {
			t.Errorf("Expected minR 3000, received %d", bounds.MinR)
		}
		if bounds.MaxR != 3000 {
			t.Errorf("Expected maxR 3000, received %d", bounds.MaxR)
		}
		if bounds.MinC != 3000 {
			t.Errorf("Expected minC 3000, received %d", bounds.MinC)
		}
		if bounds.MaxC != 3000 {
			t.Errorf("Expected maxC 3000, received %d", bounds.MaxC)
		}
	}
}

func TestInvert(t *testing.T) {
	m := NewMatrixBitSet(6001, 6001)
	m.Fill(0, 0, 100, 100)
	m.Fill(100, 100, 3000, 3000)

	m.Invert()
	first, good := m.NextSet(0)
	if !good {
		t.Error("NextSet(0) returned no results")
	} else if first != m.index(0, 100) {
		t.Errorf("Expected first index to be %d, received %d", m.index(0, 100), first)
	}
}

func TestShrink(t *testing.T) {
	m := NewMatrixBitSet(6001, 6001)
	m.Fill(10, 10, 50, 50)
	m.Fill(100, 100, 3000, 3000)

	shrink, transduce, err := m.ShrinkToBounds()
	if err != nil {
		t.Error("Shrink failed")
	} else {
		if shrink.R != 3100-10 {
			t.Errorf("Expected shrink R 3090, received %d", shrink.R)
		}
		if shrink.C != 3100-10 {
			t.Errorf("Expected shrink C 3090, received %d", shrink.C)
		}
		oldR, oldC := transduce(0, 0)
		if oldR != 10 {
			t.Errorf("EXpected transduce R to be 10, received %d", oldR)
		}
		if oldC != 10 {
			t.Errorf("EXpected transduce C to be 10, received %d", oldC)
		}
	}
}

func TestInternalN(t *testing.T) {
	m := NewMatrixBitSet(6001, 6001)
	m.Fill(100, 100, 500, 500)
	m.Fill(150, 50, 50, 50)  //kickout left from row 150 to 199 (inclusive)
	m.Fill(200, 500, 50, 50) //kickout right from row 200 to 299 (inclusive)
	if ok := m.internalN(m.index(300, 300)); !ok {
		t.Error("internalN(300, 300) failed")
	}
}

func TestNInside(t *testing.T) {
	m := NewMatrixBitSet(6001, 6001)
	m.Fill(100, 100, 500, 500)
	m.Fill(150, 50, 50, 50)  //kickout left from row 150 to 199 (inclusive)
	m.Fill(200, 500, 50, 50) //kickout right from row 200 to 299 (inclusive)
	if bounds, ok := m.BoundsOfSets(); ok {
		if !bounds.NInside(m.index(300, 300)) {
			t.Error("NInside failed for (300, 300)")
		}
	} else {
		t.Error("Bounds failed")
	}
}

func TestBorders(t *testing.T) {
	m := NewMatrixBitSet(6001, 6001)
	m.Fill(100, 100, 500, 500)
	m.Fill(150, 50, 50, 50)  //kickout left from row 150 to 199 (inclusive)
	m.Fill(200, 600, 50, 50) //kickout right from row 200 to 299 (inclusive)
	borders, ok := m.ExtractBorders()
	if !ok {
		t.Error("ExtractBorders failed")
	} else {
		if len(borders) != 2196 {
			t.Errorf("Expected 2196 borders, received %d", len(borders))
		}
	}
	blue := color.NRGBA{R: uint8(0), G: uint8(0), B: uint8(255), A: uint8(255)}
	red := color.NRGBA{R: uint8(255), G: uint8(0), B: uint8(0), A: uint8(255)}
	img := m.AsImage(blue)
	for _, mp := range borders {
		img.Set(mp.Col_i(), mp.Row_i(), red)
	}
	if f, err := os.Create("borders.png"); err == nil {
		if e := png.Encode(f, img); e != nil {
			t.Error(e)
		}
		if e := f.Close(); e != nil {
			t.Error(e)
		}
	} else {
		t.Error(err)
	}
}

func TestBoxes(t *testing.T) {
	m := NewMatrixBitSet(650, 650)
	m.Fill(0, 0, 64, 64)
	m.Fill(0, 64, 64, 64)
	m.Fill(0, 128, 64, 64)
	m.Fill(0, 512, 64, 64)
	m.Fill(0, 576, 64, 64)
	m.Fill(64, 0, 64, 64)
	m.Fill(64, 64, 64, 64)
	m.Fill(64, 128, 64, 64)
	m.Fill(64, 192, 64, 64)
	m.Fill(64, 512, 64, 64)
	m.Fill(64, 576, 64, 64)
	m.Fill(128, 0, 64, 64)
	m.Fill(128, 64, 64, 64)
	m.Fill(128, 128, 64, 64)
	m.Fill(128, 192, 64, 64)
	m.Fill(192, 320, 64, 64)
	m.Fill(192, 384, 64, 64)
	m.Fill(256, 576, 64, 64)
	m.Fill(320, 512, 64, 64)
	m.Fill(320, 576, 64, 64)
	m.Fill(384, 448, 64, 64)
	m.Fill(384, 512, 64, 64)
	m.Fill(384, 576, 64, 64)
	m.Fill(448, 384, 64, 64)
	m.Fill(448, 448, 64, 64)
	m.Fill(448, 512, 64, 64)
	m.Fill(512, 320, 64, 64)
	m.Fill(512, 384, 64, 64)
	m.Fill(512, 448, 64, 64)
	m.Fill(576, 320, 64, 64)
	m.Fill(576, 384, 64, 64)
	m.Fill(576, 448, 64, 64)
	if _, ok := m.JarvisHullOfSets(); ok {


	} else {
		t.Error("JarvisHull failed")
	}
}

func TestExtractPolygon(t *testing.T) {
	m := NewMatrixBitSet(6001, 6001)
	m.Fill(100, 100, 500, 500)
	m.Drain(150, 150, 25, 25) // add a hole
	m.Fill(150, 50, 50, 50)  //kickout left from row 150 to 199 (inclusive)
	m.Fill(200, 600, 50, 50) //kickout right from row 200 to 299 (inclusive)
	polygons, ok := m.ExtractAllPolygons()
	if !ok {
		t.Error("ExtractAllPolygons failed")
	} else {
		if len(polygons) != 1 {
			t.Errorf("Expected 1 polygon, received %d", len(polygons))
		} else {
			if len(polygons[0].Outer) != 13 {
				t.Errorf("Expected 13 vertexes, received %d", len(polygons[0].Outer))
				sort.Sort(ByRows(polygons[0].Outer))
				currentRow := -1
				prevPos := InvalidPos
				for _, mp := range polygons[0].Outer {
					if mp.Row_i() != currentRow {
						currentRow = mp.Row_i()
						prevPos = mp
					} else if mp.Row_i() == prevPos.Row_i()+1 {
						fmt.Println("Found row run from", prevPos.String(), "to", mp.String())
						prevPos = mp
					}
				}
				sort.Sort(ByCols(polygons[0].Outer))
				currentCol := -1
				prevPos = InvalidPos
				for _, mp := range polygons[0].Outer {
					if mp.Col_i() != currentCol {
						currentCol = mp.Col_i()
						prevPos = mp
					} else if mp.Col_i() == prevPos.Col_i()+1 {
						fmt.Println("Found col run from", prevPos.String(), "to", mp.String())
						prevPos = mp
					}
				}
				blue := color.NRGBA{R: uint8(0), G: uint8(0), B: uint8(255), A: uint8(255)}
				red := color.NRGBA{R: uint8(255), G: uint8(0), B: uint8(0), A: uint8(255)}
				img := m.AsImage(blue)
				for _, mp := range polygons[0].Outer {
					img.Set(mp.Col_i(), mp.Row_i(), red)
				}
				if f, err := os.Create("vertexes.png"); err == nil {
					if e := png.Encode(f, img); e != nil {
						t.Error(e)
					}
					if e := f.Close(); e != nil {
						t.Error(e)
					}
				} else {
					t.Error(err)
				}
			} else if polygons[0].Outer[0] != polygons[0].Outer[12] {
				t.Errorf("Expected first position %v to be equal to last %v", polygons[0].Outer[0], polygons[0].Outer[12])
			} else if len(polygons[0].Holes) != 1 {
				t.Errorf("Expected 1 hole, received %d", len(polygons[0].Holes))
			}
		}
	}
}
