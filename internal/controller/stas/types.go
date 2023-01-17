package stas

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

	return image, nil
}

func containerImages(pod *corev1.Pod) (map[string]podContainerImage, error) {
	images := make(map[string]podContainerImage)

	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.Image != "" && containerStatus.ImageID != "" {
			image, err := newImageFromContainerStatus(containerStatus)
			if err != nil {
				return nil, err
			}

			images[containerStatus.Name] = image
		}
	}

	for _, container := range pod.Spec.Containers {
		image, ok := images[container.Name]
		if !ok {
			// We only want to add tag to images that are resolved by CRI
			continue
		}
		ref, err := reference.ParseAnyReference(container.Image)
		if err != nil {
			return nil, err
		}
		if taggedRef, ok := ref.(reference.Tagged); ok {
			image.Tag = taggedRef.Tag()
			images[container.Name] = image
		}
	}

	return images, nil
}
