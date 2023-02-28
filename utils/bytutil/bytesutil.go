package bytutil

import "errors"

const mask = 0xff

func ToInt32(buf []byte) (int32, error) {
	if len(buf) != 4 {
		return 0, errors.New("int32 need 4 bytutil")
	}
	ret := int32(buf[0])

	ret |= int32(buf[1]) << 8
	ret |= int32(buf[2]) << 16
	ret |= int32(buf[3]) << 24
	return ret, nil
}

func ToUInt64(buf []byte) (uint64, error) {
	if len(buf) != 8 {
		return 0, errors.New("int32 need 4 bytutil")
	}
	ret := uint64(buf[0])

	ret |= uint64(buf[1]) << 8
	ret |= uint64(buf[2]) << 16
	ret |= uint64(buf[3]) << 24
	ret |= uint64(buf[4]) << 32
	ret |= uint64(buf[5]) << 40
	ret |= uint64(buf[6]) << 48
	ret |= uint64(buf[7]) << 56
	return ret, nil
}

func ToBytes(v int32) []byte {
	buf := make([]byte, 4, 4)
	uv := uint32(v)
	buf[0] = byte(uv & mask)
	buf[1] = byte(uv >> 8 & mask)
	buf[2] = byte(uv >> 16 & mask)
	buf[3] = byte(uv >> 24 & mask)

	return buf
}

func Uint64ToBytes(uv uint64) []byte {
	buf := make([]byte, 8)

	buf[0] = byte(uv & mask)
	buf[1] = byte(uv >> 8 & mask)
	buf[2] = byte(uv >> 16 & mask)
	buf[3] = byte(uv >> 24 & mask)
	buf[4] = byte(uv >> 32 & mask)
	buf[5] = byte(uv >> 40 & mask)
	buf[6] = byte(uv >> 48 & mask)
	buf[7] = byte(uv >> 56 & mask)

	return buf
}
