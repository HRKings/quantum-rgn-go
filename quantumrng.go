package quantumrng

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
)

var (
	CacheSize       = 1024
	CacheLimit      = CacheSize * 1024
	Cache           = make([]byte, CacheLimit)
	CacheStartIndex = -1
)

type AnuResponse struct {
	Type    string   `json:"type"`
	Length  int      `json:"length"`
	Size    int      `json:"size"`
	Data    []string `json:"data"`
	Success bool     `json:"success"`
}

func FetchQuantumRandomHex(quantity int) ([]string, error) {
	URL := fmt.Sprintf("https://qrng.anu.edu.au/API/jsonI.php?length=%d&type=hex16&size=1024", quantity)

	resp, err := http.Get(URL)
	if err != nil {
		return nil, err
	}

	var response AnuResponse

	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}

	return response.Data, nil
}

func RefreshCache() error {
	hexes, err := FetchQuantumRandomHex(CacheSize)
	if err != nil {
		return err
	}

	for i := 0; i < CacheSize; i++ {
		hexByte, err := hex.DecodeString(hexes[i])
		if err != nil {
			return err
		}

		for j := 0; j < 1024; j++ {
			Cache[i*1024+j] = hexByte[j]
		}
	}

	CacheStartIndex = 0
	return nil
}

func GetBytesFromCache(quantity int) ([]byte, error) {
	if CacheStartIndex >= CacheLimit ||
		CacheStartIndex == -1 ||
		CacheStartIndex+quantity > CacheLimit-1 {
		err := RefreshCache()
		if err != nil {
			return nil, err
		}
	}

	result := Cache[CacheStartIndex : CacheStartIndex+quantity]
	CacheStartIndex += quantity

	return result, nil
}

func ApproximatePowerOf2(input int) int {
	input = input - 1
	input |= input >> 1
	input |= input >> 2
	input |= input >> 4
	input |= input >> 8
	input |= input >> 16

	return input + 1
}

func GetRandomInt(min int, max int) (int, error) {
	byteQuantity := int(math.Max(math.Log2(float64(ApproximatePowerOf2(max)))/8, 1))
	hexByte, err := GetBytesFromCache(byteQuantity)
	if err != nil {
		return 0, err
	}

	complete := make([]byte, 8-byteQuantity)
	complete = append(complete, hexByte...)
	numberForm := binary.BigEndian.Uint64(complete)

	return (int(numberForm) % (max - min + 1)) + min, nil
}
