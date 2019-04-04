package matrixbitset

import (
	"fmt"
	"image"
	"image/color"
	"math/bits"
	//	"sort"
	"strings"
	"sync"
)

// Shamelessly borrowed from https://github.com/willf/bitset
// the wordSize of a bit set
const wordSize = uint(64)

// Shamelessly borrowed from https://github.com/willf/bitset
// log2WordSize is lg2(wordSize)
const log2WordSize = uint(6)

// Shamelessly borrowed from https://github.com/willf/bitset
// allBits has every bit set
const allBits uint64 = 0xffffffffffffffff

type MatrixBitSet struct {
	B    []uint64
	R, C uint
}

func NewMatrixBitSet(w, h uint) *MatrixBitSet {
	// Round up to the next uint64 if partial ie. not a multiple of 64
	return &MatrixBitSet{B: make([]uint64, int((w*h)+(wordSize-1))>>log2WordSize), R: h, C: w}
}

func NewMatrixPos(r, c, stride uint) MatrixPos {
	return MatrixPos{r: r, c: c, stride: stride}
}

func (m *MatrixBitSet) Test(r, c uint) bool {
	m.panicPastMatrix(r, c)
	return m.test(m.index(r, c))
}

func (m *MatrixBitSet) TestN(i uint) bool {
	m.panicOverSized(i)
	return m.test(i)
}

func (m *MatrixBitSet) Set(r, c uint) *MatrixBitSet {
	m.panicPastMatrix(r, c)
	return m.set(m.index(r, c))
}

func (m *MatrixBitSet) SetN(i uint) *MatrixBitSet {
	m.panicOverSized(i)
	return m.set(i)
}

func (m *MatrixBitSet) Clear(r, c uint) *MatrixBitSet {
	m.panicPastMatrix(r, c)
	return m.clear(m.index(r, c))
}

func (m *MatrixBitSet) ClearN(i uint) *MatrixBitSet {
	m.panicOverSized(i)
	return m.clear(i)
}

func (m *MatrixBitSet) LastRow() uint {
	return m.R - 1
}

func (m *MatrixBitSet) LastCol() uint {
	return m.C - 1
}

func (m *MatrixBitSet) Fill(r, c, dr, dc uint) *MatrixBitSet {
	m.panicPastMatrix(r, c)
	m.panicPastMatrix(r+dr, c+dc)
	for row := r; row < r+dr; row++ {
		for col := c; col < c+dc; col++ {
			m.set(m.index(row, col))
		}
	}
	return m
}

func (m *MatrixBitSet) Drain(r, c, dr, dc uint) *MatrixBitSet {
	m.panicPastMatrix(r, c)
	m.panicPastMatrix(r+dr, c+dc)
	for row := r; row < r+dr; row++ {
		for col := c; col < c+dc; col++ {
			m.clear(m.index(row, col))
		}
	}
	return m
}

func (m *MatrixBitSet) NewPos(i uint) MatrixPos {
	stride := m.C
	return MatrixPos{r: i / stride, c: i % stride, stride: stride}
}

// Creates the minimal *M2 that contains
// all of the bits currently set in this bitset
// ie. Shrink to its BoundsOfSets
// returns a new M2
func (m *MatrixBitSet) ShrinkToBounds() (*MatrixBitSet, func(r, c uint) (uint, uint), error) {
	if bounds, ok := m.BoundsOfSets(); ok {
		w, h := bounds.Width()+1, bounds.Height()+1

		shrunk := NewMatrixBitSet(w, h)
		for i, e := m.nextSet(0); e; i, e = m.nextSet(i + 1) {
			r, c := i/m.C, i%m.C
			shrunk.SetN(shrunk.index(r-bounds.MinR, c-bounds.MinC))
		}
		// transducer to get original coords of matrix shrunken from
		transducer := func(r, c uint) (uint, uint) {
			return r + bounds.MinR, c + bounds.MinC
		}
		return shrunk, transducer, nil
	}
	return nil, nil, fmt.Errorf("No set bits are on")
}

// Is the passed index within the bounds (left, top, right, bottom)?
// Adapted from https://wrf.ecse.rpi.edu//Research/Short_Notes/pnpoly.html
func (mb *MatrixBounds) NInside(n uint) bool {
	mb.M.panicOverSized(n)
	nvert := len(mb.vertx)
	testy, testx := n/mb.M.C, n%mb.M.C

	inside := false
	for i, j := 0, nvert-1; i < nvert; j, i = i, i+1 {
		if (mb.verty[i] > testy) != (mb.verty[j] > testy) &&
			testx < (mb.vertx[j]-mb.vertx[i])*(testy-mb.verty[i])/(mb.verty[j]-mb.verty[i])+mb.vertx[i] {
			inside = !inside
		}
	}
	return inside
}

// Returns the MatrixBounds bits currently set on
func (m *MatrixBitSet) BoundsOfSets() (bounds *MatrixBounds, goodReturn bool) {
	minR, minC := m.R, m.C
	goodReturn = false
	if first, good := m.nextSet(0); good {
		minR, minC = first/m.C, first%m.C
		maxR, maxC := minR, minC
		bounds = &MatrixBounds{
			M:      m,
			MinR:   minR,
			MinC:   minC,
			MaxR:   maxR,
			MaxC:   maxC,
			left:   first,
			right:  first,
			top:    first,
			bottom: first,
		}
		goodReturn = true
		lastIndex := m.index(m.R-1, m.C-1)
		if m.test(lastIndex) {
			// last bit is on, just test for minC
			bounds.MaxC = m.C - 1
			bounds.MaxR = m.R - 1
			bounds.right = lastIndex
			bounds.bottom = lastIndex
		} else {
			// we will, at a min, get the first point, ignore the bool return
			last, _ := m.prevSet(lastIndex)
			bounds.MaxR, bounds.MaxC = last/m.C, last%m.C
			bounds.right, bounds.bottom = last, last
			if bounds.Height() != 0 && bounds.MaxC < m.C-1 {
				bounds.MaxC, bounds.right, _ = m.FindMaxC(bounds)
			}
		}

		if bounds.Height() != 0 && bounds.MinC != 0 {
			bounds.MinC, bounds.left, _ = m.FindMinC(bounds)
		}
		bounds.setup()
	}
	return
}

// Finds the minimum column between the passed rows
func (m *MatrixBitSet) FindMinC(bounds *MatrixBounds) (minC, n uint, good bool) {
	minC = m.C

	for row := bounds.MinR + 1; row < bounds.MaxR; row++ {
		if mc, goodR := m.nextSet(m.index(row, 0)); goodR {
			c := mc % m.C
			if c < minC {
				minC = c
				n = mc
				// c low as possible?
				if c == 0 {
					break
				}
			}
		} else {
			break
		}
	}
	good = minC != m.C
	return
}

// Finds the largest col within the passed rows up to the largest row
// Does not include the maxR
func (m *MatrixBitSet) FindMaxC(bounds *MatrixBounds) (maxC, n uint, good bool) {
	maxC = uint(0)
	n = 0
	for r := int(bounds.MaxR) - 1; r >= int(bounds.MinR); r-- {
		// prev from the first column of the next row
		// so we catch the last bit of our row being on
		firstColNextRow := m.index(uint(r)+1, 0)
		if mxc, goodC := m.prevSet(firstColNextRow); goodC {
			mc := mxc % m.C
			if mc > maxC {
				maxC = mc
				n = mxc
				// Already as big as possible?
				if mc == m.LastCol() {
					break
				}
			}
		} else {
			break
		}
	}
	good = maxC != uint(0)
	return
}

// Extract just non interior positions
func (m *MatrixBitSet) ExtractBorders() ([]MatrixPos, bool) {
	borders := make([]MatrixPos, 0, 512)
	for i, e := m.nextSet(0); e; i, e = m.NextSet(i + 1) {
		if !m.internalN(i) {
			borders = append(borders, m.NewPos(i))
		}
	}
	return borders, true
}

type PosSet struct {
	mutex sync.RWMutex
	pos   map[string]MatrixPos
}

func (ps *PosSet) Add(mp MatrixPos) bool {
	s := mp.String()
	ps.mutex.Lock()
	if _, found := ps.pos[s]; !found {
		ps.pos[s] = mp
		ps.mutex.Unlock()
		return true
	}
	ps.mutex.Unlock()
	return false
}

func (ps *PosSet) Contains(mp MatrixPos) bool {
	s := mp.String()
	ps.mutex.RLock()
	_, found := ps.pos[s]
	ps.mutex.RUnlock()
	return found
}

// Border points that are themselves vertexes,
// (ie. not a part of horizontal or vertical spans)
//func (m *MatrixBitSet) ExtractVertexes(expected *PosSet) ([]MatrixPos, bool) {
//	// First whittle down to just border pos
//	if borders, ok := m.ExtractBorders(); ok {
//		firstPos := InvalidPos
//		prevPos := InvalidPos
//		lastPos := InvalidPos

//		posSet := &PosSet{pos: make(map[string]MatrixPos)}

//		output := func() {
//			if lastPos.Valid() {
//				posSet.Add(firstPos)
//				posSet.Add(lastPos)
//				if !expected.Contains(firstPos) {
//					fmt.Println("Connecting", firstPos, lastPos)
//				}
//				lastPos = InvalidPos
//				firstPos = InvalidPos
//			} else if firstPos.Valid() {
//				//				posSet.Add(firstPos)
//				firstPos = InvalidPos
//			}
//		}

//		// Now sort them by rows to find horizontal runs and eliminate
//		sort.Sort(ByRows(borders))
//		currentRow := -1
//		for _, mp := range borders {
//			if mp.Row_i() != currentRow {
//				output()
//				currentRow = mp.Row_i()
//				firstPos = mp
//				prevPos = mp
//			} else if mp.Col_i() == prevPos.Col_i()+1 {
//				// next col over, just keep span going
//				lastPos = mp
//				prevPos = mp
//			} else {
//				// row matches, but not next col
//				// output and reset
//				output()
//				currentRow = -1
//			}
//		}
//		output()

//		fmt.Println("Sorting by cols")

//		// Now sort them by cols to find vertical runs and eliminate
//		sort.Sort(ByCols(borders))
//		currentCol := -1
//		for _, mp := range borders {
//			if mp.Col_i() != currentCol {
//				output()
//				currentCol = mp.Col_i()
//				firstPos = mp
//				prevPos = mp
//			} else if mp.Row_i() == prevPos.Row_i()+1 {
//				// next row down, just keep span going
//				lastPos = mp
//				prevPos = mp
//			} else {
//				// col matches, but not next row
//				// output and reset
//				output()
//				currentCol = -1
//			}
//		}
//		output()

//		retval := make([]MatrixPos, 0, len(posSet.pos))
//		for _, v := range posSet.pos {
//			retval = append(retval, v)
//		}
//		return retval, true
//	}
//	return []MatrixPos{}, false
//}

func (m *MatrixBitSet) AsImage(clr color.NRGBA) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, int(m.R), int(m.C)))

	for i, e := m.nextSet(0); e; i, e = m.nextSet(i + 1) {
		pos := m.NewPos(i)
		img.Set(pos.Col_i(), pos.Row_i(), clr)
	}
	return img
}

func (m *MatrixBitSet) JarvisHullOfSets() ([]MatrixPos, bool) {
	points := make([]MatrixPos, 0, 512)
	if bounds, ok := m.BoundsOfSets(); ok {
		total := 0
		maxTotal := 600 * 10
		foundLeft, foundRight := false, false
		foundTop, foundBottom := false, false
		if borders, good := m.ExtractBorders(); good {
			for _, mp := range borders {
				total++
				if total >= maxTotal {
					break
				}
				i := mp.N()
				if !bounds.NInside(i) {
					points = append(points, mp)
					switch i {
					case bounds.left:
						foundLeft = true
					case bounds.top:
						foundTop = true
					case bounds.right:
						foundRight = true
					case bounds.bottom:
						foundBottom = true
					}
				}
			}
		}
		if !foundLeft {
			points = append(points, m.NewPos(bounds.left))
		}
		if !foundTop {
			points = append(points, m.NewPos(bounds.top))
		}
		if !foundRight {
			points = append(points, m.NewPos(bounds.right))
		}
		if !foundBottom {
			points = append(points, m.NewPos(bounds.bottom))
		}
		return m.jarvisHull(points, bounds)
	}
	return []MatrixPos{}, false
}

func (m *MatrixBitSet) internalN(n uint) bool {
	pos := m.NewPos(n)
	up := pos.Up()
	upLeft := up.Left()
	upRight := up.Right(m.C)
	down := pos.Down(m.R)
	downLeft := down.Left()
	downRight := down.Right(m.C)
	left := pos.Left()
	right := pos.Right(m.C)
	return m.testPos(up)+m.testPos(upLeft)+m.testPos(upRight)+
		m.testPos(down)+m.testPos(downLeft)+m.testPos(downRight)+
		m.testPos(left)+m.testPos(right) == 8
}

func (m *MatrixBitSet) jarvisHull(points []MatrixPos, bounds *MatrixBounds) ([]MatrixPos, bool) {
	n := len(points)
	p := 0
	lr, lc := points[p].Both()
	for i := 1; i < n; i++ {
		r, c := points[i].Both()
		if c < lc {
			p = i
			lr, lc = r, c
		} else if c == lc {
			if r > lr {
				p = i
				lr, lc = r, c
			}
		}
	}

	var q int
	l := p

	hull := make([]MatrixPos, 0, 512)
	maxSteps := 4096
	for step := 0; step < maxSteps; step++ {
		hull = append(hull, points[p])
		q = (p + 1) % n
		// Find the most CounterClockwise point from p
		for i := 0; i < n; i++ {
			if m.Orient(points[p], points[i], points[q]) > 0 {
				// now q points to a candidate, continue looking for better
				q = i
			}
		}
		// q now is the most CounterClockwise point to p in the list
		p = q
		if p == l {
			// All done
			// Close the polygon
			hull = append(hull, points[p])

			// Reverse it to make it clockwise
			// Borrowed from SliceTricks
			// https://github.com/golang/go/wiki/SliceTricks
			for left, right := 0, len(hull)-1; left < right; left, right = left+1, right-1 {
				hull[left], hull[right] = hull[right], hull[left]
			}
			break
		}
	}
	return hull, true
}

// Orientation of the tuple
// 0 = Colinear, >0 CounterClockwise, <0 Clockwise
// Matrix has origin top left
// this formula <https://www.geeksforgeeks.org/orientation-3-ordered-points/>
// is cartesian (origin lower left)
// this flips the slopes, so > 0 is counter clockwise
func (m *MatrixBitSet) Orient(p, q, r MatrixPos) int {
	qpRow := q.Row_i() - p.Row_i()
	rqCol := r.Col_i() - q.Col_i()
	qpCol := q.Col_i() - p.Col_i()
	rqRow := r.Row_i() - q.Row_i()
	retval := (qpRow * rqCol) - (qpCol * rqRow)
	return retval
}

func (m *MatrixBitSet) NextSet(i uint) (uint, bool) {
	m.panicOverSized(i)
	return m.nextSet(i)
}

func (m *MatrixBitSet) PrevSet(i uint) (uint, bool) {
	m.panicOverSized(i)
	return m.prevSet(i)
}

func (m *MatrixBitSet) FormatWord(w uint64) string {
	var b strings.Builder

	for i := 63; i >= 0; i-- {
		v := (w >> uint(i)) & 0x01
		fmt.Fprintf(&b, "%d", v)
	}
	return b.String()
}

// Flip every bit in place
func (m *MatrixBitSet) Invert() *MatrixBitSet {
	for i, w := range m.B {
		m.B[i] = ^w
	}
	return m
}

// Flips rows, cols, returns a new M2
func (m *MatrixBitSet) Transpose() *MatrixBitSet {
	result := NewMatrixBitSet(m.R, m.C)
	for i, e := m.nextSet(0); e; i, e = m.nextSet(i + 1) {
		r, c := i/m.C, i%m.C
		result.set(result.index(c, r))
	}
	return result
}

// Shamelessly borrowed from https://github.com/willf/bitset
func (m *MatrixBitSet) nextSet(i uint) (uint, bool) {
	x := int(i >> log2WordSize)
	if x >= len(m.B) {
		return 0, false
	}
	w := m.B[x]
	w = w >> (i & (wordSize - 1))
	if w != 0 {
		return i + uint(bits.TrailingZeros64(w)), true
	}
	x = x + 1
	for x < len(m.B) {
		if m.B[x] != 0 {
			return uint(x)*wordSize + uint(bits.TrailingZeros64(m.B[x])), true
		}
		x = x + 1
	}
	return 0, false
}

// Returns the previous bit, not including the current bit
// Returns false if no previous bits are set
func (m *MatrixBitSet) prevSet(i uint) (uint, bool) {
	x := int(i >> log2WordSize)
	mask := wordSize - 1
	if x >= len(m.B) {
		return 0, false
	}
	w := m.B[x]
	// if not the low bit, test for others to its right
	// otherwise, go to the previous word
	if i&(wordSize-1) > 0 {
		// mask off i and above
		w = w & ^(allBits << i)
		if w != 0 {
			// return the highbit (previous)
			return uint(x)*wordSize + (mask - uint(bits.LeadingZeros64(w))), true
		}
	}
	x = x - 1
	for x >= 0 {
		if m.B[x] != 0 {
			// return the highbit (previous)
			return uint(x)*wordSize + (mask - uint(bits.LeadingZeros64(m.B[x]))), true
		}
		x = x - 1
	}
	return 0, false
}

func (m *MatrixBitSet) index(r, c uint) uint {
	return (r * m.C) + c
}

func (m *MatrixBitSet) asRC(n uint) (uint, uint) {
	return n / m.C, n % m.C
}

func (m *MatrixBitSet) point(n uint) string {
	r, c := m.asRC(n)
	return fmt.Sprintf("[%d, %d]", r, c)
}

func (m *MatrixBitSet) testPos(mp MatrixPos) int {
	if mp.Valid() && m.test(mp.N()) {
		return 1
	}
	return 0
}

// Shamelessly borrowed from https://github.com/willf/bitset
func (m *MatrixBitSet) test(i uint) bool {
	return m.B[i>>log2WordSize]&(1<<(i&(wordSize-1))) != 0
}

// Shamelessly borrowed from https://github.com/willf/bitset
func (m *MatrixBitSet) set(i uint) *MatrixBitSet {
	m.B[i>>log2WordSize] |= 1 << (i & (wordSize - 1))
	return m
}

// Shamelessly borrowed from https://github.com/willf/bitset
func (m *MatrixBitSet) clear(i uint) *MatrixBitSet {
	m.B[i>>log2WordSize] &^= 1 << (i & (wordSize - 1))
	return m
}

func (m *MatrixBitSet) panicPastMatrix(r, c uint) {
	if r >= m.R || c >= m.C {
		panic(fmt.Sprintf("[%d, %d] exceeds matrix bounds %d x %d", r, c, m.R, m.C))
	}
}

func (m *MatrixBitSet) panicOverSized(i uint) {
	if i >= m.R*m.C {
		panic(fmt.Sprintf("[%d] exceeds matrix bounds %d", i, m.R*m.C))
	}
}
