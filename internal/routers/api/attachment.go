package api

import (
	"image"

	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/scshark/Hato/internal/model"
	"github.com/scshark/Hato/internal/service"
	"github.com/scshark/Hato/pkg/app"
	"github.com/scshark/Hato/pkg/convert"
	"github.com/scshark/Hato/pkg/errcode"
	"github.com/sirupsen/logrus"
)

var uploadAttachmentTypeMap = map[string]model.AttachmentType{
	"public/image":  model.ATTACHMENT_TYPE_IMAGE,
	"public/avatar": model.ATTACHMENT_TYPE_IMAGE,
	"public/video":  model.ATTACHMENT_TYPE_VIDEO,
	"attachment":    model.ATTACHMENT_TYPE_OTHER,
}

func GeneratePath(s string) string {
	n := len(s)
	if n <= 2 {
		return s
	}

	return GeneratePath(s[:n-2]) + "/" + s[n-2:]
}

func GetFileExt(s string) (string, error) {
	switch s {
	case "image/png":
		return ".png", nil
	case "image/jpg":
		return ".jpg", nil
	case "image/jpeg":
		return ".jpeg", nil
	case "image/gif":
		return ".gif", nil
	case "video/mp4":
		return ".mp4", nil
	case "video/quicktime":
		return ".mov", nil
	case "application/zip":
		return ".zip", nil
	default:
		return "", errcode.FileInvalidExt.WithDetails("仅允许 png/jpg/gif/mp4/mov/zip 类型")
	}
}
func GetImageSize(img image.Rectangle) (int, int) {
	b := img.Bounds()
	width := b.Max.X
	height := b.Max.Y
	return width, height
}

func fileCheck(uploadType string, size int64) error {
	if uploadType != "public/video" &&
		uploadType != "public/image" &&
		uploadType != "public/avatar" &&
		uploadType != "attachment" {
		return errcode.InvalidParams
	}

	if size > 1024*1024*100 {
		return errcode.FileInvalidSize.WithDetails("最大允许100MB")
	}

	return nil
}

func UploadAttachment(c *gin.Context) {
	response := app.NewResponse(c)

	uploadType := c.Request.FormValue("type")
	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		logrus.Errorf("api.UploadAttachment err: %v", err)
		response.ToErrorResponse(errcode.FileUploadFailed)
		return
	}
	defer file.Close()

	// 检查上传文件的类型和大小是否合格
	if err = fileCheck(uploadType, fileHeader.Size); err != nil {
		cErr, _ := err.(*errcode.Error)
		response.ToErrorResponse(cErr)
		return
	}

	// 通过上传内容类型，获取出文件后缀
	contentType := fileHeader.Header.Get("Content-Type")
	fileExt, err := GetFileExt(fileHeader.Header.Get("Content-Type"))
	if err != nil {
		logrus.Errorf("GetFileExt err: %v", err)
		response.ToErrorResponse(err.(*errcode.Error))
		return
	}

	// 生成随机路径
	randomPath := uuid.Must(uuid.NewV4()).String()
	ossSavePath := uploadType + "/" + GeneratePath(randomPath[:8]) + "/" + randomPath[9:] + fileExt

	// 上传，
	objectUrl, err := objectStorage.PutObject(ossSavePath, file, fileHeader.Size, contentType, false)
	if err != nil {
		logrus.Errorf("putObject err: %v", err)
		response.ToErrorResponse(errcode.FileUploadFailed)
		return
	}

	// 构造附件Model
	attachment := &model.Attachment{
		FileSize: fileHeader.Size,
		Content:  objectUrl,
	}

	if userID, exists := c.Get("UID"); exists {
		attachment.UserID = userID.(int64)
	}

	attachment.Type = uploadAttachmentTypeMap[uploadType]
	if attachment.Type == model.ATTACHMENT_TYPE_IMAGE {
		var src image.Image
		src, err = imaging.Decode(file)
		if err == nil {
			attachment.ImgWidth, attachment.ImgHeight = GetImageSize(src.Bounds())
		}
	}

	attachment, err = service.CreateAttachment(attachment)
	if err != nil {
		logrus.Errorf("service.CreateAttachment err: %v", err)
		response.ToErrorResponse(errcode.FileUploadFailed)
	}

	response.ToResponse(attachment)
}

func DownloadAttachmentPrecheck(c *gin.Context) {
	response := app.NewResponse(c)

	contentID := convert.StrTo(c.Query("id")).MustInt64()
	// 加载content
	content, err := service.GetPostContentByID(contentID)
	if err != nil {
		logrus.Errorf("service.GetPostContentByID err: %v", err)
		response.ToErrorResponse(errcode.InvalidDownloadReq)
	}
	user, _ := c.Get("USER")
	if content.Type == model.CONTENT_TYPE_CHARGE_ATTACHMENT {
		// 加载post
		post, err := service.GetPost(content.PostID)
		if err != nil {
			logrus.Errorf("service.GetPost err: %v", err)
			response.ToResponse(gin.H{
				"paid": false,
			})
			return
		}

		// 发布者或管理员免费下载
		if post.UserID == user.(*model.User).ID || user.(*model.User).IsAdmin {
			response.ToResponse(gin.H{
				"paid": true,
			})
			return
		}

		// 检测是否有购买记录
		response.ToResponse(gin.H{
			"paid": service.CheckPostAttachmentIsPaid(post.ID, user.(*model.User).ID),
		})
		return
	}
	response.ToResponse(gin.H{
		"paid": false,
	})
}

func DownloadAttachment(c *gin.Context) {
	response := app.NewResponse(c)

	contentID := convert.StrTo(c.Query("id")).MustInt64()

	// 加载content
	content, err := service.GetPostContentByID(contentID)
	if err != nil {
		logrus.Errorf("service.GetPostContentByID err: %v", err)
		response.ToErrorResponse(errcode.InvalidDownloadReq)
	}

	// 收费附件
	if content.Type == model.CONTENT_TYPE_CHARGE_ATTACHMENT {
		user, _ := c.Get("USER")

		// 加载post
		post, err := service.GetPost(content.PostID)
		if err != nil {
			logrus.Errorf("service.GetPost err: %v", err)
			response.ToResponse(gin.H{
				"paid": false,
			})
			return
		}

		paidFlag := false

		// 发布者或管理员免费下载
		if post.UserID == user.(*model.User).ID || user.(*model.User).IsAdmin {
			paidFlag = true
		}

		// 检测是否有购买记录
		if service.CheckPostAttachmentIsPaid(post.ID, user.(*model.User).ID) {
			paidFlag = true
		}

		if !paidFlag {
			// 未购买，则尝试购买
			err := service.BuyPostAttachment(&model.Post{
				Model: &model.Model{
					ID: post.ID,
				},
				UserID:          post.UserID,
				AttachmentPrice: post.AttachmentPrice,
			}, user.(*model.User))
			if err != nil {
				logrus.Errorf("service.BuyPostAttachment err: %v", err)
				if err == errcode.InsuffientDownloadMoney {

					response.ToErrorResponse(errcode.InsuffientDownloadMoney)
				} else {

					response.ToErrorResponse(errcode.DownloadExecFail)
				}
				return
			}
		}
	}

	objectKey := objectStorage.ObjectKey(content.Content)
	signedURL, err := objectStorage.SignURL(objectKey, 60)
	if err != nil {
		logrus.Errorf("client.SignURL err: %v", err)
		response.ToErrorResponse(errcode.DownloadReqError)
		return
	}
	response.ToResponse(signedURL)
}
