package parsers

import (
	"os"
	"path/filepath"
)

type KubernetesVolumeParser struct {
	p string
}

func NewKubernetesVolumeParser(path string) *KubernetesVolumeParser {
	return &KubernetesVolumeParser{path}
}

func (p *KubernetesVolumeParser) Parse() (map[string]interface{}, error) {
	values := make(map[string]interface{})

	if err := filepath.Walk(p.p, func(path string, fi os.FileInfo, _ error) error {
		if path == p.p {
			return nil
		}

		key := filepath.Base(path)
		b, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		values[key] = string(b)

		return nil
	}); err != nil {
		return nil, err
	}

	return values, nil
}
