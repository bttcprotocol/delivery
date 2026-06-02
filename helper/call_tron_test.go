package helper

import (
	"testing"

	ethTypes "github.com/ethereum/go-ethereum/core/types"
)

func TestIsTronTransactionReceiptSuccessful(t *testing.T) {
	testCases := []struct {
		name    string
		receipt *ethTypes.Receipt
		want    bool
	}{
		{
			name: "successful receipt",
			receipt: &ethTypes.Receipt{
				Status: ethTypes.ReceiptStatusSuccessful,
			},
			want: true,
		},
		{
			name: "failed receipt",
			receipt: &ethTypes.Receipt{
				Status: ethTypes.ReceiptStatusFailed,
			},
		},
		{
			name: "nil receipt",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsTronTransactionReceiptSuccessful(tc.receipt)
			if got != tc.want {
				t.Fatalf("expected %t, got %t", tc.want, got)
			}
		})
	}
}
