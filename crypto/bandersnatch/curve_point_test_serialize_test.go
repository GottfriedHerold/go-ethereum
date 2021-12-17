package bandersnatch

import (
	"bytes"
	"strconv"
	"testing"
)

func checkfun_serialization_type_consistency(s TestSample) (bool, string) {
	s.AssertNumberOfPoints(1)
	singular := s.Flags[0].CheckFlag(Case_singular)
	infinite := s.Flags[0].CheckFlag(Case_infinite)
	if infinite || singular {
		return true, "" // converted by checkfun_NaP_serialization. No need to complicate things here
	}
	var buf1, buf2 bytes.Buffer

	point_axtw := s.Points[0].AffineExtended()
	_, err1 := s.Points[0].SerializeLong(&buf1)
	_, err2 := point_axtw.SerializeLong(&buf2)

	if err1 != nil || err2 != nil {
		return false, "Unexpected error in checkfun_type_consistency. Refer to output of checkfun_NaP_serialization"
	}

	if buf1.Len() != buf2.Len() {
		return false, "SerializeLong did not write same number of bytes depending on receiver type"
	}
	if !bytes.Equal(buf1.Bytes(), buf2.Bytes()) {
		return false, "SerializeLong did not output the same bytes depending on receiver type"
	}

	buf1.Reset()
	buf2.Reset()
	_, err1 = s.Points[0].SerializeShort(&buf1)
	_, err2 = point_axtw.SerializeShort(&buf2)

	if err1 != nil || err2 != nil {
		return false, "Unexpected error in checkfun_type_consistency. Refer to output of checkfun_NaP_serialization"
	}

	if buf1.Len() != buf2.Len() {
		return false, "SerializeShort did not write same number of bytes depending on receiver type"
	}
	if !bytes.Equal(buf1.Bytes(), buf2.Bytes()) {
		return false, "SerializeShort did not output the same bytes depending on receiver type"
	}

	return true, ""

}

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
	make_samples1_and_run_tests(t, checkfun_serialization_type_consistency, "Unexpected behaviour when comparing serialization depencency on receiver type "+point_string, receiverType, 10, excludedFlags)
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
