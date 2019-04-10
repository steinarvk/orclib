package cryptopacket

type Contents struct {
	// Timestamp is the timestamp of when the packet was wrapped.
	Timestamp string `json:"timestamp"`

	// Recipient is the canonical host of the recipient, to whom the message is encrypted.
	Recipient string `json:"recipient,omitempty"`

	// Sender is the canonical host of the sender.
	Sender string `json:"sender"`

	// Payload is the message contained in the packet.
	Payload interface{} `json:"payload"`
}

type Packet struct {
	// Contents is the contents of the packet; the data to be signed.
	Contents Contents `json:"contents"`

	// Signature is the signature of the canonical JSON form of Contents, with the signing key associated with Contents.Sender.
	Signature string `json:"signature"`
}
