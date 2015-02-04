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

func (i *BytesInput) BytesFrom(index int) []byte {
	return i.Text[index:]
}

func (i *BytesInput) BytesRange(start, end int) []byte {
	return i.Text[start:end]
}

func NewInput(input input) *Input {
	return &Input{
		input: input,
	}
}
