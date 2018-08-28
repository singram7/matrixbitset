package matrixbitset

import (
	"fmt"
)

type MatrixPos struct {
	r, c, stride uint
}

var InvalidPos MatrixPos

func (mp *MatrixPos) Valid() bool {
	return mp.stride > 0
}

func (mp *MatrixPos) Row() uint {
	return mp.r
}

func (mp *MatrixPos) Col() uint {
	return mp.c
}

func (mp *MatrixPos) Row_i() int {
	return int(mp.r)
}

func (mp *MatrixPos) Col_i() int {
	return int(mp.c)
}

func (mp *MatrixPos) N() uint {
	return mp.r*mp.stride + mp.c
}

func (mp *MatrixPos) N_i() int {
	return int(mp.N())
}

func (mp *MatrixPos) Both() (uint, uint) {
	return mp.r, mp.c
}

func (mp *MatrixPos) Both_i() (int, int) {
	return int(mp.r), int(mp.c)
}

func (mp MatrixPos) String() string {
	return fmt.Sprintf("[%d, %d]", mp.r, mp.c)
}

func (mp *MatrixPos) Up() MatrixPos {
	if mp.r > 0 {
		return NewMatrixPos(mp.r-1, mp.c, mp.stride)
	}
	return InvalidPos
}

func (mp *MatrixPos) Down(maxRows uint) MatrixPos {
	if mp.r < maxRows-1 {
		return NewMatrixPos(mp.r+1, mp.c, mp.stride)
	}
	return InvalidPos
}

func (mp *MatrixPos) Left() MatrixPos {
	if mp.c > 0 {
		return NewMatrixPos(mp.r, mp.c-1, mp.stride)
	}
	return InvalidPos
}

func (mp *MatrixPos) Right(maxCols uint) MatrixPos {
	if mp.c < maxCols-1 {
		return NewMatrixPos(mp.r, mp.c+1, mp.stride)
	}
	return InvalidPos
}

type ByRows []MatrixPos

func (bl ByRows) Len() int {
	return len(bl)
}

func (bl ByRows) Swap(x, y int) {
	bl[x], bl[y] = bl[y], bl[x]
}

func (bl ByRows) Less(x, y int) bool {
	if bl[x].r < bl[y].r {
		return true
	}
	if bl[x].r == bl[y].r {
		return bl[x].c < bl[y].c
	}
	return false
}

type ByCols []MatrixPos

func (bl ByCols) Len() int {
	return len(bl)
}

func (bl ByCols) Swap(x, y int) {
	bl[x], bl[y] = bl[y], bl[x]
}

func (bl ByCols) Less(x, y int) bool {
	if bl[x].c < bl[y].c {
		return true
	}
	if bl[x].c == bl[y].c {
		return bl[x].r < bl[y].r
	}
	return false
}
