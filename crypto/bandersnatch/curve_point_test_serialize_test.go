package bandersnatch

import (
	"bytes"
	"strconv"
	"testing"
)

/*
func checkfun_type_consistency(s TestSample) (bool, string) {
	s.AssertNumberOfPoints(1)
	singular := s.Flags[0].CheckFlag(Case_singular)
	infinite := s.Flags[0].CheckFlag(Case_infinite)
	expect_error := singular || infinite
	var buf1, buf2 bytes.Buffer

	point_xtw := s.Points[0].ExtendedTwistedEdwards()
	bytes_written1, err1 := s.Points[0].SerializeLong(&buf1)
	bytes_written2, err2 := point_xtw.SerializeLong(&buf1)

	if expect_error {
		if err1 == nil {
			return false, "SerializeLong did not give an error even though it should"
		}
		if err2 == nil {
			return false, "Point_xw::SerializeLong did not give an error even though it should"
		}
	} else {
		if err1 != nil {
			return false, SerializeLong
		}
	}

}
*/

func checkfun_NaP_serialization(s TestSample) (bool, string) {
	s.AssertNumberOfPoints(1)
	singular := s.Flags[0].CheckFlag(Case_singular)
	infinite := s.Flags[0].CheckFlag(Case_infinite)
	expect_error := singular || infinite
	var buf bytes.Buffer
	var bytes_written int
	var err error
	var gotErrNaP, gotErrInfinity bool

	encounted_NaP_error := wasInvalidPointEncountered(func() { bytes_written, err = s.Points[0].SerializeLong(&buf) })
	gotErrInfinity = (err == ErrCannotSerializePointAtInfinity)
	gotErrNaP = (err == ErrCannotSerializeNaP)

	if bytes_written != buf.Len() {
		return false, "Number of bytes written in deserialization was reported wrongly: bytes reported = " + strconv.Itoa(bytes_written) + " Actually written = " + strconv.Itoa(buf.Len())
	}

	if encounted_NaP_error && !singular {
		return false, "NaP handler was called when calling SerializeLong on a non-NaP point"
	}
	if !encounted_NaP_error && singular {
		return false, "NaP handler was not called when calling SerializeLong on a NaP"
	}
	if expect_error {
		if err == nil {
			return false, "SerializeLong did not give an error even though it should"
		}
		if singular && infinite {
			// might actually be OK, but we bail out for now.
			panic("Error in testing framework: sample was flagged as both infinite and singular")
		}
		if singular && !gotErrNaP {
			return false, "Did not get NaP error when calling SerializeLong on NaP"
		}
		if infinite && !gotErrInfinity {
			return false, "Did not get Infinite error when calling SerializeLong on Infinite point"
		}
	} else {
		if err != nil {
			// Note: s.Points[0] might NOT be in the subgroup.
			return false, "SerializeLong gave an error even though the point was neither infinite nor a NaP"
		}
		if bytes_written != 64 {
			return false, "unexpeced number of bytes written in SerializeLong. Number written was " + strconv.Itoa(bytes_written)
		}
	}

	buf.Reset()

	encounted_NaP_error = wasInvalidPointEncountered(func() { bytes_written, err = s.Points[0].SerializeShort(&buf) })
	gotErrInfinity = (err == ErrCannotSerializePointAtInfinity)
	gotErrNaP = (err == ErrCannotSerializeNaP)

	if bytes_written != buf.Len() {
		return false, "Number of bytes written in deserialization was reported wrongly: bytes reported = " + strconv.Itoa(bytes_written) + " Actually written = " + strconv.Itoa(buf.Len())
	}

	if encounted_NaP_error && !singular {
		return false, "NaP handler was called when calling SerializeShort on a non-NaP point"
	}
	if !encounted_NaP_error && singular {
		return false, "NaP handler was not called when calling SerializeShort on a NaP"
	}
	if expect_error {
		if err == nil {
			return false, "SerializeShort did not give an error even though it should"
		}
		if singular && infinite {
			// might actually be OK, but we bail out for now.
			panic("Error in testing framework: sample was flagged as both infinite and singular")
		}
		if singular && !gotErrNaP {
			return false, "Did not get NaP error when calling SerializeShort on NaP"
		}
		if infinite && !gotErrInfinity {
			return false, "Did not get Infinite error when calling SerializeShort on Infinite point"
		}
	} else {
		if err != nil {
			// Note: s.Points[0] might NOT be in the subgroup.
			return false, "SerializeShort gave an error even though the point was neither infinite nor a NaP"
		}
		if bytes_written != 32 {
			return false, "unexpeced number of bytes written in SerializeShort Number written was " + strconv.Itoa(bytes_written)
		}
	}
	return true, ""
}

func test_serialization_properties(t *testing.T, receiverType PointType, excludedFlags PointFlags) {
	point_string := PointTypeToString(receiverType)
	// var type1, type2 PointType
	make_samples1_and_run_tests(t, checkfun_NaP_serialization, "Unexpected behaviour when serialializing wrt NaPs or infinite points "+point_string, receiverType, 10, excludedFlags)
}

func TestSerializationForXTW(t *testing.T) {
	test_serialization_properties(t, pointTypeXTW, 0)
}

func TestSerializationForAXTW(t *testing.T) {
	test_serialization_properties(t, pointTypeAXTW, 0)
}

func TestSerializationForEFGH(t *testing.T) {
	test_serialization_properties(t, pointTypeEFGH, 0)
}
