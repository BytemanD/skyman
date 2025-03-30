package internal

import (
	"fmt"
	"time"

	"github.com/BytemanD/go-console/console"
	"github.com/BytemanD/skyman/utility"
)

type ServerSnapshot struct {
	ServerActionTest
	EmptyCleanup
	imageId string
}

func (t *ServerSnapshot) Start() error {
	imageName := fmt.Sprintf("skyman-image-for-%s", t.Server.Name)
	imageId, err := t.Client.NovaV2().ServerCreateImage(t.Server.Id, imageName, nil)
	if err != nil {
		return err
	}
	t.imageId = imageId
	console.Info("[%s] creating image %s", t.Server.Id, imageId)
	return utility.RetryWithErrors(
		utility.RetryCondition{
			Timeout:     time.Minute * 10,
			IntervalMin: time.Second * 2},
		[]string{"ImageNotActive"},
		func() error {
			image, err := t.Client.GlanceV2().GetImage(imageId)
			if err != nil {
				return fmt.Errorf("get image %s failed: %s", imageId, err)
			}
			console.Info("[%s] image status=%s", t.Server.Id, image.Status)
			if image.IsError() {
				return fmt.Errorf("image %s is error", imageId)
			}
			if image.IsActive() {
				return nil
			}
			return utility.NewImageNotActiveError(imageId)
		},
	)
}
func (t ServerSnapshot) TearDown() error {
	if t.imageId == "" {
		return nil
	}
	console.Info("[%s] request to delete image %s", t.Server.Id, t.imageId)
	t.Client.GlanceV2().DeleteImage(t.imageId)
	return nil
}
