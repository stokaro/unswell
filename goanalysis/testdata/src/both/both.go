package both

// unswell-disable-next-block filler.announced-importance -- Required contract wording.
// It is important to note that the client retries.
func Retry() {}

// It is important to note that the server waits. // want "filler.announced-importance"
const Message = "It is important to note that the request expires." // want "filler.announced-importance"
