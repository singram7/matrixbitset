package matrixbitset

type MatrixBounds struct {
	MinR, MinC               uint
	MaxR, MaxC               uint
	left, right, top, bottom uint
	verty                    []uint
	vertx                    []uint
	M                        *MatrixBitSet
}

func NewMatrixBounds(m *MatrixBitSet, points LinearRing) *MatrixBounds {
	bounds := &MatrixBounds{M: m, MinR: 1000000, MinC: 1000000}
	bounds.vertx = make([]uint, 0, len(points)-1)
	bounds.verty = make([]uint, 0, len(points)-1)
	for i, mp := range points {
		if mp.c > bounds.MaxC {
			bounds.MaxC = mp.c
			bounds.right = mp.c
		}
		if mp.r > bounds.MaxR {
			bounds.MaxR = mp.r
			bounds.bottom = mp.r
		}
		if mp.c < bounds.MinC {
			bounds.MinC = mp.c
			bounds.left = mp.c
		}
		if mp.r < bounds.MinR {
			bounds.MinR = mp.r
			bounds.top = mp.r
		}
		// Exclude the duplicate start point
		if i < len(points)-1 {
			bounds.vertx = append(bounds.vertx, mp.c)
			bounds.verty = append(bounds.verty, mp.r)
		}
	}

	return bounds
}

func (mb *MatrixBounds) Width() uint {
	return mb.MaxC - mb.MinC
}

func (mb *MatrixBounds) Height() uint {
	return mb.MaxR - mb.MinR
}

// These are all relative to an UpperLeft Origin
// and are artificial points, ie. not on the set bits
func (mb *MatrixBounds) UpperLeftN() uint {
	return mb.M.index(mb.MinR, mb.MinC)
}

func (mb *MatrixBounds) UpperRightN() uint {
	return mb.M.index(mb.MinR, mb.MaxC)
}

func (mb *MatrixBounds) LowerLeftN() uint {
	return mb.M.index(mb.MaxR, mb.MinC)
}

func (mb *MatrixBounds) LowerRightN() uint {
	return mb.M.index(mb.MaxR, mb.MaxC)
}

func (mb *MatrixBounds) UppperLeft() (uint, uint) {
	return mb.MinR, mb.MinC
}

func (mb *MatrixBounds) UppperRight() (uint, uint) {
	return mb.MinR, mb.MaxC
}

func (mb *MatrixBounds) LowerLeft() (uint, uint) {
	return mb.MaxR, mb.MinC
}

func (mb *MatrixBounds) LowerRight() (uint, uint) {
	return mb.MaxR, mb.MaxC
}

func (mb *MatrixBounds) LastRow() uint {
	return mb.MaxR
}

func (mb *MatrixBounds) LastCol() uint {
	return mb.MaxC
}

func (mb *MatrixBounds) setup() *MatrixBounds {
	stride := mb.M.C
	mb.verty = []uint{mb.left / stride, mb.top / stride, mb.right / stride, mb.bottom / stride}
	mb.vertx = []uint{mb.left % stride, mb.top % stride, mb.right % stride, mb.bottom % stride}
	return mb
}

func (mb *MatrixBounds) Contains(mp MatrixPos) bool {
	return mp.c >= mb.MinC && mp.c <= mb.MaxC && mp.r >= mb.MinR && mp.r <= mb.MaxR
}

func (mb *MatrixBounds) panicIfOutside() {
	mb.M.panicOverSized(mb.left)
	mb.M.panicOverSized(mb.top)
	mb.M.panicOverSized(mb.right)
	mb.M.panicOverSized(mb.bottom)
}
