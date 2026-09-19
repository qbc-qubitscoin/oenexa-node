package main

import "github.com/oenexa/oenexa/internal/contracts/multisig"

//export init_multisig
func init_multisig(payloadPtr, payloadLen uint32) int32 {
	return multisig.InitMultisig(payloadPtr, payloadLen)
}

//export execute_transfer
func execute_transfer(payloadPtr, payloadLen uint32) int32 {
	return multisig.ExecuteTransfer(payloadPtr, payloadLen)
}

func main() {}
