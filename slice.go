// Auto-generated by genslice.go, DON'T EDIT IT!
package log

func (recorder *Recorder) Ints(key string, value []int) *Recorder {
	if recorder != nil {
		recorder.encoder.encodeKey(key)
		recorder.encoder.encodeInts(value)
	}
	return recorder
}

func (recorder *Recorder) Int8s(key string, value []int8) *Recorder {
	if recorder != nil {
		recorder.encoder.encodeKey(key)
		recorder.encoder.encodeInt8s(value)
	}
	return recorder
}

func (recorder *Recorder) Int16s(key string, value []int16) *Recorder {
	if recorder != nil {
		recorder.encoder.encodeKey(key)
		recorder.encoder.encodeInt16s(value)
	}
	return recorder
}

func (recorder *Recorder) Int32s(key string, value []int32) *Recorder {
	if recorder != nil {
		recorder.encoder.encodeKey(key)
		recorder.encoder.encodeInt32s(value)
	}
	return recorder
}

func (recorder *Recorder) Int64s(key string, value []int64) *Recorder {
	if recorder != nil {
		recorder.encoder.encodeKey(key)
		recorder.encoder.encodeInt64s(value)
	}
	return recorder
}

func (recorder *Recorder) Uints(key string, value []uint) *Recorder {
	if recorder != nil {
		recorder.encoder.encodeKey(key)
		recorder.encoder.encodeUints(value)
	}
	return recorder
}

func (recorder *Recorder) Uint8s(key string, value []uint8) *Recorder {
	if recorder != nil {
		recorder.encoder.encodeKey(key)
		recorder.encoder.encodeUint8s(value)
	}
	return recorder
}

func (recorder *Recorder) Uint16s(key string, value []uint16) *Recorder {
	if recorder != nil {
		recorder.encoder.encodeKey(key)
		recorder.encoder.encodeUint16s(value)
	}
	return recorder
}

func (recorder *Recorder) Uint32s(key string, value []uint32) *Recorder {
	if recorder != nil {
		recorder.encoder.encodeKey(key)
		recorder.encoder.encodeUint32s(value)
	}
	return recorder
}

func (recorder *Recorder) Uint64s(key string, value []uint64) *Recorder {
	if recorder != nil {
		recorder.encoder.encodeKey(key)
		recorder.encoder.encodeUint64s(value)
	}
	return recorder
}

func (recorder *Recorder) Float32s(key string, value []float32) *Recorder {
	if recorder != nil {
		recorder.encoder.encodeKey(key)
		recorder.encoder.encodeFloat32s(value)
	}
	return recorder
}

func (recorder *Recorder) Float64s(key string, value []float64) *Recorder {
	if recorder != nil {
		recorder.encoder.encodeKey(key)
		recorder.encoder.encodeFloat64s(value)
	}
	return recorder
}

func (recorder *Recorder) Complex64s(key string, value []complex64) *Recorder {
	if recorder != nil {
		recorder.encoder.encodeKey(key)
		recorder.encoder.encodeComplex64s(value)
	}
	return recorder
}

func (recorder *Recorder) Complex128s(key string, value []complex128) *Recorder {
	if recorder != nil {
		recorder.encoder.encodeKey(key)
		recorder.encoder.encodeComplex128s(value)
	}
	return recorder
}

func (recorder *Recorder) Bools(key string, value []bool) *Recorder {
	if recorder != nil {
		recorder.encoder.encodeKey(key)
		recorder.encoder.encodeBools(value)
	}
	return recorder
}

func (recorder *Recorder) Strings(key string, value []string) *Recorder {
	if recorder != nil {
		recorder.encoder.encodeKey(key)
		recorder.encoder.encodeStrings(value)
	}
	return recorder
}

func (recorder *Recorder) Bytes(key string, value []byte) *Recorder {
	if recorder != nil {
		recorder.encoder.encodeKey(key)
		recorder.encoder.encodeBytes(value)
	}
	return recorder
}

func (enc *encoder) encodeInts(s []int) {
	enc.writeByte('[')
	for i := range s {
		if i > 0 {
			enc.writeByte(',')
		}
		enc.encodeInt(int64(s[i]))
	}
	enc.writeByte(']')
}

func (enc *encoder) encodeInt8s(s []int8) {
	enc.writeByte('[')
	for i := range s {
		if i > 0 {
			enc.writeByte(',')
		}
		enc.encodeInt(int64(s[i]))
	}
	enc.writeByte(']')
}

func (enc *encoder) encodeInt16s(s []int16) {
	enc.writeByte('[')
	for i := range s {
		if i > 0 {
			enc.writeByte(',')
		}
		enc.encodeInt(int64(s[i]))
	}
	enc.writeByte(']')
}

func (enc *encoder) encodeInt32s(s []int32) {
	enc.writeByte('[')
	for i := range s {
		if i > 0 {
			enc.writeByte(',')
		}
		enc.encodeInt(int64(s[i]))
	}
	enc.writeByte(']')
}

func (enc *encoder) encodeInt64s(s []int64) {
	enc.writeByte('[')
	for i := range s {
		if i > 0 {
			enc.writeByte(',')
		}
		enc.encodeInt(int64(s[i]))
	}
	enc.writeByte(']')
}

func (enc *encoder) encodeUints(s []uint) {
	enc.writeByte('[')
	for i := range s {
		if i > 0 {
			enc.writeByte(',')
		}
		enc.encodeUint(uint64(s[i]))
	}
	enc.writeByte(']')
}

func (enc *encoder) encodeUint8s(s []uint8) {
	enc.writeByte('[')
	for i := range s {
		if i > 0 {
			enc.writeByte(',')
		}
		enc.encodeUint(uint64(s[i]))
	}
	enc.writeByte(']')
}

func (enc *encoder) encodeUint16s(s []uint16) {
	enc.writeByte('[')
	for i := range s {
		if i > 0 {
			enc.writeByte(',')
		}
		enc.encodeUint(uint64(s[i]))
	}
	enc.writeByte(']')
}

func (enc *encoder) encodeUint32s(s []uint32) {
	enc.writeByte('[')
	for i := range s {
		if i > 0 {
			enc.writeByte(',')
		}
		enc.encodeUint(uint64(s[i]))
	}
	enc.writeByte(']')
}

func (enc *encoder) encodeUint64s(s []uint64) {
	enc.writeByte('[')
	for i := range s {
		if i > 0 {
			enc.writeByte(',')
		}
		enc.encodeUint(uint64(s[i]))
	}
	enc.writeByte(']')
}

func (enc *encoder) encodeFloat32s(s []float32) {
	enc.writeByte('[')
	for i := range s {
		if i > 0 {
			enc.writeByte(',')
		}
		enc.encodeFloat(float64(s[i]), 32)
	}
	enc.writeByte(']')
}

func (enc *encoder) encodeFloat64s(s []float64) {
	enc.writeByte('[')
	for i := range s {
		if i > 0 {
			enc.writeByte(',')
		}
		enc.encodeFloat(float64(s[i]), 64)
	}
	enc.writeByte(']')
}

func (enc *encoder) encodeComplex64s(s []complex64) {
	enc.writeByte('[')
	for i := range s {
		if i > 0 {
			enc.writeByte(',')
		}
		enc.encodeComplex(float64(real(s[i])), float64(imag(s[i])), 32)
	}
	enc.writeByte(']')
}

func (enc *encoder) encodeComplex128s(s []complex128) {
	enc.writeByte('[')
	for i := range s {
		if i > 0 {
			enc.writeByte(',')
		}
		enc.encodeComplex(float64(real(s[i])), float64(imag(s[i])), 64)
	}
	enc.writeByte(']')
}

func (enc *encoder) encodeBools(s []bool) {
	enc.writeByte('[')
	for i := range s {
		if i > 0 {
			enc.writeByte(',')
		}
		enc.encodeBool(s[i])
	}
	enc.writeByte(']')
}

func (enc *encoder) encodeStrings(s []string) {
	enc.writeByte('[')
	for i := range s {
		if i > 0 {
			enc.writeByte(',')
		}
		enc.encodeString(s[i])
	}
	enc.writeByte(']')
}

func (enc *encoder) encodeBytes(s []byte) {
	enc.writeString("0x")
	for i := range s {
		h, l := s[i]>>4, s[i]&0xF
		enc.writeByte(hex[h])
		enc.writeByte(hex[l])
	}
}
