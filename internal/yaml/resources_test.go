package yaml

import (
	"path"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestFromFile(t *testing.T) {
	type args struct {
		filename string
		obj      runtime.Object
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "good resource", args: args{filename: path.Join("test", "pod.yaml"), obj: &corev1.Pod{}}},
		{name: "invalid resource", args: args{filename: path.Join("test", "invalid-pod.yaml"), obj: &corev1.Pod{}}, wantErr: true},
		{name: "missing file", args: args{filename: "does-not-exist.yaml"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := FromFile(tt.args.filename, tt.args.obj); (err != nil) != tt.wantErr {
				t.Errorf("FromFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
