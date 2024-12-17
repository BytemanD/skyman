package internal

import (
	"fmt"
	"time"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/utility"
)

type ServerSnapshot struct {
	ServerActionTest
	EmptyCleanup
	imageId string
}

func (t *ServerSnapshot) Start() error {
	imageName := fmt.Sprintf("skyman-image-for-%s", t.Server.Name)
	imageId, err := t.Client.NovaV2().Server().CreateImage(t.Server.Id, imageName, nil)
	if err != nil {
		return err
	}
	t.imageId = imageId
	logging.Info("[%s] creating image %s", t.Server.Id, imageId)
	return utility.RetryWithErrors(
		utility.RetryCondition{
			Timeout:     time.Minute * 10,
			IntervalMin: time.Second * 2},
		[]string{"ImageNotActive"},
		func() error {
			image, err := t.Client.GlanceV2().Images().Show(imageId)
			if err != nil {
				return fmt.Errorf("get image %s failed: %s", imageId, err)
			}
			logging.Info("[%s] image status=%s", t.Server.Id, image.Status)
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
	logging.Info("[%s] request to delete image %s", t.Server.Id, t.imageId)
	t.Client.GlanceV2().Images().Delete(t.imageId)
	return nil
}
