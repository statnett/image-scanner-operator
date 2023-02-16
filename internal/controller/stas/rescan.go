package stas

import (
	"context"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	stasv1alpha1 "github.com/statnett/image-scanner-operator/api/stas/v1alpha1"
	"github.com/statnett/image-scanner-operator/internal/config"
)

type RescanTrigger struct {
	client.Client
	config.Config
	EventChan     chan event.GenericEvent
	CheckInterval time.Duration
}

func (r *RescanTrigger) Start(ctx context.Context) error {
	log := logf.FromContext(ctx)

	ticker := time.NewTicker(r.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			cisList := &stasv1alpha1.ContainerImageScanList{}
			if err := r.List(ctx, cisList, client.InNamespace("")); err != nil {
				log.Error(err, "failed to list CISes")
				continue
			}

			for i := range cisList.Items {
				cis := cisList.Items[i]

				lastScanTime := cis.Status.LastScanTime
				if !lastScanTime.IsZero() && time.Since(lastScanTime.Time) > r.ScanInterval {
					r.EventChan <- event.GenericEvent{Object: &cis}
				}
			}
		}
	}
}
