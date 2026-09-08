package literals

// It is important to note that this is an external contract.
const Message = "It is \u0069mportant to note that the client retries." // want "filler.announced-importance"

const Raw = `It is important to note that the server waits.` // want "filler.announced-importance"
