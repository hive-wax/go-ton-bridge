package tonbridge

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/xssnick/tonutils-go/tvm/cell"
	"math/big"
)

type Signature struct {
	V uint64
	R *big.Int
	S *big.Int
}

func generateMessageInCell(
	hash common.Hash,
	signs []*Signature,
	receiptRoot common.Hash,
	version common.Hash,
	blockNum int64,
	chainId int64,
	addr common.Address,
	topics []common.Hash,
	message []byte,
) (*cell.Cell, error) {

	signsCell := cell.BeginCell()
	for i := 0; i < len(signs); i++ {
		signsCell = signsCell.MustStoreRef(cell.BeginCell().MustStoreUInt(signs[i].V, 8).MustStoreBigUInt(signs[i].R, 256).MustStoreBigUInt(signs[i].S, 256).EndCell())
	}

	msg, err := EncodeMessage(message)
	if err != nil {
		return nil, err
	}

	return cell.BeginCell().
		MustStoreUInt(0xd5f86120, 32).
		MustStoreUInt(0, 64).
		MustStoreBigUInt(hash.Big(), 256).
		MustStoreUInt(uint64(len(signs)), 8).
		MustStoreRef(signsCell.EndCell()).
		MustStoreRef(
			cell.BeginCell().
				MustStoreBigUInt(receiptRoot.Big(), 256).
				MustStoreBigUInt(version.Big(), 256).
				MustStoreBigUInt(big.NewInt(blockNum), 256).
				MustStoreInt(chainId, 64).
				EndCell()).
		MustStoreRef(
			cell.BeginCell().
				MustStoreBigUInt(addr.Big(), 256).
				MustStoreRef(
					cell.BeginCell().
						MustStoreBigUInt(topics[0].Big(), 256).
						MustStoreBigUInt(topics[1].Big(), 256).
						MustStoreBigUInt(topics[2].Big(), 256).
						EndCell()).
				MustStoreRef(msg).
				EndCell(),
		).EndCell(), nil

}
