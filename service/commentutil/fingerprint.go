package commentutil

import (
	"fmt"
	"hash/fnv"

	"github.com/reviewdog/reviewdog/proto/rdf"
	"google.golang.org/protobuf/proto"
)

func Fingerprint(d *rdf.Diagnostic) (string, error) {
	h := fnv.New64a()
	// Ideally, we should not use proto.Marshal since Proto Serialization Is Not
	// Canonical.
	// https://protobuf.dev/programming-guides/serialization-not-canonical/
	//
	// However, I left it as-is for now considering the same reviewdog binary
	// should re-calculate and compare fingerprint for almost all cases.
	data, err := proto.Marshal(d)
	if err != nil {
		return "", err
	}
	if _, err := h.Write(data); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum64()), nil
}
