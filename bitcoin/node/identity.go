package node

import "github.com/google/uuid"

// GenerateID creates a random UUID v4 string for a node identity.
func GenerateID() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}

	return id.String(), nil
}
