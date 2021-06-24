// Auto-generated by gen.go, DON'T EDIT IT!
package log

func (fields *Fields) Ints(key string, value []int) *Fields {
	if fields != nil {
		fields.encoder.encodeKey(key)
		fields.encoder.encodeInts(value)
	}
	return fields
}

func (fields *Fields) Int8s(key string, value []int8) *Fields {
	if fields != nil {
		fields.encoder.encodeKey(key)
		fields.encoder.encodeInt8s(value)
	}
	return fields
}

func (fields *Fields) Int16s(key string, value []int16) *Fields {
	if fields != nil {
		fields.encoder.encodeKey(key)
		fields.encoder.encodeInt16s(value)
	}
	return fields
}

func (fields *Fields) Int32s(key string, value []int32) *Fields {
	if fields != nil {
		fields.encoder.encodeKey(key)
		fields.encoder.encodeInt32s(value)
	}
	return fields
}

func (fields *Fields) Int64s(key string, value []int64) *Fields {
	if fields != nil {
		fields.encoder.encodeKey(key)
		fields.encoder.encodeInt64s(value)
	}
	return fields
}

func (fields *Fields) Uints(key string, value []uint) *Fields {
	if fields != nil {
		fields.encoder.encodeKey(key)
		fields.encoder.encodeUints(value)
	}
	return fields
}

func (fields *Fields) Uint8s(key string, value []uint8) *Fields {
	if fields != nil {
		fields.encoder.encodeKey(key)
		fields.encoder.encodeUint8s(value)
	}
	return fields
}

func (fields *Fields) Uint16s(key string, value []uint16) *Fields {
	if fields != nil {
		fields.encoder.encodeKey(key)
		fields.encoder.encodeUint16s(value)
	}
	return fields
}

func (fields *Fields) Uint32s(key string, value []uint32) *Fields {
	if fields != nil {
		fields.encoder.encodeKey(key)
		fields.encoder.encodeUint32s(value)
	}
	return fields
}

func (fields *Fields) Uint64s(key string, value []uint64) *Fields {
	if fields != nil {
		fields.encoder.encodeKey(key)
		fields.encoder.encodeUint64s(value)
	}
	return fields
}

func (fields *Fields) Float32s(key string, value []float32) *Fields {
	if fields != nil {
		fields.encoder.encodeKey(key)
		fields.encoder.encodeFloat32s(value)
	}
	return fields
}

func (fields *Fields) Float64s(key string, value []float64) *Fields {
	if fields != nil {
		fields.encoder.encodeKey(key)
		fields.encoder.encodeFloat64s(value)
	}
	return fields
}

func (fields *Fields) Complex64s(key string, value []complex64) *Fields {
	if fields != nil {
		fields.encoder.encodeKey(key)
		fields.encoder.encodeComplex64s(value)
	}
	return fields
}

func (fields *Fields) Complex128s(key string, value []complex128) *Fields {
	if fields != nil {
		fields.encoder.encodeKey(key)
		fields.encoder.encodeComplex128s(value)
	}
	return fields
}

func (fields *Fields) Bools(key string, value []bool) *Fields {
	if fields != nil {
		fields.encoder.encodeKey(key)
		fields.encoder.encodeBools(value)
	}
	return fields
}

func (fields *Fields) Strings(key string, value []string) *Fields {
	if fields != nil {
		fields.encoder.encodeKey(key)
		fields.encoder.encodeStrings(value)
	}
	return fields
}

func (fields *Fields) Bytes(key string, value []byte) *Fields {
	if fields != nil {
		fields.encoder.encodeKey(key)
		fields.encoder.encodeBytes(value)
	}
	return fields
}

func (b *jsonx) encodeInts(s []int) {
	b.writeByte('[')
	for i := range s {
		if i > 0 {
			b.writeByte(',')
		}
		b.encodeInt(int64(s[i]))
	}
	b.writeByte(']')
}

func (b *jsonx) encodeInt8s(s []int8) {
	b.writeByte('[')
	for i := range s {
		if i > 0 {
			b.writeByte(',')
		}
		b.encodeInt(int64(s[i]))
	}
	b.writeByte(']')
}

func (b *jsonx) encodeInt16s(s []int16) {
	b.writeByte('[')
	for i := range s {
		if i > 0 {
			b.writeByte(',')
		}
		b.encodeInt(int64(s[i]))
	}
	b.writeByte(']')
}

func (b *jsonx) encodeInt32s(s []int32) {
	b.writeByte('[')
	for i := range s {
		if i > 0 {
			b.writeByte(',')
		}
		b.encodeInt(int64(s[i]))
	}
	b.writeByte(']')
}

func (b *jsonx) encodeInt64s(s []int64) {
	b.writeByte('[')
	for i := range s {
		if i > 0 {
			b.writeByte(',')
		}
		b.encodeInt(int64(s[i]))
	}
	b.writeByte(']')
}

func (b *jsonx) encodeUints(s []uint) {
	b.writeByte('[')
	for i := range s {
		if i > 0 {
			b.writeByte(',')
		}
		b.encodeUint(uint64(s[i]))
	}
	b.writeByte(']')
}

func (b *jsonx) encodeUint8s(s []uint8) {
	b.writeByte('[')
	for i := range s {
		if i > 0 {
			b.writeByte(',')
		}
		b.encodeUint(uint64(s[i]))
	}
	b.writeByte(']')
}

func (b *jsonx) encodeUint16s(s []uint16) {
	b.writeByte('[')
	for i := range s {
		if i > 0 {
			b.writeByte(',')
		}
		b.encodeUint(uint64(s[i]))
	}
	b.writeByte(']')
}

func (b *jsonx) encodeUint32s(s []uint32) {
	b.writeByte('[')
	for i := range s {
		if i > 0 {
			b.writeByte(',')
		}
		b.encodeUint(uint64(s[i]))
	}
	b.writeByte(']')
}

func (b *jsonx) encodeUint64s(s []uint64) {
	b.writeByte('[')
	for i := range s {
		if i > 0 {
			b.writeByte(',')
		}
		b.encodeUint(uint64(s[i]))
	}
	b.writeByte(']')
}

func (b *jsonx) encodeFloat32s(s []float32) {
	b.writeByte('[')
	for i := range s {
		if i > 0 {
			b.writeByte(',')
		}
		b.encodeFloat(float64(s[i]), 32)
	}
	b.writeByte(']')
}

func (b *jsonx) encodeFloat64s(s []float64) {
	b.writeByte('[')
	for i := range s {
		if i > 0 {
			b.writeByte(',')
		}
		b.encodeFloat(float64(s[i]), 64)
	}
	b.writeByte(']')
}

func (b *jsonx) encodeComplex64s(s []complex64) {
	b.writeByte('[')
	for i := range s {
		if i > 0 {
			b.writeByte(',')
		}
		b.encodeComplex(float64(real(s[i])), float64(imag(s[i])), 32)
	}
	b.writeByte(']')
}

func (b *jsonx) encodeComplex128s(s []complex128) {
	b.writeByte('[')
	for i := range s {
		if i > 0 {
			b.writeByte(',')
		}
		b.encodeComplex(float64(real(s[i])), float64(imag(s[i])), 64)
	}
	b.writeByte(']')
}

func (b *jsonx) encodeBools(s []bool) {
	b.writeByte('[')
	for i := range s {
		if i > 0 {
			b.writeByte(',')
		}
		b.encodeBool(s[i])
	}
	b.writeByte(']')
}

func (b *jsonx) encodeStrings(s []string) {
	b.writeByte('[')
	for i := range s {
		if i > 0 {
			b.writeByte(',')
		}
		b.encodeString(s[i])
	}
	b.writeByte(']')
}

func (b *jsonx) encodeBytes(s []byte) {
	b.writeString("0x")
	for i := range s {
		h, l := s[i]>>4, s[i]&0xF
		b.writeByte(hex[h])
		b.writeByte(hex[l])
	}
}