package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	dc "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"

	// "github.com/docker/docker/pkg/term"
	"github.com/moby/term"
)

type dockerClient struct {
	client   *dc.Client
	platform string
	debug    bool
}

func newDockerClient(platform string) (*dockerClient, error) {
	client, err := dc.NewClientWithOpts(dc.FromEnv, dc.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create a new docker API client, reason: %w", err)
	}
	return &dockerClient{
		client:   client,
		platform: platform,
		debug:    isLogLevelDebug() || isLogLevelTrace(),
	}, nil
}

func (d *dockerClient) pullImage(ctx context.Context, imageName string) error {
	progress, err := d.client.ImagePull(ctx, imageName, image.PullOptions{Platform: d.platform})
	if err != nil {
		return fmt.Errorf("failed to pull the image %s, reason: %w", imageName, err)
	}
	defer progress.Close()

	// Store info about existing locally available image.
	availLocal, id := d.isImageAvailableLocally(ctx, imageName)
	// Show verbose pull progress only if either in debug mode or
	// there is no existing locally available image.
	showPullProgress := d.debug || !availLocal

	// Perform the actual image pull.
	if showPullProgress {
		log.Infof("Pulling image: %s", imageName)
		log.InfoEmpty()
		termFd, isTerm := term.GetFdInfo(os.Stdout)
		err = jsonmessage.DisplayJSONMessagesStream(progress, os.Stdout, termFd, isTerm, nil)
		log.InfoEmpty()
	} else {
		_, err = io.Copy(io.Discard, progress)
	}
	if err != nil {
		return fmt.Errorf("failed while pulling the image %s, reason: %w", imageName, err)
	}

	// If pull progress was already shown, no need to show the updates again.
	if showPullProgress {
		log.Debugf("Pulled image successfully: %s", imageName)
		return nil
	}

	// Otherwise, determine if the image was updated and show the updated ID
	// of the image.
	avail, newId := d.isImageAvailableLocally(ctx, imageName)
	if !avail {
		log.Fatalf("Image is expected to be available after pull, but is unavailable possibly indicating a bug or system failure!")
	}
	if newId != id {
		log.Infof("Pulled newer version of image %s: %s", imageName, newId)
	}
	return nil
}

func (d *dockerClient) isImageAvailableLocally(ctx context.Context, imageName string) (bool, string) {
	filter := filters.NewArgs()
	filter.Add("reference", imageName)
	images, err := d.client.ImageList(ctx, image.ListOptions{
		All:            false,
		Filters:        filter,
		SharedSize:     false,
		ContainerCount: false,
		Manifests:      false,
	})

	// Ignore errors by considering the image is not available locally in
	// case of errors.
	if err != nil || len(images) == 0 {
		return false, ""
	}

	return true, images[0].ID
}

func (d *dockerClient) close() {
	d.client.Close()
}
