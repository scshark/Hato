package service

import "github.com/scshark/Hato/internal/model"

func CreateAttachment(attachment *model.Attachment) (*model.Attachment, error) {
	return ds.CreateAttachment(attachment)
}
