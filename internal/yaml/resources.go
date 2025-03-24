package yaml

import (
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"
)

func FromFile(filename string, obj runtime.Object) error {
	file, err := os.ReadFile(filepath.Clean(filename))
	if err != nil {
		return err
	}

	return yaml.UnmarshalStrict(file, obj)
}
