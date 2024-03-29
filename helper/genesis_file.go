package helper

import (
	"bytes"
	"embed"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/pkg/errors"
	tmTypes "github.com/tendermint/tendermint/types"
)

// nolint
var (
	//go:embed allocs
	allocs embed.FS
)

func WriteGenesisFile(chain string, filePath string, cdc *codec.Codec) (bool, error) {
	switch chain {
	case "mainnet", "donau":
		fn := fmt.Sprintf("allocs/%s.json", chain)

		genDoc, err := readPrealloc(fn, cdc)
		if err == nil {
			err = genDoc.SaveAs(filePath)
		}

		return err == nil, nil

	default:
		return false, nil
	}
}

func readPrealloc(filename string, cdc *codec.Codec) (result tmTypes.GenesisDoc, err error) {
	file, err := allocs.Open(filename)
	if err != nil {
		err = errors.Errorf("Could not open genesis preallocation for %s: %v", filename, err)

		return
	}
	defer file.Close()

	buf := bytes.NewBuffer(nil)

	_, err = buf.ReadFrom(file)
	if err == nil {
		err = cdc.UnmarshalJSON(buf.Bytes(), &result)
	}

	return
}
