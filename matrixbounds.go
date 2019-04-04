package matrixbitset

type MatrixBounds struct {
	MinR, MinC               uint
	MaxR, MaxC               uint
	left, right, top, bottom uint
	verty                    []uint
	vertx                    []uint
	M                        *MatrixBitSet
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

func (mb *MatrixBounds) panicIfOutside() {
	mb.M.panicOverSized(mb.left)
	mb.M.panicOverSized(mb.top)
	mb.M.panicOverSized(mb.right)
	mb.M.panicOverSized(mb.bottom)
}
