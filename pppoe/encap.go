package pppoe

import (
	"encoding/binary"
	"fmt"
	"strings"
)

const identSpaces = "    "

// Pkt is the struct for PPPoE pkt
type Pkt struct {
	Vertype   byte
	Code      Code
	SessionID uint16
	Len       uint16
	Payload   []byte
	Tags      []Tag
}

// MaxTags is the max tags allowed in a PPPoE pkt
const MaxTags = 32

// Parse buf into pkt
func (pkt *Pkt) Parse(buf []byte) error {
	if len(buf) < 6 {
		return fmt.Errorf("invalid PPPoE packet length %d", len(buf))
	}
	pkt.Vertype = buf[0]
	if pkt.Vertype != pppoeVerType {
		return fmt.Errorf("invalid PPPoE version&type, should be 0x11, got 0x%X ", pkt.Vertype)
	}
	pkt.Code = Code(buf[1])
	pkt.SessionID = binary.BigEndian.Uint16(buf[2:4])
	pkt.Len = binary.BigEndian.Uint16(buf[4:6])
	pkt.Payload = buf[6 : 6+pkt.Len]
	pkt.Tags = []Tag{}
	if pkt.Code == CodeSession {
		// no parsing of tag for session pkt
		return nil
	}
	newFunc := func(t TagType) Tag {
		switch t {
		case TagTypeACName, TagTypeServiceName, TagTypeGenericError, TagTypeServiceNameError, TagTypeACSystemError:
			return new(TagString)
		case TagTypeEndOfList:
			return new(TagEndofList)
		}
		return new(TagByteSlice)
	}
	pos := 0
	var i int
	for i = 0; i < MaxTags; i++ {
		tag := newFunc(TagType(binary.BigEndian.Uint16(pkt.Payload[pos : pos+2])))
		n, err := tag.Parse(pkt.Payload[pos:])
		if err != nil {
			return fmt.Errorf("failed to parse PPPoE tag,%w", err)
		}
		pos += n
		pkt.Tags = append(pkt.Tags, tag)
		if pos >= len(pkt.Payload) {
			break
		}
	}
	if i == MaxTags {
		return fmt.Errorf("invalid PPPoE packe, exceed max number of tags: %d", MaxTags)
	}
	return nil
}

// Serialize pkt into bytes, without copying, and no padding
func (pkt *Pkt) Serialize() ([]byte, error) {
	if pkt.Code != CodeSession {
		pkt.Payload = []byte{}
		for _, tag := range pkt.Tags {
			buf, err := tag.Serialize()
			if err != nil {
				return nil, err
			}
			pkt.Payload = append(pkt.Payload, buf...)
		}
	}
	header := make([]byte, 6)
	header[0] = pppoeVerType
	header[1] = byte(pkt.Code)
	binary.BigEndian.PutUint16(header[2:4], pkt.SessionID)
	binary.BigEndian.PutUint16(header[4:6], uint16(len(pkt.Payload)))
	return append(header, pkt.Payload...), nil
}

// GetTag return a slice of tag with type t
func (pkt *Pkt) GetTag(t TagType) (r []Tag) {
	for _, tag := range pkt.Tags {
		if tag.Type() == uint16(t) {
			r = append(r, tag)
		}
	}
	return
}

// String returns a string representation of pkt
func (pkt Pkt) String() string {
	s := fmt.Sprintf("VerType:%x\n", pkt.Vertype)
	s += fmt.Sprintf("Code:%v\n", pkt.Code)
	s += fmt.Sprintf("SessionId:%X\n", pkt.SessionID)
	s += fmt.Sprintf("Len:%d\n", pkt.Len)
	s += "Tags:\n"
	for _, t := range pkt.Tags {
		s += fmt.Sprintf("%v%v\n", identSpaces, strings.ReplaceAll(t.String(), "\n", "\n"+identSpaces))
	}
	return s
}

// Tag is the interface for PPPoE Tag
type Tag interface {
	// Serialize Tag into byte slice
	Serialize() ([]byte, error)
	// Parse buf into Tag
	Parse(buf []byte) (int, error)
	// Type return PPPoE Tag type as uint16
	Type() uint16
	// String returns a string representation of Tag
	String() string
}

// TagEndofList is the End-of-List tag
type TagEndofList struct{}

// String implements Tag interface
func (eol TagEndofList) String() string {
	return TagTypeEndOfList.String()
}

// Type implements Tag interface
func (eol TagEndofList) Type() uint16 {
	return uint16(TagTypeEndOfList)
}

// Serialize implements Tag interface
func (eol TagEndofList) Serialize() ([]byte, error) {
	var r [4]byte
	return r[:], nil
}

// Parse implements Tag interface
func (eol TagEndofList) Parse(buf []byte) (int, error) {
	if binary.BigEndian.Uint16(buf[:2]) != 0 {
		return 0, fmt.Errorf("failed to parse %v, type is not %d", TagTypeEndOfList.String(), TagTypeEndOfList)
	}
	if binary.BigEndian.Uint16(buf[2:4]) != 0 {
		return 0, fmt.Errorf("failed to parse %v, length is not zero", TagTypeEndOfList.String())
	}
	return 4, nil
}

// TagString is for all string type of tag, like ACName,SVCName
type TagString struct {
	Value   string
	TagType TagType
}

// NewSvcTag return a new Service-Name tag
func NewSvcTag(svc string) *TagString {
	return &TagString{
		TagType: TagTypeServiceName,
		Value:   svc,
	}
}

// String implements Tag interface
func (str TagString) String() string {
	if str.Value == "" && str.TagType == TagTypeServiceName {
		return fmt.Sprintf("%v: %v", TagTypeServiceName, "<any service>")
	}
	return fmt.Sprintf("%v: %v", str.TagType, str.Value)
}

// Type implements Tag interface
func (str TagString) Type() uint16 {
	return uint16(str.TagType)
}

// Serialize implements Tag interface
func (str TagString) Serialize() ([]byte, error) {
	header := make([]byte, 4)
	binary.BigEndian.PutUint16(header[0:2], uint16(str.TagType))
	binary.BigEndian.PutUint16(header[2:4], uint16(len(str.Value)))
	return append(header, []byte(str.Value)...), nil
}

// Parse implements Tag interface
func (str *TagString) Parse(buf []byte) (int, error) {
	str.TagType = TagType(binary.BigEndian.Uint16(buf[:2]))
	str.Value = string(buf[4 : binary.BigEndian.Uint16(buf[2:4])+4])
	return 4 + len(str.Value), nil
}

// TagByteSlice is for all byte slice and unknown type of tag, e.g. ACuniq, ACCookie
type TagByteSlice struct {
	Value   []byte
	TagType TagType
}

// String implements Tag interface
func (bslice TagByteSlice) String() string {
	return fmt.Sprintf("%v: %X", bslice.TagType, bslice.Value)
}

// Type implements Tag interface
func (bslice TagByteSlice) Type() uint16 {
	return uint16(bslice.TagType)
}

// Serialize implements Tag interface
func (bslice TagByteSlice) Serialize() ([]byte, error) {
	header := make([]byte, 4)
	binary.BigEndian.PutUint16(header[0:2], uint16(bslice.TagType))
	binary.BigEndian.PutUint16(header[2:4], uint16(len(bslice.Value)))
	return append(header, bslice.Value...), nil
}

// Parse implements Tag interface
func (bslice *TagByteSlice) Parse(buf []byte) (int, error) {
	bslice.TagType = TagType(binary.BigEndian.Uint16(buf[:2]))
	bslice.Value = buf[4 : binary.BigEndian.Uint16(buf[2:4])+4]
	return 4 + len(bslice.Value), nil
}

// BBFSubTagString is for string type of BBF sub-tag
type BBFSubTagString struct {
	TagType BBFSubTagNum
	Value   string
}

// String implements Tag interface
func (str BBFSubTagString) String() string {
	return fmt.Sprintf("%v: %v", str.TagType, str.Value)
}

// Type implements Tag interface
func (str BBFSubTagString) Type() uint16 {
	return uint16(str.TagType)
}

// Serialize implements Tag interface
func (str BBFSubTagString) Serialize() ([]byte, error) {
	if len(str.Value) > 255 {
		return nil, fmt.Errorf("subtag is too long")
	}
	header := make([]byte, 2)
	header[0] = byte(str.TagType)
	header[1] = byte(len(str.Value))
	return append(header, []byte(str.Value)...), nil
}

// Parse implements Tag interface
func (str *BBFSubTagString) Parse(buf []byte) (int, error) {
	str.TagType = BBFSubTagNum(buf[0])
	str.Value = string(buf[2 : buf[1]+2])
	return 2 + len(str.Value), nil
}

// BBFSubTagUint32 is for all numberic type of BBF sub-tag
type BBFSubTagUint32 struct {
	TagType BBFSubTagNum
	Value   uint32
}

// String implements Tag interface
func (num BBFSubTagUint32) String() string {
	return fmt.Sprintf("%v: %v", num.TagType, num.Value)
}

// Type implements Tag interface
func (num BBFSubTagUint32) Type() uint16 {
	return uint16(num.TagType)
}

// Serialize implements Tag interface
func (num BBFSubTagUint32) Serialize() ([]byte, error) {
	buf := make([]byte, 6)
	buf[0] = byte(num.TagType)
	buf[1] = 4
	return buf, nil
}

// Parse implements Tag interface
func (num *BBFSubTagUint32) Parse(buf []byte) (int, error) {
	num.TagType = BBFSubTagNum(buf[0])
	num.Value = binary.BigEndian.Uint32(buf[2:6])
	return 6, nil
}

// BBFSubTagByteSlice is for all byte slice type BBF sub-tag
type BBFSubTagByteSlice struct {
	Value   []byte
	TagType BBFSubTagNum
}

// String implements Tag interface
func (bslice BBFSubTagByteSlice) String() string {
	return fmt.Sprintf("%v: %X", bslice.TagType, bslice.Value)
}

// Type implements Tag interface
func (bslice BBFSubTagByteSlice) Type() uint16 {
	return uint16(bslice.TagType)
}

// Serialize implements Tag interface
func (bslice BBFSubTagByteSlice) Serialize() ([]byte, error) {
	header := make([]byte, 2)
	if len(bslice.Value) > 255 {
		return nil, fmt.Errorf("slice is too long")
	}
	header[0] = byte(bslice.TagType)
	header[1] = byte(len(bslice.Value))
	return append(header, bslice.Value...), nil
}

// Parse implements Tag interface
func (bslice *BBFSubTagByteSlice) Parse(buf []byte) (int, error) {
	bslice.TagType = BBFSubTagNum(buf[0])
	bslice.Value = buf[2 : buf[1]+2]
	return 2 + len(bslice.Value), nil
}

// BBFTag represents a BBF PPPoE tag, which could include multiple sub-tag
type BBFTag []Tag

// Parse implements Tag interface
func (bbf *BBFTag) Parse(buf []byte) (int, error) {
	if len(buf) < 8 {
		return 0, fmt.Errorf("not enought bytes for a BBF tag")
	}
	if binary.BigEndian.Uint16(buf[:2]) != 0x105 || binary.BigEndian.Uint32(buf[4:8]) != 0xde9 {
		return 0, fmt.Errorf("invalid BBF tag")
	}
	tagLen := binary.BigEndian.Uint16(buf[2:4])
	if tagLen < 4 {
		return 0, fmt.Errorf("invalid BBF tag length")
	}
	newFunc := func(t BBFSubTagNum) Tag {
		switch t {
		case BBFSubTagNumRemoteID, BBFSubTagNumCircuitID:
			return new(BBFSubTagString)
		case BBFSubTagActualDataRateUpstream, BBFSubTagActualDataRateDownstream, BBFSubTagMinimumDataRateUpstream, BBFSubTagMinimumDataRateDownstream, BBFSubTagAttainableDataRateUpstream, BBFSubTagAttainableDataRateDownstream, BBFSubTagMaximumDataRateUpstream, BBFSubTagMaximumDataRateDownstream, BBFSubTagMinDataRateUpstreaminlow, BBFSubTagMinimumDataRateDownstreaminlow, BBFSubTagMaxInterleavingDelay, BBFSubTagActualInterleavingUpstreamDelay, BBFSubTagMaximumInterleavingDelay, BBFSubTagActualInterleavingDownstreamDelay:
			return new(BBFSubTagUint32)
		}
		return new(BBFSubTagByteSlice)
	}
	pos := 6
	var i int
	for i = 0; i < MaxTags; i++ {
		tag := newFunc(BBFSubTagNum(buf[pos]))
		n, err := tag.Parse(buf[pos:])
		if err != nil {
			return 0, fmt.Errorf("failed to parse BBF subtag,%w", err)
		}
		pos += n
		*bbf = append(*bbf, tag)
		if pos >= len(buf) {
			break
		}
	}
	if i == MaxTags {
		return 0, fmt.Errorf("invalid BBF tag, exceed max number of subtags: %d", MaxTags)
	}
	return int(tagLen) + 4, nil
}

// Serialize implements Tag interface
func (bbf BBFTag) Serialize() ([]byte, error) {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint16(buf[:2], 0x105)
	binary.BigEndian.PutUint32(buf[4:8], 0xde9)
	tagLen := 4
	for _, t := range bbf {
		newt, err := t.Serialize()
		if err != nil {
			return nil, err
		}
		tagLen += len(newt)
		buf = append(buf, newt...)
	}
	binary.BigEndian.PutUint16(buf[2:4], uint16(tagLen))
	return buf, nil
}

// Type implements Tag interface
func (bbf BBFTag) Type() uint16 {
	return uint16(TagTypeVendorSpecific)
}

// String implements Tag interface
func (bbf BBFTag) String() string {
	s := "VendorSpecific, BBF:"
	for _, t := range bbf {
		s += fmt.Sprintf("\n%v%v", identSpaces, t.String())
	}
	return s
}

// NewCircuitRemoteIDTag return a BBF Tag with circuit-id and remote-id sub tag.
// if cid or rid is empty string, then it will not be included
func NewCircuitRemoteIDTag(cid, rid string) *BBFTag {
	newFunc := func(s string, t BBFSubTagNum) *BBFSubTagString {
		r := new(BBFSubTagString)
		r.Value = s
		r.TagType = t
		return r
	}
	bbftag := &BBFTag{}
	if cid != "" {
		*bbftag = append(*bbftag, newFunc(cid, BBFSubTagNumCircuitID))
	}
	if rid != "" {
		*bbftag = append(*bbftag, newFunc(rid, BBFSubTagNumRemoteID))
	}
	return bbftag
}
