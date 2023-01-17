package controllers

import (
	"github.com/distribution/distribution/reference"
	corev1 "k8s.io/api/core/v1"

	stasv1alpha1 "github.com/statnett/image-scanner-operator/api/stas/v1alpha1"
)

type podContainerImage struct {
	stasv1alpha1.Image
	Tag string
}

func newImageFromContainerStatus(containerStatus corev1.ContainerStatus) (podContainerImage, error) {
	image := podContainerImage{}

	idRef, err := reference.ParseAnyReference(containerStatus.ImageID)
	if err != nil {
		return image, err
	}

	nameRef, err := reference.ParseAnyReference(containerStatus.Image)
	if err != nil {
		return image, err
	}

	if ref, ok := idRef.(reference.Named); ok {
		image.Name = ref.Name()
	} else if ref, ok := nameRef.(reference.Named); ok {
		image.Name = ref.Name()
	}

	if ref, ok := idRef.(reference.Digested); ok {
		image.Digest = ref.Digest()
	} else if ref, ok := nameRef.(reference.Digested); ok {
		image.Digest = ref.Digest()
	}

	if ref, ok := idRef.(reference.Tagged); ok {
		image.Tag = ref.Tag()
	} else if ref, ok := nameRef.(reference.Tagged); ok {
		image.Tag = ref.Tag()
	}

	return image, nil
}
