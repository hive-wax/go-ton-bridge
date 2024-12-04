package tonbridge

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/xssnick/tonutils-go/tvm/cell"
)

const (
	maxBits     = 1016
	maxHexChars = 1016 / 4
)

func hexToBinary(hexStr string) string {
	var binary strings.Builder
	for _, ch := range hexStr {
		val := 0
		if ch >= '0' && ch <= '9' {
			val = int(ch - '0')
		} else {
			val = int(ch-'a') + 10
		}
		binary.WriteString(strings.Repeat("0", 4-len(fmt.Sprintf("%b", val))))
		binary.WriteString(fmt.Sprintf("%b", val))
	}
	return binary.String()
}

func binaryToHex(binary string) string {
	var hex strings.Builder
	for i := 0; i < len(binary); i += 4 {
		end := i + 4
		if end > len(binary) {
			end = len(binary)
		}
		chunk := binary[i:end]
		val := 0
		for j, bit := range chunk {
			if bit == '1' {
				val |= 1 << (3 - j)
			}
		}
		hex.WriteString(fmt.Sprintf("%x", val))
	}
	return hex.String()
}

func EncodePayload(hexData string) (*cell.Cell, error) {
	if !strings.HasPrefix(hexData, "0x") {
		return nil, errors.New("hex data must start with 0x")
	}

	cleanHex := strings.ToLower(hexData[2:])
	binaryStr := hexToBinary(cleanHex)

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

func DecodePayload(rootCell *cell.Cell) (string, error) {
	var binaryResult strings.Builder
	currentCell := rootCell

	for {
		slice := currentCell.BeginParse()

		for slice.BitsLeft() > 0 {
			bit, err := slice.LoadBoolBit()
			if err != nil {
				return "", fmt.Errorf("error loading bit: %w", err)
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
			return "", fmt.Errorf("error loading next cell: %w", err)
		}
		currentCell = nextCell
	}

	return "0x" + binaryToHex(binaryResult.String()), nil
}

func CalculateCellCount(hexData string) int {
	cleanHex := hexData
	if strings.HasPrefix(hexData, "0x") {
		cleanHex = hexData[2:]
	}
	return int(math.Ceil(float64(len(cleanHex)) / 254))
}
