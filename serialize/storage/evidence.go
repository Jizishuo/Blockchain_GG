package storage

import (
	"Blockchain_GG/serialize/cp"
	"io"
)

// 证据
type Evidence struct {
	*cp.Evidence
}

func UnmaishalEvidence(data io.Reader) (*Evidence, error) {
	result := &Evidence{}
	var err error
	if result.Evidence, err = cp.u
}