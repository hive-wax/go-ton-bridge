package tonbridge

import (
	"fmt"
	"math"
	"strings"

	"github.com/xssnick/tonutils-go/tvm/cell"
)

const (
	maxBits         = 1016
	maxBytesPerCell = maxBits / 8 // ~127 bytes
)

func bytesToBinary(data []byte) string {
	var binary strings.Builder
	for _, b := range data {
		binary.WriteString(fmt.Sprintf("%08b", b))
	}
	return binary.String()
}

func binaryToBytes(binary string) []byte {
	// Pad binary string to multiple of 8
	padding := len(binary) % 8
	if padding != 0 {
		binary = binary + strings.Repeat("0", 8-padding)
	}

	result := make([]byte, len(binary)/8)
	for i := 0; i < len(binary); i += 8 {
		end := i + 8
		if end > len(binary) {
			end = len(binary)
		}
		chunk := binary[i:end]

		var val byte
		for j, bit := range chunk {
			if bit == '1' {
				val |= 1 << (7 - j)
			}
		}
		result[i/8] = val
	}
	return result
}

func EncodePayload(data []byte) (*cell.Cell, error) {
	binaryStr := bytesToBinary(data)

	var cells []*cell.Builder
	position := 0

	for position < len(binaryStr) {
		builder := cell.BeginCell()
		end := position + maxBits
		if end > len(binaryStr) {
			end = len(binaryStr)
		}
		bitsForCurrentCell := binaryStr[position:end]

		for _, bit := range bitsForCurrentCell {
			if bit == '1' {
				builder.StoreBoolBit(true)
			} else {
				builder.StoreBoolBit(false)
			}
		}

		cells = append(cells, builder)
		position = end
	}

	lastCell := cells[len(cells)-1].EndCell()
	for i := len(cells) - 2; i >= 0; i-- {
		cells[i].StoreRef(lastCell)
		lastCell = cells[i].EndCell()
	}

	return lastCell, nil
}

func DecodePayload(rootCell *cell.Cell) ([]byte, error) {
	var binaryResult strings.Builder
	currentCell := rootCell

	for {
		slice := currentCell.BeginParse()

		for slice.BitsLeft() > 0 {
			bit, err := slice.LoadBoolBit()
			if err != nil {
				return nil, fmt.Errorf("error loading bit: %w", err)
			}
			if bit {
				binaryResult.WriteString("1")
			} else {
				binaryResult.WriteString("0")
			}
		}

		if slice.RefsNum() == 0 {
			break
		}

		nextCell, err := slice.LoadRefCell()
		if err != nil {
			return nil, fmt.Errorf("error loading next cell: %w", err)
		}
		currentCell = nextCell
	}

	return binaryToBytes(binaryResult.String()), nil
}

func CalculateCellCount(data []byte) int {
	return int(math.Ceil(float64(len(data)) / float64(maxBytesPerCell)))
}
