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

func checkfun_recognize_neutral(s TestSample) (bool, string) {
	s.AssertNumberOfPoints(1)
	var singular = s.AnyFlags().CheckFlag(Case_singular)
	var expected bool = s.Flags[0].CheckFlag(Case_zero)
	var got bool
	if wasInvalidPointEncountered(func() { got = s.Points[0].IsNeutralElement() }) != singular {
		return false, "NaP not recongized"
	}
	return expected == got, ""
}

func checkfun_recognize_neutral_exact(s TestSample) (bool, string) {
	s.AssertNumberOfPoints(1)
	var singular = s.AnyFlags().CheckFlag(Case_singular)
	var expected bool = s.Flags[0].CheckFlag(Case_zero_exact)
	var got bool
	if wasInvalidPointEncountered(func() { got = s.Points[0].IsNeutralElement_exact() }) != singular {
		return false, "NaP not recoginized"
	}
	return expected == got, ""
}

func checkfun_recognize_infinity(s TestSample) (bool, string) {
	s.AssertNumberOfPoints(1)
	var singular = s.AnyFlags().CheckFlag(Case_singular)
	var expected bool = s.Flags[0].CheckFlag(Case_infinite)
	var got bool

	if wasInvalidPointEncountered(func() { got = s.Points[0].IsAtInfinity() }) != singular {
		return false, "NaP not recognized"
	}
	return expected == got, ""
}

func checkfun_recognize_singularities(s TestSample) (bool, string) {
	s.AssertNumberOfPoints(1)
	var expected bool = s.Flags[0].CheckFlag(Case_singular)
	var got = s.Points[0].IsSingular()
	return expected == got, "Test sample marked as singular, but IsSingular() does not agree"
}

func checkfun_recognize_equality(s TestSample) (bool, string) {
	s.AssertNumberOfPoints(2)
	var singular bool = s.AnyFlags().CheckFlag(Case_singular)
	var expected bool = s.AnyFlags().CheckFlag(Case_equal)
	var got bool
	if wasInvalidPointEncountered(func() { got = s.Points[0].IsEqual(s.Points[1]) }) != singular {
		return false, "NaP not regognized"
	}
	return expected == got, ""
}

func checkfun_recognize_equality_exact(s TestSample) (bool, string) {
	s.AssertNumberOfPoints(2)
	var singular bool = s.AnyFlags().CheckFlag(Case_singular)
	var expected bool = s.AnyFlags().CheckFlag(Case_equal_exact)
	var got bool
	if wasInvalidPointEncountered(func() { got = s.Points[0].IsEqual_exact(s.Points[1]) }) != singular {
		return false, "NaP not regognized"
	}
	return expected == got, ""
}

func checkfun_conversion_to_affine(s TestSample) (ok bool, error_reason string) {
	s.AssertNumberOfPoints(1)
	var singular bool = s.Flags[0].CheckFlag(Case_singular)
	var infinite bool = s.Flags[0].CheckFlag(Case_infinite)
	var affine_point Point_axtw
	if singular {
		affine_point = s.Points[0].AffineExtended()
		return affine_point.IsSingular(), "conversion to affine of NaP does not result in NaP"
	}
	if infinite {
		return true, "" // FIXME
		ok = true       // return value in case of a recover()'ed panic
		defer recover()
		affine_point = s.Points[0].AffineExtended()
		return affine_point.IsSingular(), "conversion to affine of ininite point neither panics nor results in NaP"
	}
	affine_point = s.Points[0].AffineExtended()
	return affine_point.IsEqual_exact(s.Points[0]), ""
}

func checkfun_conversion_to_xtw(s TestSample) (bool, string) {
	s.AssertNumberOfPoints(1)
	var singular bool = s.Flags[0].CheckFlag(Case_singular)
	var point_xtw Point_xtw
	if singular {
		point_xtw = s.Points[0].ExtendedTwistedEdwards()
		return point_xtw.IsSingular(), "conversion of NaP to xtw point did not result in NaP"
	}
	point_xtw = s.Points[0].ExtendedTwistedEdwards()
	return point_xtw.IsEqual_exact(s.Points[0]), "conversion to xtw did not result in point that was considered equal"
}

func checkfun_clone(s TestSample) (bool, string) {
	s.AssertNumberOfPoints(1)
	var singular bool = s.Flags[0].CheckFlag(Case_singular)
	var point_copy CurvePointRead = s.Points[0].Clone()
	if singular {
		if !point_copy.IsSingular() {
			return false, "cloning NaP did not result in a NaP"
		}
		return !point_copy.IsEqual_exact(s.Points[0]), "cloning NaP resulted in point that was considered equal to the original"
	}
	return point_copy.IsEqual_exact(s.Points[0]), ""
}

func make_checkfun_addition_commutes(receiverType PointType) (returned_function checkfunction) {
	returned_function = func(s TestSample) (bool, string) {
		s.AssertNumberOfPoints(2)
		var singular bool = s.AnyFlags().CheckFlag(Case_singular)
		var result1 CurvePoint = MakeCurvePointFromType(receiverType)
		var result2 CurvePoint = MakeCurvePointFromType(receiverType)
		result1.Add(s.Points[0], s.Points[1])
		result2.Add(s.Points[1], s.Points[0])
		var expected, got1, got2 bool
		expected = !singular
		if wasInvalidPointEncountered(func() { got1 = result1.IsEqual(result2) }) != singular {
			return false, "comparison of P+Q =? Q+P with NaPs involved did not trigger error handler"
		}
		if wasInvalidPointEncountered(func() { got2 = result1.IsEqual_exact(result2) }) != singular {
			return false, "exact Comparison of P+Q =? Q+P with NaPs involved did not trigger error handler"
		}
		return expected == got1 && expected == got2, ""
	}
	return
}

// ensure that P + neutral element == P for P finite curve point (for infinite P, the addition law does not work for P + neutral element)
func make_checkfun_addition_of_zero(receiverType PointType, zeroType PointType) (returned_function checkfunction) {
	var zero CurvePoint = MakeCurvePointFromType(zeroType)
	zero.SetNeutral()
	returned_function = func(s TestSample) (bool, string) {
		var singular bool = s.AnyFlags().CheckFlag(Case_singular)
		if s.AnyFlags().CheckFlag(Case_infinite) {
			return true, "" // This case is skipped.
		}
		var result CurvePoint = MakeCurvePointFromType(receiverType)
		result.Add(s.Points[0], zero)
		var expected, got1, got2 bool
		expected = !singular
		if wasInvalidPointEncountered(func() { got1 = result.IsEqual(s.Points[0]) }) != singular {
			return false, "comparison of P + neutral element =? P with NaP P did not trigger error handler"
		}
		if wasInvalidPointEncountered(func() { got2 = result.IsEqual_exact(s.Points[0]) }) != singular {
			return false, "exact omparison of P + neutral element =? P with NaP P did not trigger error handler"
		}
		return expected == got1 && expected == got2, "P + 0 != P"
	}
	return
}

func make_checkfun_negative(receiverType PointType) (returned_function checkfunction) {
	returned_function = func(s TestSample) (bool, string) {
		s.AssertNumberOfPoints(1)
		var singular bool = s.AnyFlags().CheckFlag(Case_singular)
		var negative_of_point CurvePoint = MakeCurvePointFromType(receiverType)
		var sum CurvePoint = MakeCurvePointFromType(receiverType)
		negative_of_point.Neg(s.Points[0])
		if singular != negative_of_point.IsSingular() {
			return false, "Taking negative of NaP did not result in NaP"
		}
		sum.Add(s.Points[0], negative_of_point)
		expected := !singular
		var got bool
		if wasInvalidPointEncountered(func() { got = sum.IsNeutralElement_exact() }) != singular {
			return false, "comparing P + (-P) =? neutral with P NaP did not trigger error handler"
		}
		return expected == got, "P + (-P) != neutral"
	}
	return
}

func test_general(t *testing.T, pointType PointType, excluded_flags PointFlags) {
	allTypes := []PointType{pointTypeXTW, pointTypeAXTW}
	var type1, type2 PointType
	point_string := PointTypeToString(pointType)
	make_samples1_and_run_tests(t, checkfun_recognize_neutral, "Did not recognize neutral element for "+point_string, pointType, 10, excluded_flags)
	make_samples1_and_run_tests(t, checkfun_recognize_neutral_exact, "Did not recognize exact neutral element for "+point_string, pointType, 10, excluded_flags)
	make_samples1_and_run_tests(t, checkfun_recognize_infinity, "Did not recognize infinite points "+point_string, pointType, 10, excluded_flags)
	make_samples1_and_run_tests(t, checkfun_recognize_singularities, "Did not recognize invalid points arising from singularities "+point_string, pointType, 10, excluded_flags)
	make_samples2_and_run_tests(t, checkfun_recognize_equality, "Did not recognize equality "+point_string, pointType, pointType, 10, excluded_flags)
	make_samples2_and_run_tests(t, checkfun_recognize_equality_exact, "Did not recognize exact equality "+point_string, pointType, pointType, 10, excluded_flags)
	for _, type1 = range allTypes {
		make_samples2_and_run_tests(t, checkfun_recognize_equality, "Did not recognize equality "+point_string, pointType, type1, 10, excluded_flags|Case_infinite)
		make_samples2_and_run_tests(t, checkfun_recognize_equality_exact, "Did not recognize exact equality "+point_string, pointType, type1, 10, excluded_flags|Case_infinite)
	}
	make_samples1_and_run_tests(t, checkfun_conversion_to_affine, "Conversion to affine did not work "+point_string, pointType, 10, excluded_flags)
	make_samples1_and_run_tests(t, checkfun_conversion_to_xtw, "Conversion to xtw did not work "+point_string, pointType, 10, excluded_flags)
	make_samples1_and_run_tests(t, checkfun_clone, "cloning did not work"+point_string, pointType, 10, excluded_flags)
	make_samples2_and_run_tests(t, make_checkfun_addition_commutes(pointType), "Addition did not commute for "+point_string, pointType, pointType, 10, excluded_flags|Case_differenceInfinite)
	make_samples2_and_run_tests(t, make_checkfun_addition_commutes(pointType), "Addition did not commute for "+point_string, pointTypeXTW, pointTypeXTW, 10, excluded_flags|Case_differenceInfinite|Case_outside_goodgroup|Case_infinite)
	make_samples2_and_run_tests(t, make_checkfun_addition_commutes(pointType), "Addition did not commute for "+point_string, pointTypeAXTW, pointTypeAXTW, 10, excluded_flags|Case_differenceInfinite|Case_infinite|Case_outside_goodgroup)
	make_samples2_and_run_tests(t, make_checkfun_addition_commutes(pointType), "Addition did not commute for "+point_string, pointTypeXTW, pointTypeAXTW, 10, excluded_flags|Case_differenceInfinite|Case_infinite|Case_outside_goodgroup)

	for _, type1 = range allTypes {
		for _, type2 = range allTypes {
			make_samples1_and_run_tests(t, make_checkfun_addition_of_zero(pointType, type1), "Addition of neutral changes point for"+point_string, type2, 10, excluded_flags|Case_infinite) // Infinite + Neutral will cause a NaP
		}
	}

	make_samples1_and_run_tests(t, make_checkfun_negative(pointType), "Negating points did not work as expected"+point_string, pointType, 10, excluded_flags)
	for _, type1 = range allTypes {
		make_samples1_and_run_tests(t, make_checkfun_negative(pointType), "Negating points did not work as expected"+point_string, type1, 10, excluded_flags|Case_infinite)
	}
}

func TestGeneralTestsForXTW(t *testing.T) {
	test_general(t, pointTypeXTW, 0)
}

func TestGeneralTestForAXTW(t *testing.T) {
	test_general(t, pointTypeAXTW, Case_infinite)
}
