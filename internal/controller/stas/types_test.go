package stas

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"

	stasv1alpha1 "github.com/statnett/image-scanner-operator/api/stas/v1alpha1"
)

var _ = Describe("ImageReference", func() {
	DescribeTable("Creating from pod",
		func(pod *corev1.Pod, expectedImage podContainerImage) {
			images, err := containerImages(pod)
			Expect(err).To(Succeed())
			Expect(images).To(HaveLen(1))

			for _, image := range images {
				Expect(*image).To(Equal(expectedImage))
			}
		},
		Entry("Standard FQ image",
			&corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "standard-fq",
							Image: "my.registry/repository/app:f54a333e",
						},
					},
				},
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name:    "standard-fq",
							Image:   "my.registry/repository/app:f54a333e",
							ImageID: "my.registry/repository/app@sha256:4b59f7dacd37c688968756d176139715df69d89eb0be1802e059316f9d58d9ef",
						},
					},
				},
			},
			podContainerImage{
				Image: stasv1alpha1.Image{
					Name:   "my.registry/repository/app",
					Digest: "sha256:4b59f7dacd37c688968756d176139715df69d89eb0be1802e059316f9d58d9ef",
				},
				Tag: "f54a333e",
			}),
		Entry("Standard digested image",
			&corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "standard-digested",
							Image: "stas/echo-server@sha256:793485b42b5c6d97ab10f8cea08467b77711b865e4512aae6a7e70a38145469e",
						},
					},
				},
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name:    "standard-digested",
							Image:   "stas/echo-server@sha256:793485b42b5c6d97ab10f8cea08467b77711b865e4512aae6a7e70a38145469e",
							ImageID: "docker.io/stas/echo-server@sha256:793485b42b5c6d97ab10f8cea08467b77711b865e4512aae6a7e70a38145469e",
						},
					},
				},
			},
			podContainerImage{
				Image: stasv1alpha1.Image{
					Name:   "docker.io/stas/echo-server",
					Digest: "sha256:793485b42b5c6d97ab10f8cea08467b77711b865e4512aae6a7e70a38145469e",
				},
			}),
		Entry("Standard digested image in k3s",
			&corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "ks3-digested",
							Image: "stas/echo-server@sha256:793485b42b5c6d97ab10f8cea08467b77711b865e4512aae6a7e70a38145469e",
						},
					},
				},
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name:    "ks3-digested",
							Image:   "sha256:793485b42b5c6d97ab10f8cea08467b77711b865e4512aae6a7e70a38145469e",
							ImageID: "docker.io/stas/echo-server@sha256:793485b42b5c6d97ab10f8cea08467b77711b865e4512aae6a7e70a38145469e",
						},
					},
				},
			},
			podContainerImage{
				Image: stasv1alpha1.Image{
					Name:   "docker.io/stas/echo-server",
					Digest: "sha256:793485b42b5c6d97ab10f8cea08467b77711b865e4512aae6a7e70a38145469e",
				},
			}),
		Entry("Image imported into k3s",
			&corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "ks3-imported",
							Image: "application-operator/controller:latest",
						},
					},
				},
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name:    "ks3-imported",
							Image:   "docker.io/application-operator/controller:latest",
							ImageID: "sha256:f991b3a7a93c5c0070dde555a1542d5a34508f16e52eced9237f0967e28ddaff",
						},
					},
				},
			},
			podContainerImage{
				Image: stasv1alpha1.Image{
					Name:   "docker.io/application-operator/controller",
					Digest: "sha256:f991b3a7a93c5c0070dde555a1542d5a34508f16e52eced9237f0967e28ddaff",
				},
				Tag: "latest",
			}),
		Entry("Untagged Docker Hub image on corporate OCP",
			&corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "ocp-corp",
							Image: "mysql",
						},
					},
				},
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name:    "ocp-corp",
							Image:   "dummy.registry.mycorp.com/mysql:latest",
							ImageID: "dummy.registry.mycorp.com/mysql@sha256:83469837189400492f32d23cadbfc97fae3dc019871337a841609f0b71a34907",
						},
					},
				},
			},
			podContainerImage{
				Image: stasv1alpha1.Image{
					Name:   "dummy.registry.mycorp.com/mysql",
					Digest: "sha256:83469837189400492f32d23cadbfc97fae3dc019871337a841609f0b71a34907",
				},
				Tag: "latest",
			}),
	)
})
