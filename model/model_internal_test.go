package model

func (m *Measurement) MakeGenericTestKeysEx(
	marshal func(v interface{}) ([]byte, error),
) (map[string]interface{}, error) {
	return m.makeGenericTestKeys(marshal)
}
