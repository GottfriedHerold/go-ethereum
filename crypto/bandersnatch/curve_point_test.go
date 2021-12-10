package bandersnatch

import (
	"math/big"
	"strconv"
	"testing"
)

/*
	This file contains tests on curve points that can be expressed properties on the exported interface of CurvePoint.
	Using our testing framework and a little bit of reflection (hidden in helper functions) and interfaces, these tests are then run on all concrete curve point types.
*/

// Tests properties of some global parameters
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

// Ensures that types satisfy the intended interfaces.
// Note that the package will not compile anyway if these are not satisfied.
func TestInterfaces(t *testing.T) {
	var _ CurvePointRead = &Point_xtw{}
	var _ CurvePointWrite = &Point_xtw{}
	var _ CurvePointRead = &Point_axtw{}
	var _ CurvePointWrite = &Point_axtw{}
	var _ CurvePointRead = &Point_efgh{}
}

/*
	checkfun_<foo> are functions of type checkfun (i.e. func(TestSample)(bool, string))
	They are to be run on TestSamples containing a Tuple of CurvePoints and Flags and return true, <ignored> on success
	and false, optional_error_reason on failure.

	Be aware that our checkfunction also verify the intended behaviour at NaP's (even though we might not guarantee it)

	In some cases, the checkfunction needs an extra argument.
	E.g. when testing addition z.Add(x,y), the arguments x,y are given by the TestSample, but we need to specify the type of the receiver z intended to store the argument
	(this is important, as it selects the actuall method used), so we need an extra argument of type PointType (which is based on reflect.Type).
	In order to do that, we define functions with names
	make_checkfun_<foo>(extra arguments) that return checkfunctions with the extra arguments bound.
*/

// checks whether IsNeutralElement correctly recognized neutral elements
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

// checks whether IsNeutralElement_exact correctly recognizes neutral elements
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

// checks whether IsAtInfinity correctly recognizes points at infinity.
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

// checks whether IsSingular() correctly recognizes NaPs
func checkfun_recognize_singularities(s TestSample) (bool, string) {
	s.AssertNumberOfPoints(1)
	var expected bool = s.Flags[0].CheckFlag(Case_singular)
	var got = s.Points[0].IsSingular()
	return expected == got, "Test sample marked as singular, but IsSingular() does not agree"
}

// checks whether IsEqual correctly recognizes pairs of equal points (modulo P = P+A)
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

// checks whether IsEqual_exact correctly recognizes pairs of exactly equal points.
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

// checks whether AffineExtended gives a point that is considered equal to the original.
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

// checks whether ExtendedTwistedEdwards gives a point that is considered equal to the original
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

// checks whether Clone() gives a point that is considered equal to the original.
func checkfun_clone(s TestSample) (bool, string) {
	s.AssertNumberOfPoints(1)
	var singular bool = s.Flags[0].CheckFlag(Case_singular)
	var point_copy CurvePoint = s.Points[0].Clone().(CurvePoint)
	var expected = !singular
	var got bool

	if singular != point_copy.IsSingular() {
		return false, "cloning did not result in the same NaP status as the original"
	}

	if wasInvalidPointEncountered(func() { got = point_copy.IsEqual_exact(s.Points[0]) }) != singular {
		return false, "cloning NaP resulted in point that was considered equal to the original"
	}
	if expected != got {
		return false, "Cloning did not result in identical point"
	}

	// modify point_copy and try again to make sure clone is not tied to the original (Note that CurvePoint's concrete values are pointers)
	point_copy.AddEq(&example_generator_xtw)
	expected = false
	if wasInvalidPointEncountered(func() { got = point_copy.IsEqual_exact(s.Points[0]) }) != singular {
		return false, "cloning NaP resulted in point that was considered equal to the original"
	}
	if expected != got {
		return false, "Cloning did not result in identical point"
	}
	return true, ""

}

// check whether P+Q == Q+P
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

// Checks that Neg results in an additive inverse
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

// Checks that subtraction is compatible with addition, i.e. (P-Q)+Q == P and P-Q == P + (-Q)
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

// This tests whether our functions work even if receiver and arguments (which are usually pointers) alias.
// This is probably the most important test.
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

		if wasInvalidPointEncountered(func() { got = clone1.IsEqual_exact(result1) }) != singular {
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

			if wasInvalidPointEncountered(func() { got = clone1.IsEqual_exact(result1) }) != singular {
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

		if wasInvalidPointEncountered(func() { got = clone1.IsEqual_exact(result1) }) != singular {
			return false, "Endo_safe(P) ?= Endo_safe(P) did not trigger error handler on NaP"
		}
		if got != expected {
			return false, "Computing Endomorphism (for full curve) did not work when receiver aliases argument"
		}

		clone1 = s.Points[0].Clone().(CurvePoint)
		result1 = s.Points[0].Clone().(CurvePoint)
		clone1.SetFrom(clone1)

		if wasInvalidPointEncountered(func() { got = clone1.IsEqual_exact(result1) }) != singular {
			return false, "Comparison of supposedly identical points did not trigger error handler on NaP"
		}
		if got != expected {
			return false, "SetFrom did not work when receiver aliases argument"
		}

		clone1 = s.Points[0].Clone().(CurvePoint)
		clone2 = s.Points[0].Clone().(CurvePoint)
		result1 = s.Points[0].Clone().(CurvePoint)
		clone1.AddEq(clone1)
		result1.AddEq(clone2)
		if wasInvalidPointEncountered(func() { got = clone1.IsEqual_exact(result1) }) != singular {
			return false, "Alias-testing on AddEq did not trigger error handler on NaP"
		}
		if got != expected {
			return false, "AddEq does not work on aliased arguments"
		}

		clone1 = s.Points[0].Clone().(CurvePoint)
		clone2 = s.Points[0].Clone().(CurvePoint)
		result1 = s.Points[0].Clone().(CurvePoint)
		clone1.SubEq(clone1)
		result1.SubEq(clone2)
		if wasInvalidPointEncountered(func() { got = clone1.IsEqual_exact(result1) }) != singular {
			return false, "Alias-testing on SubEq did not trigger error handler on NaP"
		}
		if got != expected {
			return false, "SubEq does not work on aliased arguments"
		}

		return true, ""
	}
	return
}

// This function checks the associative law on point addition.
// Note that we assume that our testsamples do not contain triples where exceptional cases for the addition laws occur.
// (This is why the generator for testsamples that contain triples only produces random output)
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

// This function checks whether the endomorphism factors through P=P+A and Endo and Endo_safe agree
func make_checkfun_endo_sane(receiverType PointType) (returned_function checkfunction) {
	returned_function = func(s TestSample) (bool, string) {
		s.AssertNumberOfPoints(1)
		var singular bool = s.AnyFlags().CheckFlag(Case_singular)
		var infinite bool = s.AnyFlags().CheckFlag(Case_infinite)
		var result1 = MakeCurvePointFromType(receiverType)
		var result2 = MakeCurvePointFromType(receiverType)
		result1.Endo_safe(s.Points[0])

		if singular {
			if !result1.IsSingular() {
				return false, "Endo_safe(NaP) did not result in NaP"
			}
			result2.Endo(s.Points[0])
			if !result2.IsSingular() {
				return false, "Endo(NaP) did not result in NaP"
			}
			// No further checks
			return true, ""
		}

		if result1.IsSingular() {
			return false, "Endo_safe(P) resulted in NaP for non-NaP P"
		}

		var result_xtw Point_xtw = result1.ExtendedTwistedEdwards()
		if !result_xtw.isPointOnCurve() {
			return false, "Endo result is not on curve"
		}
		if !result_xtw.legendre_check_point() {
			return false, "Endo result not in subgroup"
		}

		if !infinite { // Endo may not work on points at infinity
			result2.Endo(s.Points[0])
			if !result1.IsEqual_exact(result2) {
				return false, "Endo(P) and Endo_exact(P) differ"
			}
		} else {
			// input was point at infinity. Output should be affine order-2 point
			if !result1.IsEqual_exact(&orderTwoPoint_xtw) {
				return false, "Endo_safe(infinite point) != affine order-2 point"
			}
		}

		if s.Points[0].IsNeutralElement() != result1.IsNeutralElement_exact() {
			return false, "Endo_safe act as expected wrt neutral elements"
		}
		if !infinite { // On infinite points, AddEq(&orderTwoPoint) might not work
			var point_copy = s.Points[0].Clone().(CurvePoint)
			point_copy.AddEq(&orderTwoPoint_xtw)
			result2.Endo_safe(point_copy)
			if !result1.IsEqual_exact(result2) {
				return false, "Endo_safe(P) != Endo_safe(P+A)"
			}
		}
		return true, ""
	}
	return
}

// checks whether Endo(P) + Endo(Q) == Endo(P+Q)
func make_checkfun_endo_homomorphic(receiverType PointType) (returned_function checkfunction) {
	returned_function = func(s TestSample) (bool, string) {
		s.AssertNumberOfPoints(2)
		// This should be ruled out at the call site
		if s.AnyFlags().CheckFlag(Case_singular) {
			panic("Should not call chechfun_endo_homomorphic on NaP test samples")
		}
		if s.AnyFlags().CheckFlag(Case_differenceInfinite) {
			return true, "" // need to skip test
		}
		endo1 := MakeCurvePointFromType(receiverType)
		endo2 := MakeCurvePointFromType(receiverType)
		sum := MakeCurvePointFromType(receiverType)
		result1 := MakeCurvePointFromType(receiverType)
		result2 := MakeCurvePointFromType(receiverType)
		endo1.Endo_safe(s.Points[0])
		endo2.Endo_safe(s.Points[1])
		sum.Add(s.Points[0], s.Points[1])
		result1.Add(endo1, endo2)
		result2.Endo_safe(sum)
		if result1.IsSingular() {
			return false, "Endo(P) + Endo(Q) resulted in NaP" // cannot trigger exceptional cases of addition, because the range of Endo is the good subgroup (verified by endo_sane).
		}
		if result2.IsSingular() {
			return false, "Endo(P+Q) resulted in unexpected NaP"
		}
		if !result1.IsEqual_exact(result2) {
			return false, "Endo(P+Q) != Endo(P) + Endo(Q)"
		}
		return true, ""
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

	make_samples1_and_run_tests(t, make_checkfun_endo_sane(pointType), "Endomorphism did not pass sanity checks"+point_string, pointType, 10, excluded_flags)
	for _, type1 = range allTypes {
		make_samples1_and_run_tests(t, make_checkfun_endo_sane(pointType), "Endomorphism did not pass sanity checks"+point_string, type1, 10, excluded_flags|Case_infinite)
	}

	make_samples2_and_run_tests(t, make_checkfun_endo_homomorphic(pointType), "Endomorphism is not homomorphic"+point_string, pointType, pointType, 10, excluded_flags|Case_singular)
	for _, type1 = range allTypes {
		for _, type2 = range allTypes {
			make_samples2_and_run_tests(t, make_checkfun_endo_homomorphic(pointType), "Endomorphism is not homomorphic"+point_string, type1, type2, 10, excluded_flags|Case_singular|Case_infinite)
		}
	}
}

func TestGeneralTestsForXTW(t *testing.T) {
	test_general(t, pointTypeXTW, 0)
}

func TestGeneralTestForAXTW(t *testing.T) {
	test_general(t, pointTypeAXTW, Case_infinite)
}
