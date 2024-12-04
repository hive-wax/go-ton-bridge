package tonbridge

import (
	"encoding/hex"
	"fmt"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

func EncodeMessage(hexData string) (*cell.Cell, error) {
	if len(hexData) < 2 || hexData[:2] != "0x" {
		return nil, fmt.Errorf("hex data must start with 0x")
	}

	data, err := hex.DecodeString(hexData[2:])
	if err != nil {
		return nil, fmt.Errorf("invalid hex data: %w", err)
	}

	if len(data) < 64 {
		return nil, fmt.Errorf("data too short, minimum 64 bytes required")
	}

	// Extract metadata and data positions
	offset := data[:32]
	length := data[32:64]
	pos := 64

	// Build metadata cell
	metadataCell := cell.BeginCell().
		MustStoreSlice(offset, uint(len(offset))*8).
		MustStoreSlice(length, uint(len(length))*8).
		EndCell()

	// Read header fields
	version := data[pos]
	pos++
	relay := data[pos]
	pos++
	tokenLen := data[pos]
	pos++
	mosLen := data[pos]
	pos++
	fromLen := data[pos]
	pos++
	toLen := data[pos]
	pos++
	payloadLen := uint16(data[pos])<<8 | uint16(data[pos+1])
	pos += 2

	// Read reserved and token amount
	reserved := data[pos : pos+8]
	pos += 8
	tokenAmount := data[pos : pos+16]
	pos += 16

	// Read addresses
	tokenAddr := data[pos : pos+int(tokenLen)]
	pos += int(tokenLen)
	mosTarget := data[pos : pos+int(mosLen)]
	pos += int(mosLen)
	fromAddr := data[pos : pos+int(fromLen)]
	pos += int(fromLen)
	toAddr := data[pos : pos+int(toLen)]
	pos += int(toLen)
	payload := data[pos : pos+int(payloadLen)]

	// Build header cell
	headerCell := cell.BeginCell().
		MustStoreUInt(uint64(version), 8).
		MustStoreUInt(uint64(relay), 8).
		MustStoreUInt(uint64(tokenLen), 8).
		MustStoreUInt(uint64(mosLen), 8).
		MustStoreUInt(uint64(fromLen), 8).
		MustStoreUInt(uint64(toLen), 8).
		MustStoreUInt(uint64(payloadLen), 16).
		MustStoreSlice(reserved, uint(len(reserved))*8).
		MustStoreSlice(tokenAmount, uint(len(tokenAmount))*8).
		EndCell()

	// Build addresses cells
	tokenMosCell := cell.BeginCell().
		MustStoreSlice(tokenAddr, uint(len(tokenAddr))*8).
		MustStoreSlice(mosTarget, uint(len(mosTarget))*8).
		EndCell()

	fromToCell := cell.BeginCell().
		MustStoreSlice(fromAddr, uint(len(fromAddr))*8).
		MustStoreSlice(toAddr, uint(len(toAddr))*8).
		EndCell()

	// Build payload cell
	payloadCell := cell.BeginCell().
		MustStoreSlice(payload, uint(len(payload))*8).
		EndCell()

	// Link all cells together
	metadataAndHeader := cell.BeginCell().
		MustStoreRef(metadataCell).
		MustStoreRef(headerCell).
		EndCell()

	return cell.BeginCell().
		MustStoreRef(metadataAndHeader).
		MustStoreRef(tokenMosCell).
		MustStoreRef(fromToCell).
		MustStoreRef(payloadCell).
		EndCell(), nil
}
