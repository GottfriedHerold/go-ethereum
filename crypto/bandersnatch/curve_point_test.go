package bandersnatch

import (
	"math/big"
	"strconv"
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
		defer func() { recover() }()
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

func make_checkfun_subtraction(receiverType PointType) (returned_function checkfunction) {
	returned_function = func(s TestSample) (bool, string) {
		s.AssertNumberOfPoints(2)

		// If the points in the sample are outside of the subgroup, we might hit the singular cases of addition/subtraction.
		// For addition, we have a Case - flag, but not for subtraction. We just skip the test.
		if s.AnyFlags().CheckFlag(Case_outside_goodgroup) {
			return true, ""
		}

		var singular bool = s.AnyFlags().CheckFlag(Case_singular)
		var result_of_subtraction CurvePoint = MakeCurvePointFromType(receiverType)
		var negative_of_point CurvePoint = MakeCurvePointFromType(receiverType)
		var result1 CurvePoint = MakeCurvePointFromType(receiverType)
		var result2 CurvePoint = MakeCurvePointFromType(receiverType)

		result_of_subtraction.Sub(s.Points[0], s.Points[1])
		result1.Add(result_of_subtraction, s.Points[1])
		var got bool
		var expected bool = !singular
		if wasInvalidPointEncountered(func() { got = result1.IsEqual_exact(s.Points[0]) }) != singular {
			return false, "Wrong NaP behaviour when checking (P-Q) + Q ?= P"
		}
		if got != expected {
			return false, "(P-Q) + Q != P"
		}

		// Check that P - Q == P + (-Q)
		negative_of_point.Neg(s.Points[1])
		result2.Add(s.Points[0], negative_of_point)
		if wasInvalidPointEncountered(func() { got = result2.IsEqual_exact(result_of_subtraction) }) != singular {
			return false, "Wrong NaP behaviour when checking P - Q ?= P + (-Q)"
		}
		if got != expected {
			return false, "P - Q != P + (-Q)"
		}
		return true, ""
	}
	return
}

func make_checkfun_alias(receiverType PointType) (returned_function checkfunction) {
	returned_function = func(s TestSample) (bool, string) {
		s.AssertNumberOfPoints(1)

		var singular bool = s.AnyFlags().CheckFlag(Case_singular)
		var result1 CurvePoint = MakeCurvePointFromType(receiverType)
		var result2 CurvePoint = MakeCurvePointFromType(receiverType)
		var clone1, clone2, clone3, clone4 CurvePoint

		clone1 = s.Points[0].Clone().(CurvePoint)
		clone2 = s.Points[0].Clone().(CurvePoint)
		clone3 = s.Points[0].Clone().(CurvePoint)
		clone4 = s.Points[0].Clone().(CurvePoint)
		clone1.Add(clone1, clone1)
		result1.Add(clone2, clone2)
		result2.Add(clone3, clone4)

		if singular {
			return clone1.IsSingular() && result1.IsSingular() && result2.IsSingular(), "Alias test for add did not get NaP when expected"
		}
		if !(clone1.IsEqual_exact(result1) && clone1.IsEqual_exact(result2)) {
			return false, "Addition gives inconsistent results when arguments alias"
		}

		clone1 = s.Points[0].Clone().(CurvePoint)
		clone2 = s.Points[0].Clone().(CurvePoint)
		clone3 = s.Points[0].Clone().(CurvePoint)
		clone4 = s.Points[0].Clone().(CurvePoint)
		clone1.Sub(clone1, clone1)
		result1.Sub(clone2, clone2)
		result2.Sub(clone3, clone4)
		if singular {
			return clone1.IsSingular() && result1.IsSingular() && result2.IsSingular(), "Alias test for sub did not get NaP when expected"
		}
		if !(clone1.IsEqual_exact(result1) && clone1.IsEqual_exact(result2)) {
			return false, "Subtraction gives inconsistent results when arguments alias"
		}

		var expected bool = !singular
		var got bool
		clone1 = s.Points[0].Clone().(CurvePoint)
		if wasInvalidPointEncountered(func() { got = clone1.IsEqual_exact(clone1) }) != singular {
			return false, "P =? P did not handle NaP"
		}
		if got != expected {
			return false, "P = P did not work as expected for aliasing arguments"
		}

		clone1 = s.Points[0].Clone().(CurvePoint)
		if wasInvalidPointEncountered(func() { got = clone1.IsEqual(clone1) }) != singular {
			return false, "P =? P did not handle NaP"
		}
		if got != expected {
			return false, "P = P did not work as expected for aliasing arguments"
		}

		clone1 = s.Points[0].Clone().(CurvePoint)
		clone2 = s.Points[0].Clone().(CurvePoint)
		clone1.Neg(clone1)
		result1.Neg(clone2)

		if wasInvalidPointEncountered(func() { got = clone1.IsEqual(result1) }) != singular {
			return false, "-P ?= -P did not trigger error handler on NaP"
		}
		if got != expected {
			return false, "Computing negative did not work when receiver aliases argument"
		}

		if !s.AnyFlags().CheckFlag(Case_infinite) { // Endo does not work correctly on infinte points, only Endo_safe does
			clone1 = s.Points[0].Clone().(CurvePoint)
			clone2 = s.Points[0].Clone().(CurvePoint)
			clone1.Endo(clone1)
			result1.Endo(clone2)

			if wasInvalidPointEncountered(func() { got = clone1.IsEqual(result1) }) != singular {
				return false, "Endo(P) ?= Endo(P) did not trigger error handler on NaP, was expecting:" + strconv.FormatBool(singular)
			}
			if got != expected {
				return false, "Computing Endomorphism did not work when receiver aliases argument"
			}
		}

		clone1 = s.Points[0].Clone().(CurvePoint)
		clone2 = s.Points[0].Clone().(CurvePoint)
		clone1.Endo_safe(clone1)
		result1.Endo_safe(clone2)

		if wasInvalidPointEncountered(func() { got = clone1.IsEqual(result1) }) != singular {
			return false, "Endo_safe(P) ?= Endo_safe(P) did not trigger error handler on NaP"
		}
		if got != expected {
			return false, "Computing Endomorphism (for full curve) did not work when receiver aliases argument"
		}

		clone1 = s.Points[0].Clone().(CurvePoint)
		result1 = s.Points[0].Clone().(CurvePoint)
		clone1.SetFrom(clone1)

		if wasInvalidPointEncountered(func() { got = clone1.IsEqual(result1) }) != singular {
			return false, "Comparison of supposedly identical points did not trigger error handler on NaP"
		}
		if got != expected {
			return false, "SetFrom did not work when receiver aliases argument"
		}

		return true, ""
	}
	return
}

func checkfun_associative_law(s TestSample) (bool, string) {
	s.AssertNumberOfPoints(3)
	var singular bool = s.AnyFlags().CheckFlag(Case_singular)
	var result1 CurvePoint = MakeCurvePointFromType(GetPointType(s.Points[0]))
	var result2 CurvePoint = MakeCurvePointFromType(GetPointType(s.Points[0]))

	result1.Add(s.Points[0], s.Points[1])
	result1.Add(result1, s.Points[2])
	result2.Add(s.Points[1], s.Points[2])
	result2.Add(s.Points[0], result2)

	var got bool
	var expected bool = !singular
	if wasInvalidPointEncountered(func() { got = result1.IsEqual_exact(result2) }) != singular {
		return false, "Wrong NaP behvaiour when checking (P+Q)+R ?= P+(Q+R)"
	}
	return got == expected, ""
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

	make_samples2_and_run_tests(t, make_checkfun_subtraction(pointType), "Subtraction did not work as expected"+point_string, pointType, pointType, 10, excluded_flags)
	for _, type1 = range allTypes {
		for _, type2 = range allTypes {
			make_samples2_and_run_tests(t, make_checkfun_subtraction(pointType), "Subtraction did not work as expected"+point_string, type1, type2, 10, excluded_flags|Case_infinite)
		}
	}

	for _, type1 = range allTypes {
		for _, type2 = range allTypes {
			samples := MakeTestSamples3(5, pointType, type1, type2, excluded_flags|Case_outside_goodgroup)
			run_tests_on_samples(checkfun_associative_law, t, samples, "Associative law does not hold "+point_string+" "+PointTypeToString(type1)+" "+PointTypeToString(type2))
		}
	}

	make_samples1_and_run_tests(t, make_checkfun_alias(pointType), "Aliasing did not work as expected"+point_string, pointType, 10, excluded_flags)
	for _, type1 = range allTypes {
		make_samples1_and_run_tests(t, make_checkfun_alias(pointType), "Aliasing did not work as expected"+point_string, type1, 10, excluded_flags|Case_infinite)
	}

}

func TestGeneralTestsForXTW(t *testing.T) {
	test_general(t, pointTypeXTW, 0)
}

func TestGeneralTestForAXTW(t *testing.T) {
	test_general(t, pointTypeAXTW, Case_infinite)
}
