package bandersnatch

import (
	"math/big"
	"testing"
)

func TestGlobalParameter(t *testing.T) {
	if big.Jacobi(big.NewInt(TwistedEdwardsA), BaseFieldSize) == 1 {
		t.Fatal("Parameter a of curve is a square")
	}
	if big.Jacobi(TwistedEdwardsD_Int, BaseFieldSize) == 1 {
		t.Fatal("Parameter d of curve is a square")
	}
	var temp FieldElement
	temp.Square(&SqrtDDivA_fe)
	temp.multiply_by_five()
	temp.Neg(&temp)
	if !temp.IsEqual(&TwistedEdwardsD_fe) {
		t.Fatal("SqrtDDivA is not a square root of d/a")
	}
}

func TestInterfaces(t *testing.T) {
	var _ CurvePointRead = &Point_xtw{}
	var _ CurvePointWrite = &Point_xtw{}
	var _ CurvePointRead = &Point_axtw{}
	var _ CurvePointWrite = &Point_axtw{}
	var _ CurvePointRead = &Point_efgh{}
}

func checkfun_recognize_neutral(s TestSample) bool {
	if s.Len != 1 {
		panic("Wrong type of sample for check")
	}
	var expected bool = s.Flags[0].CheckFlag(Case_zero)
	var got = s.Points[0].IsNeutralElement()
	return expected == got
}

func checkfun_recognize_neutral_exact(s TestSample) bool {
	if s.Len != 1 {
		panic("Wrong type of sample for check")
	}
	var expected bool = s.Flags[0].CheckFlag(Case_zero_exact)
	var got = s.Points[0].IsNeutralElement_exact()
	return expected == got
}

func checkfun_recognize_infinity(s TestSample) bool {
	if s.Len != 1 {
		panic("Wrong type of sample for check")
	}
	var expected bool = s.Flags[0].CheckFlag(Case_infinite)
	var got = s.Points[0].IsAtInfinity()
	return expected == got
}

func checkfun_recognize_singularities(s TestSample) bool {
	if s.Len != 1 {
		panic("Wrong type of sample for check")
	}
	var expected bool = s.Flags[0].CheckFlag(Case_singular)
	var got = s.Points[0].IsSingular()
	return expected == got
}

func checkfun_recognize_equality(s TestSample) bool {
	if s.Len != 2 {
		panic("Wrong type of sample for check")
	}
	var expected bool = s.AnyFlags().CheckFlag(Case_equal)
	var got = s.Points[0].IsEqual(s.Points[1])
	return expected == got
}

func checkfun_recognize_equality_exact(s TestSample) bool {
	if s.Len != 2 {
		panic("Wrong type of sample for check")
	}
	var expected bool = s.AnyFlags().CheckFlag(Case_equal_exact)
	var got = s.Points[0].IsEqual_exact(s.Points[1])
	return expected == got
}

func checkfun_conversion_to_affine(s TestSample) bool {
	if s.Len != 1 {
		panic("Wrong type of sample for check")
	}
	var singular bool = s.Flags[0].CheckFlag(Case_singular)
	var infinite bool = s.Flags[0].CheckFlag(Case_infinite)
	if singular || infinite {
		return true
	}
	var affine_point Point_axtw = s.Points[0].AffineExtended()
	return affine_point.IsEqual_exact(s.Points[0])
}

func checkfun_conversion_to_xtw(s TestSample) bool {
	if s.Len != 1 {
		panic("Wrong type of sample for check")
	}
	var singular bool = s.Flags[0].CheckFlag(Case_singular)
	if singular {
		return true // TODO: Should we expect some specific behaviour?
	}
	var point_copy Point_xtw = s.Points[0].ExtendedTwistedEdwards()
	return point_copy.IsEqual_exact(s.Points[0])
}

func checkfun_clone(s TestSample) bool {
	if s.Len != 1 {
		panic("Wrong type of sample for check")
	}
	var singular bool = s.Flags[0].CheckFlag(Case_singular)
	var point_copy CurvePointRead = s.Points[0].Clone()
	if singular {
		return !point_copy.IsEqual_exact(s.Points[0])
	}
	return point_copy.IsEqual_exact(s.Points[0])
}

/*
func checkfun_addition_commutes(s TestSample) bool {
	if s.Len != 2 {
		panic("Wrong type of sample for check")
	}
}
*/

func test_general(t *testing.T, pointType PointType, excluded_flags PointFlags) {
	point_string := PointTypeToString(pointType)
	make_samples1_and_run_tests(t, checkfun_recognize_neutral, "Did not recognize neutral element for "+point_string, pointType, 10, excluded_flags)
	make_samples1_and_run_tests(t, checkfun_recognize_neutral_exact, "Did not recognize exact neutral element for "+point_string, pointType, 10, excluded_flags)
	make_samples1_and_run_tests(t, checkfun_recognize_infinity, "Did not recognize infinite points"+point_string, pointType, 10, excluded_flags)
	make_samples1_and_run_tests(t, checkfun_recognize_singularities, "Did not recognize invalid points arising from singularities"+point_string, pointType, 10, excluded_flags)
	make_samples2_and_run_tests(t, checkfun_recognize_equality, "Did not recognize equality"+point_string, pointType, pointType, 10, excluded_flags)
	make_samples2_and_run_tests(t, checkfun_recognize_equality_exact, "Did not recognize exact equality"+point_string, pointType, pointType, 10, excluded_flags)
	make_samples1_and_run_tests(t, checkfun_conversion_to_affine, "Conversion to affine did not work"+point_string, pointType, 10, excluded_flags)
	make_samples1_and_run_tests(t, checkfun_conversion_to_xtw, "Conversion to xtw did not work"+point_string, pointType, 10, excluded_flags)
	make_samples1_and_run_tests(t, checkfun_clone, "cloning did not work"+point_string, pointType, 10, excluded_flags)
}

func TestGeneralTestsForXTW(t *testing.T) {
	test_general(t, pointTypeXTW, 0)
}

func TestGeneralTestForAXTW(t *testing.T) {
	test_general(t, pointTypeAXTW, Case_infinite)
}
