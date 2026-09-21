package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/oenexa/oenexa/internal/core"
	"github.com/oenexa/oenexa/internal/crypto"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]

	switch command {
	case "keygen":
		handleKeygen()
	case "balance":
		if len(os.Args) < 3 {
			fmt.Println("Usage: oenexa-cli balance <address>")
			return
		}
		handleBalance(os.Args[2])
	case "transfer":
		if len(os.Args) < 5 {
			fmt.Println("Usage: oenexa-cli transfer <private_key_hex> <to_address> <amount>")
			return
		}
		handleTransfer(os.Args[2], os.Args[3], os.Args[4])
	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Oenexa Wallet CLI")
	fmt.Println("Usage:")
	fmt.Println("  oenexa-cli keygen                                        - Generate a new ML-DSA-65 wallet")
	fmt.Println("  oenexa-cli balance <address>                             - Check OEN balance of an address")
	fmt.Println("  oenexa-cli transfer <private_key> <to_address> <amount>  - Send OEN to another address")
}

func handleKeygen() {
	wallet, err := crypto.NewWallet()
	if err != nil {
		fmt.Printf("Failed to generate wallet: %v\n", err)
		return
	}

	fmt.Println("--- New Oenexa Wallet Generated ---")
	fmt.Printf("Address     : %s\n", hex.EncodeToString(wallet.Address[:]))
	fmt.Printf("Public Key  : %s\n", hex.EncodeToString(wallet.PublicKey))
	fmt.Printf("Private Key : %s\n", hex.EncodeToString(wallet.PrivateKey))
	fmt.Println("-----------------------------------")
	fmt.Println("SAVE YOUR PRIVATE KEY SAFELY. IT IS REQUIRED TO SEND FUNDS.")
}

func handleBalance(addrHex string) {
	// Call the JSON-RPC endpoint
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_getBalance",
		"params":  []interface{}{addrHex, "latest"},
		"id":      1,
	}
	body, _ := json.Marshal(payload)

	resp, err := http.Post("http://127.0.0.1:8545", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("Error connecting to node: %v\n", err)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	fmt.Printf("Node Response: %s\n", string(respBody))
}

func handleTransfer(privKeyHex, toAddrHex, amountStr string) {
	// This is a stub for the transfer command.
	// In reality, it would construct a core.Transaction, sign it with the private key,
	// and send it via eth_sendRawTransaction to the RPC.
	
	privKey, err := hex.DecodeString(privKeyHex)
	if err != nil || len(privKey) != crypto.PrivateKeySize {
		fmt.Printf("Invalid private key format or length\n")
		return
	}

	toAddr, err := hex.DecodeString(toAddrHex)
	if err != nil || len(toAddr) != crypto.AddressSize {
		fmt.Printf("Invalid to address format or length\n")
		return
	}

	var toAddress [crypto.AddressSize]byte
	copy(toAddress[:], toAddr)

	// Since we don't know the exact nonce or amount parsing in this stub, we just construct a mock Tx.
	tx := &core.Transaction{
		Version: 1,
		Type:    core.TxTransfer,
		To:      toAddress,
		Amount:  100, // Hardcoded for demo
	}
	_ = tx.Sign(privKey)

	txBytes, _ := json.Marshal(tx)
	
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_sendRawTransaction",
		"params":  []interface{}{hex.EncodeToString(txBytes)},
		"id":      1,
	}
	body, _ := json.Marshal(payload)

	resp, err := http.Post("http://127.0.0.1:8545", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("Error connecting to node: %v\n", err)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	fmt.Printf("Transaction Sent! Node Response: %s\n", string(respBody))
}
