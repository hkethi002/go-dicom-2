package dicom_test

import (
	"encoding/binary"
	"reflect"
	"testing"

	"github.com/grailbio/go-dicom"
	"github.com/grailbio/go-dicom/dicomio"
	"github.com/grailbio/go-dicom/dicomtag"
	"github.com/grailbio/go-dicom/dicomuid"
)

func testWriteDataElement(t *testing.T, bo binary.ByteOrder, implicit dicomio.IsImplicitVR) {
	// Encode two scalar elements.
	e := dicomio.NewBytesEncoder(bo, implicit)
	var values []interface{}
	values = append(values, string("FooHah"))
	dicom.WriteElement(e, &dicom.Element{
		Tag:   dicomtag.Tag{0x0018, 0x9755},
		Value: values})
	values = nil
	values = append(values, uint32(1234))
	values = append(values, uint32(2345))
	dicom.WriteElement(e, &dicom.Element{
		Tag:   dicomtag.Tag{0x0020, 0x9057},
		Value: values})
	data := e.Bytes()
	// Read them back.
	d := dicomio.NewBytesDecoder(data, bo, implicit)
	elem0 := dicom.ReadElement(d, dicom.ReadOptions{})
	if d.Error() != nil {
		t.Fatal(d.Error())
	}
	tag := dicomtag.Tag{0x18, 0x9755}
	if elem0.Tag != tag {
		t.Error("Bad tag", elem0)
	}
	if len(elem0.Value) != 1 {
		t.Error("Bad value", elem0)
	}
	if elem0.Value[0].(string) != "FooHah" {
		t.Error("Bad value", elem0)
	}

	tag = dicomtag.Tag{Group: 0x20, Element: 0x9057}
	elem1 := dicom.ReadElement(d, dicom.ReadOptions{})
	if d.Error() != nil {
		t.Fatal(d.Error())
	}
	if elem1.Tag != tag {
		t.Error("Bad tag")
	}
	if len(elem1.Value) != 2 {
		t.Error("Bad value", elem1)
	}
	if elem1.Value[0].(uint32) != 1234 {
		t.Error("Bad value", elem1)
	}
	if elem1.Value[1].(uint32) != 2345 {
		t.Error("Bad value", elem1)
	}
	if err := d.Finish(); err != nil {
		t.Error(err)
	}
}

func TestWriteDataElementImplicit(t *testing.T) {
	testWriteDataElement(t, binary.LittleEndian, dicomio.ImplicitVR)
}

func TestWriteDataElementExplicit(t *testing.T) {
	testWriteDataElement(t, binary.LittleEndian, dicomio.ExplicitVR)
}

func TestWriteDataElementBigEndianExplicit(t *testing.T) {
	testWriteDataElement(t, binary.BigEndian, dicomio.ExplicitVR)
}

func TestReadWriteFileHeader(t *testing.T) {
	e := dicomio.NewBytesEncoder(binary.LittleEndian, dicomio.ImplicitVR)
	dicom.WriteFileHeader(
		e,
		[]*dicom.Element{
			dicom.MustNewElement(dicomtag.TransferSyntaxUID, dicomuid.ImplicitVRLittleEndian),
			dicom.MustNewElement(dicomtag.MediaStorageSOPClassUID, "1.2.840.10008.5.1.4.1.1.1.2"),
			dicom.MustNewElement(dicomtag.MediaStorageSOPInstanceUID, "1.2.3.4.5.6.7"),
		})
	bytes := e.Bytes()
	d := dicomio.NewBytesDecoder(bytes, binary.LittleEndian, dicomio.ImplicitVR)
	elems := dicom.ParseFileHeader(d)
	if err := d.Finish(); err != nil {
		t.Fatal(err)
	}
	elem, err := dicom.FindElementByTag(elems, dicomtag.TransferSyntaxUID)
	if err != nil {
		t.Fatal(err)
	}
	if elem.MustGetString() != dicomuid.ImplicitVRLittleEndian {
		t.Error(elem)
	}
	elem, err = dicom.FindElementByTag(elems, dicomtag.MediaStorageSOPClassUID)
	if err != nil {
		t.Fatal(err)
	}
	if elem.MustGetString() != "1.2.840.10008.5.1.4.1.1.1.2" {
		t.Error(elem)
	}
	elem, err = dicom.FindElementByTag(elems, dicomtag.MediaStorageSOPInstanceUID)
	if err != nil {
		t.Fatal(err)
	}
	if elem.MustGetString() != "1.2.3.4.5.6.7" {
		t.Error(elem)
	}
}

func TestNewElement(t *testing.T) {
	elem, err := dicom.NewElement(dicomtag.TriggerSamplePosition, uint32(10), uint32(11))
	if err != nil {
		t.Fatal(err)
	}
	if elem.Tag != dicomtag.TriggerSamplePosition || !reflect.DeepEqual(elem.MustGetUint32s(), []uint32{10, 11}) {
		t.Error(elem)
	}

	// Pass a wrong value type.
	elem, err = dicom.NewElement(dicomtag.TriggerSamplePosition, "foo")
	if err == nil {
		t.Fatal("Should fail")
	}
}
