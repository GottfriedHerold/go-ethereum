package bandersnatch

import "io"

// Point_efgh describes points (usually on the p253-subgroup of) the bandersnatch curve in E:G, H:F - coordinates (called double-projective), i.e.
// we represent X/Z as E/G and Y/Z as H/F. From a computational view, this effectively means that we use a separate denominator for X and Y (instead of a joint one Z).
// We can recover X:Y:Z coordinates by computing Z = F*G, X = E*F, Y = G*H. Then T = E*H. This is meaningful even if one of E,G is zero. There are no rational points with F=0 or H=0.
// Observe that in fact all default formulae in extended twisted edwards coordinates *produce* points in such efgh coordinates and then transform them into the desired form
// Using double-projective coordinates can be used to make this explicit and can save computation if a coordinate is unused:
// The doubling formula and the endomorphism can be expressed in double-projective coordinates by first converting to extended twisted edwards and then computing the double/endo(rather than the other way round).
// Since these formulae do not use the input's t coordinate, this saves a multiplication.
// On the p253-subgroup, the only coordinate that may be zero is actually e.

// Note: Conversion from X:Y:T:Z to EFGH is available as e.g.
// E:=X, F:=X, G:=Z, H:=T or
// E:=T, F:=X, G:=Y, H:=T or
// E:=X, F:=Z, G:=Z, H:=Y or
// (These first two options have singularities at neutral and affine-order-2, the third option at the points at infinity)
type Point_efgh struct {
	e FieldElement
	f FieldElement
	g FieldElement
	h FieldElement
}

var (
	NeutralElement_efgh     = Point_efgh{e: FieldElementZero, f: FieldElementOne, g: FieldElementOne, h: FieldElementOne}
	orderTwoPoint_efgh      = Point_efgh{e: FieldElementZero, f: FieldElementOne, g: FieldElementOne, h: FieldElementMinusOne}
	exceptionalPoint_1_efgh = Point_efgh{e: FieldElementOne, f: squareRootDbyA_fe, g: FieldElementZero, h: FieldElementOne}
	exceptionalPoint_2_efgh = Point_efgh{e: FieldElementOne, f: squareRootDbyA_fe, g: FieldElementZero, h: FieldElementMinusOne}
)

func (p *Point_efgh) is_normalized() bool {
	return p.f.IsOne() && p.g.IsOne()
}

func (p *Point_efgh) normalize_affine() {
	if p.is_normalized() {
		return
	}
	var temp FieldElement
	temp.Mul(&p.f, &p.g)
	if temp.IsZero() {
		if p.IsNaP() {
			napEncountered("Trying to normalize singular point", false, p)
			*p = Point_efgh{}
			return
		}
		panic("Trying to normalize point at infinity")
	}
	temp.Inv(&temp)
	p.e.MulEq(&p.f)
	p.h.MulEq(&p.g)
	p.e.MulEq(&temp)
	p.h.MulEq(&temp)
	p.f.SetOne()
	p.g.SetOne()
}

func (p *Point_efgh) X_projective() (X FieldElement) {
	X.Mul(&p.e, &p.f)
	return
}

func (p *Point_efgh) Y_projective() (Y FieldElement) {
	Y.Mul(&p.g, &p.h)
	return
}

func (p *Point_efgh) T_projective() (T FieldElement) {
	T.Mul(&p.e, &p.h)
	return
}

func (p *Point_efgh) Z_projective() (Z FieldElement) {
	Z.Mul(&p.f, &p.g)
	return
}

func (p *Point_efgh) X_affine() FieldElement {
	p.normalize_affine()
	return p.e
}

func (p *Point_efgh) Y_affine() FieldElement {
	p.normalize_affine()
	return p.h
}

func (p *Point_efgh) T_affine() (T FieldElement) {
	p.normalize_affine()
	T.Mul(&p.e, &p.h)
	return
}

func (p *Point_efgh) IsNeutralElement() bool {
	// The only valid points with e==0 are the neutral element and the affine order-2 point
	if p.IsNaP() {
		return napEncountered("Comparing NaP with neutral element for efgh", true, p)
	}
	return p.e.IsZero()
}

func (p *Point_efgh) IsNeutralElement_exact() bool {
	return p.IsNeutralElement() && p.f.IsEqual(&p.h)
}

func (p *Point_efgh) IsEqual(other CurvePointRead) bool {
	if p.IsNaP() || other.IsNaP() {
		return napEncountered("NaP encountered when comparing efgh-point with other point", true, p, other)
	}
	switch other_real := other.(type) {
	default:
		var other_x = other_real.X_projective()
		var other_y = other_real.Y_projective()
		// other.x * p.y == other.y * p.x
		other_x.MulEq(&p.g)
		other_x.MulEq(&p.h)
		other_y.MulEq(&p.e)
		other_y.MulEq(&p.f)
		return other_x.IsEqual(&other_y)
	}
}

func (p *Point_efgh) IsEqual_exact(other CurvePointRead) bool {
	temp := p.ExtendedTwistedEdwards()
	return temp.IsEqual_exact(other)
}

func (p *Point_efgh) IsAtInfinity() bool {
	if p.IsNaP() {
		return napEncountered("NaP encountered when asking where efgh-point is at infinity", true, p)
	}
	// The only valid points with g==0 are are those at infinity
	return p.g.IsZero()
}

// NaP points have either f==h==0 ("true" NaP-type1) or e==g==0 ("true" NaP-type2) or e==h==0 (result of working on affine NaP).
// Note that no valid points ever have h==0 or f==0.

func (p *Point_efgh) IsNaP() bool {
	if p.h.IsZero() {
		if !(p.f.IsZero() || p.e.IsZero()) {
			panic("efgh-Point is NaP with h==0, but ef != 0")
		}
		return true
	}

	if p.g.IsZero() && p.e.IsZero() {
		return true
		panic("Type-2 efgh NaP encountered") // This is for testing only. -- remove / reconsider later; maybe we can avoid NaP-type2.
	}

	if p.f.IsZero() {
		panic("efgh-Point with f==0 and h!=0 encountered")
	}

	return false
}

// Note: Going eghj -> axtw directly is cheaper by 1 multiplication compared to going via xtw.
// The reason is that we normalize first and then compute the t coordinate. This effectively saves comptuing t *= z^-1.

func (P *Point_efgh) AffineExtended() (ret Point_axtw) {
	P.normalize_affine()
	ret.x = P.e
	ret.y = P.h
	ret.t.Mul(&P.e, &P.h)
	return
}

func (P *Point_efgh) ExtendedTwistedEdwards() (ret Point_xtw) {
	ret.x.Mul(&P.e, &P.f)
	ret.y.Mul(&P.g, &P.h)
	ret.t.Mul(&P.e, &P.h)
	ret.z.Mul(&P.f, &P.g)
	return
}

func (P *Point_efgh) Clone() CurvePointRead {
	p_copy := *P
	return &p_copy
}

func (p *Point_efgh) SerializeShort(output io.Writer) (bytes_written int, err error) {
	return default_SerializeShort(p, output)
}

func (p *Point_efgh) SerializeLong(output io.Writer) (bytes_written int, err error) {
	return default_SerializeLong(p, output)
}

func (p *Point_efgh) String() string {
	return "E=" + p.e.String() + " F=" + p.f.String() + " G=" + p.g.String() + " H=" + p.h.String()
}

func (p *Point_efgh) Add(x, y CurvePointRead) {
	switch x := x.(type) {
	case *Point_xtw:
		switch y := y.(type) {
		case *Point_xtw:
			p.add_stt(x, y)
		case *Point_axtw:
			p.add_sta(x, y)
		default: // including *Point_efgh
			var y_conv Point_xtw = y.ExtendedTwistedEdwards()
			p.add_stt(x, &y_conv)
		}
	case *Point_axtw:
		switch y := y.(type) {
		case *Point_xtw:
			p.add_sta(y, x)
		case *Point_axtw:
			p.add_saa(x, y)
		default: // including *Point_efgh
			var y_conv Point_xtw = y.ExtendedTwistedEdwards()
			p.add_sta(&y_conv, x)
		}
	default:
		var x_conv Point_xtw = x.ExtendedTwistedEdwards()
		p.Add(&x_conv, y)
	}
}

func (p *Point_efgh) Sub(x, y CurvePointRead) {
	switch x := x.(type) {
	case *Point_xtw:
		switch y := y.(type) {
		case *Point_xtw:
			p.sub_stt(x, y)
		case *Point_axtw:
			p.sub_sta(x, y)
		default:
			var y_conv Point_xtw = y.ExtendedTwistedEdwards()
			p.sub_stt(x, &y_conv)
		}
	case *Point_axtw:
		switch y := y.(type) {
		case *Point_xtw:
			p.sub_sat(x, y)
		case *Point_axtw:
			p.sub_saa(x, y)
		default:
			var y_conv Point_xtw = y.ExtendedTwistedEdwards()
			p.sub_sat(x, &y_conv)
		}
	default:
		var x_conv Point_xtw = x.ExtendedTwistedEdwards()
		p.Sub(&x_conv, y)
	}
}

func (p *Point_efgh) Double(x CurvePointRead) {
	// TODO: improve!
	default_Double(p, x)
}

func (p *Point_efgh) Neg(input CurvePointRead) {
	switch input := input.(type) {
	case *Point_efgh:
		p.neg_ss(input)
	default:
		p.SetFrom(input)
		p.NegEq()
	}
}

func (p *Point_efgh) Endo(input CurvePointRead) {
	switch input := input.(type) {
	case *Point_efgh:
		p.computeEndomorphism_ss(input)
	case *Point_xtw:
		p.computeEndomorphism_st(input)
	case *Point_axtw:
		p.computeEndomorphism_sa(input)
	default:
		var input_conv = input.ExtendedTwistedEdwards()
		p.computeEndomorphism_st(&input_conv)
	}
}

func (p *Point_efgh) Endo_fullCurve(input CurvePointRead) {
	switch input := input.(type) {
	case *Point_efgh:
		if input.IsAtInfinity() {
			*p = orderTwoPoint_efgh
		} else {
			p.computeEndomorphism_ss(input)
		}
	case *Point_axtw:
		p.computeEndomorphism_sa(input)
	case *Point_xtw:
		if input.IsAtInfinity() {
			*p = orderTwoPoint_efgh
		} else {
			p.computeEndomorphism_st(input)
		}
	default:
		if input.IsAtInfinity() {
			*p = orderTwoPoint_efgh
		} else {
			var input_conv = input.ExtendedTwistedEdwards()
			p.computeEndomorphism_st(&input_conv)
		}
	}
}

func (p *Point_efgh) SetNeutral() {
	*p = NeutralElement_efgh
}

func (p *Point_efgh) AddEq(input CurvePointRead) {
	p.Add(p, input)
}

func (p *Point_efgh) SubEq(input CurvePointRead) {
	p.Sub(p, input)
}

func (p *Point_efgh) DoubleEq() {
	p.Double(p)
}

func (p *Point_efgh) NegEq() {
	p.e.NegEq()
}

func (p *Point_efgh) EndoEq() {
	p.computeEndomorphism_ss(p)
}

// Note: We usually want to convert FROM efgh to other types, not TO efgh. So this function is rarely used.

func (p *Point_efgh) SetFrom(input CurvePointRead) {
	switch input := input.(type) {
	case *Point_efgh:
		*p = *input
	case *Point_xtw:
		if !input.z.IsZero() {
			// usual case: This is singular iff input is at infinity (which means y==z==0)
			p.e = input.x
			p.f = input.z
			p.g = input.z
			p.h = input.y
		} else { // Point at infinity or NaP
			// usually equivalent to the above, but singular iff input has x==t==0
			p.e = input.x
			p.f = input.x
			p.g.SetZero() // = input.z
			p.h = input.t
		}
	case *Point_axtw:
		p.e = input.x
		p.f.SetOne()
		p.g.SetOne()
		p.h = input.y
	default:
		if input.IsNaP() {
			napEncountered("Trying to convert NaP of unknown type to efgh", false, input)
			*p = Point_efgh{}
		} else if !input.IsAtInfinity() {
			p.e = input.X_projective()
			p.f = input.Z_projective()
			p.g = p.f
			p.h = input.Y_projective()
		} else {
			// The general interface does not allow to distinguish the two points at infinity.
			// We could fix that, but it seems hardly worth it.
			panic("Trying to convert point of unknown type in efgh when point is at infinity")
		}
	}
}

func (p *Point_efgh) DeserializeShort(input io.Reader, trusted bool) (bytes_read int, err error) {
	return default_DeserializeShort(p, input, trusted)
}

func (p *Point_efgh) DeserializeLong(input io.Reader, trusted bool) (bytes_read int, err error) {
	return default_DeserializeLong(p, input, trusted)
}

func (p *Point_efgh) DeserializeAuto(input io.Reader, trusted bool) (bytes_read int, err error) {
	return default_DeserializeAuto(p, input, trusted)
}
