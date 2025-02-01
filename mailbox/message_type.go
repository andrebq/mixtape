package mailbox

import "github.com/google/uuid"

//go:generate msgp
//msgp:replace uuid.UUID with:[16]byte
type (
	Message struct {
		ID      uuid.UUID           `msg:"i"`
		From    Address             `msg:"f"`
		To      Address             `msg:"t"`
		Payload []byte              `msg:"p"`
		ReplyTo uuid.UUID           `msg:"rt"`
		Headers map[string][]string `msg:"h,omitempty"`
	}

	Address struct {
		Node    uuid.UUID `msg:"n"`
		Process uint64    `msg:"p"`
	}
)
