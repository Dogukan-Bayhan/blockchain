package block

import "encoding/json"

// AmountResponse is returned by balance lookup endpoints.
type AmountResponse struct {
	Amount float32 `json:"amount"`
}

// MarshalJSON serializes an amount response.
func (ar *AmountResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Amount float32 `json:"amount"`
	}{
		Amount: ar.Amount,
	})
}
