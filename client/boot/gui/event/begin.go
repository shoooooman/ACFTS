package event

// HandleBegin is called when receiving "begin-req" message from frontend
func HandleBegin(br chan bool) func([]byte) {
	return func(bin []byte) {
		br <- true
	}
}
