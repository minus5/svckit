package broker

type singleBuffer struct {
	buffer []byte
}

func (sb *singleBuffer) put(msg []byte) {
	sb.buffer = msg
}

func (sb *singleBuffer) get() []byte {
	return sb.buffer
}

func (sb *singleBuffer) emit(ch chan []byte) {
	ch <- sb.buffer
}
