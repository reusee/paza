package paza

type BytesInput struct {
	Text []byte
}

func NewBytesInput(text []byte) *Input {
	return &Input{
		input: &BytesInput{
			Text: text,
		},
	}
}

func (i *BytesInput) Len() int {
	return len(i.Text)
}

func (i *BytesInput) At(index int) []byte {
	return []byte{i.Text[index]}
}

func (i *BytesInput) BytesSlice(start, end int) []byte {
	if start < 0 && end < 0 {
		return i.Text[:]
	} else if start < 0 {
		return i.Text[:end]
	} else if end < 0 {
		return i.Text[start:]
	}
	return i.Text[start:end]
}

func NewInput(input input) *Input {
	return &Input{
		input: input,
	}
}
