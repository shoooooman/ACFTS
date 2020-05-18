package event

// HandleReady is called when receiving "ready" message from frontend
func HandleReady(rdy chan bool) func([]byte) {
	return func(bin []byte) {
		rdy <- true
	}
}
