package broker

type singleBuffer struct {
	buffer *Message
}

func (sb *singleBuffer) put(msg *Message) {
	sb.buffer = msg
}

func (sb *singleBuffer) get() *Message {
	return sb.buffer
}

func (sb *singleBuffer) emit(ch chan *Message) {
	ch <- sb.buffer
}
