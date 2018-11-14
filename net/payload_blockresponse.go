package net

import (
	"nimona.io/go/crypto"
)

// BlockResponse -
//proteus:generate
type BlockResponse struct {
	RequestID      string            `json:"requestID"`
	RequestedBlock interface{}       `json:"requestedBlock"`
	Sender         *crypto.Key       `json:"sender"`
	Signature      *crypto.Signature `json:"@sig"`
}
