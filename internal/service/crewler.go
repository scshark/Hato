/**
 * @Author: scshark
 * @Description:
 * @File:  crewler
 * @Date: 12/22/22 12:12 PM
 */
package service

import (
	"errors"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/robfig/cron"

	"github.com/tidwall/gjson"

	"github.com/scshark/Hato/pkg/util"

	"gorm.io/gorm"

	"github.com/sirupsen/logrus"

	"github.com/scshark/Hato/internal/conf"
	"github.com/scshark/Hato/internal/model"
)

type Platform struct {
	Nickname string
	Username string
	Avatar   string
	IP       string
	Password string
}
type PostFormat struct {
	Content []PostFormatItems
	Users   string
	Tags    string
}
type PostFormatItems struct {
	Content string
	Type    model.PostContentT
}

func InitCrawler() {

	// initCrawlerUser 初始化平台数据
	err := initPlatform()
	if err != nil {
		logrus.Errorf("initPlatform 平台数据初始化失败 ：%s", err)
	}
	// 一天清理一次临时文件目录
	err = deleteTempDir(conf.LocalOSSSetting.SavePath + "/media_temp/")
	if err != nil {
		logrus.Errorf("删除临时目录错误 ：%s", err)
	}
	// 平台一天初始化一次
	_dayCron := cron.New()
	err = _dayCron.AddFunc("@daily", cornSyncPlatform)
	err = _dayCron.AddFunc("@daily", cornDeleteTmpDir)
	_dayCron.Start()

	//err = syncTwitterUserImage(10)
	//if err != nil {
	//	logrus.Errorf("syncTwitterUserImage 推特图片数据同步失败 ：%s", err)
	//}
	//同步用户
	err = cornSyncUser()
	if err != nil {
		logrus.Errorf("cornSyncUser 定时同步用户数据启动失败 ：%s", err)
	}
	// 同步推文
	err = cornSyncPost()
	if err != nil {
		logrus.Errorf("cornSyncUser 定时同步推文数据启动失败 ：%s", err)
	}

	// Sync live and twitter data
}
func cornDeleteTmpDir() {
	err := deleteTempDir(conf.LocalOSSSetting.SavePath + "/twitter_image_temp/")
	if err != nil {
		logrus.Errorf("删除临时目录错误 ：%s", err)
	}
}
func cornSyncPost() error {
	_cron := cron.New()
	err := _cron.AddFunc("@every 65s", cornSyncTwitter)
	err = _cron.AddFunc("@every 40s", cornSyncPlatformLives)
	_cron.Start()
	return err
}
func cornSyncUser() error {
	_cron := cron.New()
	err := _cron.AddFunc("@every 10m", cornSyncTwitterUser)
	err = _cron.AddFunc("@every 3m", cornSyncTwitterUserImage)
	_cron.Start()
	return err
}
func cornSyncPlatform() {
	// 初始化平台信息
	err := initPlatform()
	if err != nil {
		logrus.Errorf("initPlatform 平台数据初始化失败 ：%s", err)
	}
}
func cornSyncTwitterUser() {
	// Twitter用户信息 每5分钟同步一次 一次同步 20
	// execNum 每次同步 N 个用户
	err := syncTwitterUser(50)
	if err != nil {
		logrus.Errorf("syncTwitterUser 推特用户数据同步失败 ：%s", err)
	}
}
func cornSyncTwitterUserImage() {
	// Twitter用户图片数据 7分钟同步一次 一次 10
	// execNum 每次同步 N 个用户
	err := syncTwitterUserImage(10)
	if err != nil {
		logrus.Errorf("syncTwitterUserImage 推特图片数据同步失败 ：%s", err)
	}
}

func cornSyncTwitter() {
	// 同步推特 10s , 一次 10 个用户 ，单个用户 50 条
	// userNum 每次同步 N 个用户
	// tweetNum 每个用户同步 N 条推特
	err := syncTwitter(10, 7)
	if err != nil {
		logrus.Errorf("syncTwitter 推特数据同步失败 ：%s", err)
	}
}

func cornSyncPlatformLives() {
	// 同步平台lives 10s , 全部平台 （4） ，单个 平台 25 条
	// livesNu 每个平台同步 N 条lives
	err := syncPlatform(10)
	if err != nil {
		logrus.Errorf("syncPlatform 平台lives 数据同步失败 ：%s", err)
	}
}
func deleteTempDir(dirPath string) error {
	pwd, _ := os.Getwd()
	tempDir := pwd + "/" + dirPath
	//获取文件或目录相关信息
	fileInfoList, err := ioutil.ReadDir(tempDir)
	if err != nil {
		logrus.Errorf("获取目录失败 ioutil.ReadDir err %s", err)
		return err
	}
	today := time.Now().Format("2006-01-02")
	for _, f := range fileInfoList {
		if f.Name() == today {
			continue
		}
		err = os.RemoveAll(tempDir + f.Name())
		if err != nil {
			logrus.Errorf("删除临时目录失败 os.Remove err %s", err)
		}
	}
	return nil
}

// 初始化平台信息 （4）

// 初始化推特用户信息
// 同步信息到 Hato
// TODO syncPlatformLive

// 初始化平台信息
func initPlatform() error {

	initPlatformNum := 0
	for _, platform := range conf.CrawlerSetting.Platform {

		//
		if len(platform) != 5 {
			return errors.New("crawler platform 配置错误")
		}

		// 0 nickname
		// 1 username
		// 2 PWD
		// 3 Avatar
		// 4 ip

		// 查找用户信息
		user, err := ds.GetUserByUsername(platform[1])

		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			logrus.Errorf("初始化platform错误，GetUserByUsername error : %s", err)
			continue
		}

		if err == nil && user.Model != nil && user.ID > 0 {

			// not need update
			user.Nickname = platform[0]
			user.Avatar = platform[3]
			user.LoginIp = platform[4]

			err = ds.UpdateUser(user)
			if err != nil {
				logrus.Errorf("初始化platform错误，UpdateUserInfo error : %s", err)
				continue
			}
		} else {
			// create
			password, salt := EncryptPasswordAndSalt(platform[2])

			user := &model.User{
				Nickname:          platform[0],
				Username:          platform[1],
				Password:          password,
				Avatar:            platform[3],
				IsCrawlerPlatform: 1,
				Salt:              salt,
				Status:            model.UserStatusNormal,
			}

			_, err := ds.CreateUser(user)
			if err != nil {
				logrus.Errorf("初始化platform错误，CreateUser error : %s", err)
				continue
			}

		}
		initPlatformNum++

	}
	if initPlatformNum != len(conf.CrawlerSetting.Platform) {
		return errors.New("初始化platform错误，未全部实现初始化")
	}
	logrus.Info("Platform初始化完成")
	return nil
}

func syncTwitterUser(execNum int) error {

	conditions := &model.ConditionsT{
		"need_hato_update = ?": 1,
		"ORDER":                "hato_updated_at ASC,id ASC",
	}

	twUser, err := ds.GetTweetUserList(conditions, 0, execNum)

	if err != nil {
		logrus.Errorf("同步推特用户数据失败 ，errors ：%s", err)
		return err
	}
	// 每次同步 200 条

	logrus.Infof("开始同步推特用户数据，此次需要处理数据 %d  条", len(twUser))
	var updateTwUserId = make([]int64, 0)
	for k, u := range twUser {

		// 查找用户信息
		user, e := ds.GetUserByUsername(u.ScreenName)

		if e != nil && !errors.Is(e, gorm.ErrRecordNotFound) {
			logrus.Errorf("同步推特用户数据错误，GetUserByUsername error : %s", err)
			continue
		}

		if user.Model != nil && user.ID > 0 {

			tUser := user
			logrus.Infof("开始第 %d 个推特用户数据同步，twitter username: %s，用户已存在Hato，准备更新数据", k+1, u.ScreenName)
			// not need update
			user.HtTwitterUserId = u.Model.ID
			user.Nickname = u.Name
			user.Description = u.Description
			user.DescriptionUrls = u.DescriptionUrls
			user.Location = u.Location
			user.Urls = u.Urls
			user.FollowersCount = u.FollowersCount
			user.FriendsCount = u.FriendsCount
			user.ProfileImgUrl = u.ProfileImageUrl
			user.ProfileBannerUrl = u.ProfileBannerUrl
			user.Model.CreatedOn = u.TweetCreatedAt
			user.IsCrawlerUser = 1
			//user.Model.ModifiedOn = time.Now().Unix()

			if user.ProfileImgUrl != u.ProfileImageUrl || user.ProfileBannerUrl != u.ProfileBannerUrl {
				user.IsSyncImage = 1
			}
			if tUser != user {
				err = ds.UpdateUser(user)

				if err != nil {
					logrus.Errorf("同步推特用户数据错误，user screen name %s,UpdateUserInfo error : %s", u.ScreenName, err)
					continue
				}
			}
			logrus.Infof("同步推特用户数据，user screen name %s 用户数据无需更新", u.ScreenName)
		} else {
			logrus.Infof("开始第 %d 个推特用户数据同步，twitter username: %s，新用户准备新增数据", k+1, u.ScreenName)

			// create
			password, salt := EncryptPasswordAndSalt(u.ScreenName + "@123")

			user := &model.User{
				Model: &model.Model{
					ModifiedOn: u.TweetCreatedAt,
				},
				HtTwitterUserId:  u.Model.ID,
				Nickname:         u.Name,
				Username:         u.ScreenName,
				DescriptionUrls:  u.DescriptionUrls,
				Description:      u.Description,
				Location:         u.Location,
				Urls:             u.Urls,
				FollowersCount:   u.FollowersCount,
				FriendsCount:     u.FriendsCount,
				Avatar:           u.ProfileImageUrl,
				BannerUrl:        u.ProfileBannerUrl,
				ProfileImgUrl:    u.ProfileImageUrl,
				ProfileBannerUrl: u.ProfileBannerUrl,
				IsCrawlerUser:    1,
				Password:         password,
				Salt:             salt,
				Status:           model.UserStatusNormal,
				IsSyncImage:      1,
			}

			_, err := ds.CreateUser(user)
			if err != nil {
				logrus.Errorf("同步推特用户数据错误，user screen name %s,CreateUser error : %s", u.ScreenName, err)
				continue
			}

		}

		logrus.Infof("第 %d 个推特用户数据，twitter username: %s，同步完成", k+1, u.ScreenName)

		// 已更新用户修改更新时间
		updateTwUserId = append(updateTwUserId, u.Model.ID)
	}
	if len(updateTwUserId) < 1 {
		return errors.New("没有已完成同步的数据")
	}
	if len(twUser) != len(updateTwUserId) {
		logrus.Errorf("同步推特用户数据未完成 ，需要同步 %d 条，实际同步 %d 条，请检查", len(twUser), len(updateTwUserId))
	}

	logrus.Info("同步推特用户数据完成，开始更新采集数据Hato更新时间")
	// 保存后更新hato_updated_at
	err = ds.UpdateTweetUserHatoUpdatedAt(updateTwUserId)
	if err != nil {
		logrus.Errorf("同步推特用户数据完成后更新采集数据Hato更新时间失败 ，errors %s ", err)
		return err
	}
	logrus.Infof("本次同步推特用户数据已全部完成，需要同步 %d 条，实际同步 %d 条", len(twUser), len(updateTwUserId))
	return nil
}

func syncTwitterUserImage(execNum int) error {

	conditions := &model.ConditionsT{
		"is_sync_image = ?":      1,
		"is_crawler_user = ?":    1,
		"ht_twitter_user_id > ?": 0,
		"ORDER":                  "modified_on ASC,id DESC",
	}

	twUser, err := ds.GetUserList(conditions, 0, execNum)

	if err != nil {
		logrus.Errorf("同步推特用户图片数据失败 ，errors ：%s", err)
		return err
	}
	logrus.Infof("开始同步推特用户图片数据，本次执行 %d 条数据 ", len(twUser))
	var updateTwUserId = make([]int64, 0)
	for _, u := range twUser {

		if u.ProfileImgUrl == "" && u.ProfileBannerUrl == "" {
			logrus.Infof("用户 %s 头像和背景都为空，无需同步 ", u.Username)
			updateTwUserId = append(updateTwUserId, u.Model.ID)
			continue
		}
		// 查找用户信息
		user, e := ds.GetUserByID(u.ID)

		if e != nil {
			logrus.Errorf("同步推特用户图片数据失败，GetUserByUsername error : %s", err)
			continue
		}

		// 下载用户图片到临时

		if u.ProfileImgUrl != "" {
			logrus.Infof("用户 %s 开始同步头像 链接为 %s ", u.Username, u.ProfileImgUrl)

			avatarUrl, err := downLoadImageUrlToUploadOss(u.ProfileImgUrl)
			if avatarUrl == "" {
				err = errors.New("头像地址为空")
			}
			if err != nil {
				logrus.Errorf("推特用户同步头像图片数据失败  downLoadImageUrlToUploadOss 用户ScreenName ：%s,  error %s", u.Username, err)
				continue
			}

			user.Avatar = avatarUrl
		}

		if u.ProfileBannerUrl != "" {
			logrus.Infof("用户 %s 开始同步 banner图片 链接为 %s ", u.Username, u.ProfileBannerUrl)

			bannerUrl, err := downLoadImageUrlToUploadOss(u.ProfileBannerUrl)
			if bannerUrl == "" {
				err = errors.New("背景banner地址为空")
			}
			if err != nil {
				logrus.Errorf("推特用户同步banner图片数据失败  downLoadImageUrlToUploadOss 用户ScreenName ：%s,  error %s", u.Username, err)
				continue
			}
			user.BannerUrl = bannerUrl
		}
		// 同步到数据
		logrus.Infof("用户 %s 已经完成 数据同步，开始保存数据到 hato ,头像url %s", u.Username, user.Avatar)
		logrus.Infof("用户 %s 已经完成 数据同步，开始保存数据到 hato 背景url %s", u.Username, user.BannerUrl)

		err = ds.UpdateUser(user)

		if err != nil {
			logrus.Errorf("同步推特用户图片数据错误，user screen name %s,UpdateUser error : %s", u.Username, err)
			continue
		}

		// 更新同步标签
		updateTwUserId = append(updateTwUserId, u.Model.ID)
	}
	// 更新推特数据同步标签

	logrus.Info("同步推特用户图片数据完成，开始更新图片同步标签")
	// 保存后更新hato_updated_at
	err = ds.UpdateUserSyncImage(updateTwUserId)
	if err != nil {
		logrus.Errorf("同步推特用户图片数据完成后更新图片同步标签失败 ，errors %s ", err)
		return err
	}
	logrus.Infof("同步推特用户图片数据已全部完成，需要同步 %d 条，实际同步 %d 条", len(twUser), len(updateTwUserId))
	// 同步回数据库
	return nil
}

func downLoadImageUrlToUploadOss(iUrl string) (imageUrl string, err error) {

	urlParse, err := url.Parse(iUrl)

	if err != nil {
		return "", err
	}
	localFile, err := util.DownloaderSave(iUrl, urlParse.Path)
	if err != nil || localFile == "" {
		logrus.Errorf("下载图片失败 DownloaderSave  error %s or localfile is empty", err)
		return "", err
	}

	// 上传oss
	obKey := "twitter/" + time.Now().Format("2006-01-02") + urlParse.Path
	if path.Ext(urlParse.Path) == "" {
		obKey = obKey + path.Ext(localFile)
	}
	imageUrl, err = oss.PutFileInput(obKey, localFile)
	if err != nil {
		logrus.Errorf("图片上传oss失败 PutFileInput  error %s ,obKey %s, localFile %s", err, obKey, localFile)
		return "", err
	}

	return imageUrl, nil
}

func syncTwitter(userNum int, tweetNum int) error {
	// 获取hato用户列表

	logrus.Info("-----------------开始推文同步-----------------")
	conditions := &model.ConditionsT{
		"is_crawler_user = ?":    1,
		"ht_twitter_user_id > ?": 0,
		"ORDER":                  "post_updated_at ASC,id ASC",
	}

	htUser, err := ds.GetUserList(conditions, 0, userNum)

	if err != nil {
		// 获取hato 用户列表信息失败
		logrus.Errorf("同步推文，获取Hato用户列表信息失败 err:%s", err)
		return err
	}
	logrus.Info("同步推文，获取Hato用户列表，获取到 %d 个用户", len(htUser))

	for _, u := range htUser {

		tweetCond := &model.ConditionsT{
			"ht_twitter_user_id = ?": u.HtTwitterUserId,
			"full_text != ?":         "",
			"is_tweet = ?":           0,
			"ORDER":                  "tw_created_at DESC,id DESC",
		}
		logrus.Infof("同步推文，获取Hato用户推特信息 用户id %d，名称 %s", u.ID, u.Username)

		tweet, err := ds.GetTweetList(tweetCond, 0, tweetNum)
		if err != nil {
			logrus.Errorf("同步推文，获取用户推文失败 用户id%d ， err:%s", u.Model.ID, err)
			continue
		}
		if len(tweet) == 0 {
			logrus.Infof("同步推文，用户名称 %s , 所有推文都已同步完成，无需更新", u.Username)
			logrus.Infof("用户名称 %s，开始更新推特用户同步推文时间", u.Username)
			u.PostUpdatedAt = time.Now().Unix()
			err = ds.UpdateUser(u)
			if err != nil {
				logrus.Infof("用户名称 %s，更新推特用户同步推文时间失败， error %s ", u.Username, err)
				return err
			}
			continue

		}
		// 推特用户同步推文
		logrus.Infof("同步推文，用户名称 %s ,获取到推文 %d 条，正在同步", u.Username, len(tweet))

		var tweetIds = make([]int64, 0)
		for k, t := range tweet {

			// 检查推文是否已存在

			// tags == hashtags
			tweetTags := make([]string, 0)
			if t.Hashtags != "" {
				tagsParse := gjson.Parse(t.Hashtags).Array()
				for _, tg := range tagsParse {
					tweetTags = append(tweetTags, tg.String())
				}
			}
			// user == user_mentions

			// content
			tags := tagsFrom(tweetTags)

			var postContent = make([]*model.PostContent, 0)
			// 去掉full_text里面的 url
			re, _ := regexp.Compile(`(http|https):\/\/([\w\-_]+(?:(?:\.[\w\-_]+)+))([\w\-\.,@?^=%&amp;:/~\+#]*[\w\-\@?^=%&amp;/~\+#])?`)

			fullText := re.ReplaceAllString(t.FullText, "")

			if fullText != "" {
				// full_text
				pContent := &model.PostContent{
					UserID:  u.ID,
					Content: fullText,
					Type:    model.CONTENT_TYPE_TEXT,
					Sort:    100,
					Model: &model.Model{
						CreatedOn:  t.TwCreatedAt,
						ModifiedOn: t.TwCreatedAt,
					},
				}
				postContent = append(postContent, pContent)
			}

			// extended_entities
			if t.ExtendedEntities != "" {

				extendEnt := gjson.Parse(t.ExtendedEntities)
				if extendEnt.IsArray() {
					extendEnt.ForEach(func(key, value gjson.Result) bool {

						mediaType := value.Get("type")
						var contentMediaUrl string
						var mediaTye model.PostContentT
						switch mediaType.String() {

						default:
							mediaUrl := value.Get("media_url").String()
							if mediaUrl == "" {
								return true
							}
							// download mediaUrl
							contentMediaUrl, err = downLoadImageUrlToUploadOss(mediaUrl)
							if err != nil {
								logrus.Errorf("**** 推文信息Media同步（图片） 推文图片下载上传Oss操作错误，用户名称 %s ，推文 id %s , URL %s,downLoadImageUrlToUploadOss error : %s", u.Username, t.IdStr, contentMediaUrl, err)
								return true
							}
							mediaTye = model.CONTENT_TYPE_IMAGE
						case "video":
							videoUrl := value.Get("video_url").String()
							if videoUrl == "" {
								return true
							}
							// download mediaUrl
							contentMediaUrl, err = downLoadImageUrlToUploadOss(videoUrl)
							if err != nil {
								logrus.Errorf("**** 推文信息Media同步（视频） 推文视频下载上传Oss操作错误，用户名称 %s ，推文 id %s  URL %s,downLoadImageUrlToUploadOss error : %s", u.Username, t.IdStr, contentMediaUrl, err)
								return true
							}
							mediaTye = model.CONTENT_TYPE_VIDEO
						}

						mediaContent := &model.PostContent{
							UserID:  u.ID,
							Content: contentMediaUrl,
							Type:    mediaTye,
							Sort:    101,
							Model: &model.Model{
								CreatedOn:  t.TwCreatedAt,
								ModifiedOn: t.TwCreatedAt,
							},
						}
						postContent = append(postContent, mediaContent)
						return true
					})
				}
			}
			// urls
			if t.Urls != "" && t.Urls != "[]" {

				urlsParse := gjson.Parse(t.Urls).Array()

				for _, tUrl := range urlsParse {
					tweetUrl := tUrl.String()
					urlContent := &model.PostContent{
						UserID:  u.ID,
						Content: tweetUrl,
						Type:    model.CONTENT_TYPE_LINK,
						Sort:    102,
						Model: &model.Model{
							CreatedOn:  t.TwCreatedAt,
							ModifiedOn: t.TwCreatedAt,
						},
					}
					postContent = append(postContent, urlContent)
				}

			}

			post := &model.Post{
				UserID:          u.Model.ID,
				Tags:            strings.Join(tags, ","),
				IPLoc:           u.Location,
				LatestRepliedOn: t.TwCreatedAt,
				TweetId:         t.IdStr,
				Model: &model.Model{
					CreatedOn:  t.TwCreatedAt,
					ModifiedOn: t.TwCreatedAt,
				},
			}
			post, err = ds.CreateCrawlerPost(post, postContent)
			if err != nil {
				if post != nil && post.Model != nil {
					tweetIds = append(tweetIds, t.ID)
				}
				logrus.Errorf("**** 推文信息同步失败，用户名称 %s ，推文 id %s ，CreatePost error : %s", u.Username, t.IdStr, err)
				continue
			}

			// 创建标签
			for _, ts := range tags {

				tag := &model.Tag{
					UserID: u.ID,
					Tag:    ts,
					Model: &model.Model{
						CreatedOn:  t.TwCreatedAt,
						ModifiedOn: t.TwCreatedAt,
					},
				}
				_, err := ds.CreateTag(tag)
				if err != nil {
					logrus.Errorf("**** 推文信息Tags同步失败，用户名称 %s ，用户 id %d  ,CreateTag error : %s", u.Username, u.ID, err)
				}
			}

			//if t.UserMentions != "" {
			//	userMt := gjson.Parse(t.UserMentions)
			//	if userMt.IsArray() {
			//		userMt.ForEach(func(key, value gjson.Result) bool {
			//			// 创建用户消息提醒
			//
			//			user, err := ds.GetUserByUsername(value.Get("screen_name").String())
			//			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			//				logrus.Errorf("**** 推文信息UserMentions同步失败，用户名称 %s ，用户 id %d , user screen name is %s ,GetUserByUsername error : %s", u.Username, u.ID, value.Get("screen_name").String(), err)
			//				return true
			//			}
			//
			//			if user.Model == nil {
			//				logrus.Warnf("推文信息UserMentions @ 的账号不存在")
			//				return true
			//			}
			//			// 创建消息提醒
			//			// TODO: 优化消息提醒处理机制
			//			go ds.CreateMessage(&model.Message{
			//				SenderUserID:   u.ID,
			//				ReceiverUserID: user.ID,
			//				Type:           model.MsgTypePost,
			//				Brief:          "在新发布的Hato动态中@了你",
			//				PostID:         post.ID,
			//				Model: &model.Model{
			//					CreatedOn:  t.TwCreatedAt,
			//					ModifiedOn: t.TwCreatedAt,
			//				},
			//			})
			//			return true
			//		})
			//	}
			//}

			// 推送Search
			PushPostToSearch(post)

			logrus.Infof("同步推文，用户名称 %s ,获取到推文 %d 条，正在同步推文，正在处理第 %d 条", u.Username, len(tweet), k+1)

			tweetIds = append(tweetIds, t.ID)
		}
		logrus.Infof("同步推文，用户名称 %s ,获取到推文 %d 条，已完成该用户的同步操作", u.Username, len(tweet))
		// 同步完成

		if len(tweetIds) <= 0 {
			logrus.Errorf("同步推文失败，用户名称 %s ,获取到推文 %d 条，完成同步 0 条", u.Username, len(tweet))
			continue
		}
		// 更新推特同步标签 is_tweet
		err = ds.UpdateTweetSyncStatus(tweetIds)
		if err != nil {
			logrus.Errorf("同步推文更新推特同步标签失败，用户名称 %s ,用户id %d，error %s", u.Username, u.ID, err)
			continue
		}
		logrus.Infof("同步推文完成，用户名称 %s  用户 ID %d ,获取到推文 %d 条，成功处理 %d 条", u.Username, u.ID, len(tweet), len(tweetIds))

		// 更新推特用户同步推文时间 post_updated_at

		logrus.Infof("用户名称 %s，开始更新推特用户同步推文时间", u.Username)

		u.PostUpdatedAt = time.Now().Unix()

		err = ds.UpdateUser(u)
		if err != nil {
			logrus.Infof("用户名称 %s，更新推特用户同步推文时间失败， error %s ", u.Username, err)
			return err
		}
		logrus.Infof("用户名称 %s，完成本次推文同步操作", u.Username)

	}

	logrus.Infof("已完成本次列表所有用户推文同步")

	return nil

}

func syncPlatform(livesNum int) error {

	logrus.Info("-----------------开始平台Lives同步-----------------")
	conditions := &model.ConditionsT{
		"is_crawler_platform = ?": 1,
	}

	htUser, err := ds.GetUserList(conditions, 0, 0)

	if err != nil {
		// 获取hato 用户列表信息失败
		logrus.Errorf("平台Lives同步，获取Hato用户列表信息失败 err:%s", err)
		return err
	}
	logrus.Info("平台Lives同步，获取Hato用户列表，获取到 %d 个 平台用户", len(htUser))

	for _, u := range htUser {

		var postLivesId = make([]int64, 0)
		// 获取数据
		conditions := &model.ConditionsT{
			"is_tweet": 0,
			"ORDER":    "created_on DESC,id DESC",
		}
		listData, err := ds.GetLivesList(u.Username, conditions, 0, livesNum)

		if err != nil {
			logrus.Errorf("平台Lives同步,获取 平台 %s lives 列表失败 err %s", u.Nickname, err)
			continue
		}
		// 组建数据
		for c, d := range listData {

			liveTags := d.Tags

			// tags
			tags := tagsFrom(liveTags)

			post := &model.Post{
				UserID:          u.Model.ID,
				Tags:            strings.Join(tags, ","),
				IP:              u.LoginIp,
				IPLoc:           util.GetIPLoc(u.LoginIp),
				LatestRepliedOn: d.CreatedOn,
				Model: &model.Model{
					CreatedOn:  d.CreatedOn,
					ModifiedOn: d.CreatedOn,
				},
			}

			var postContent = make([]*model.PostContent, 0)
			for k, items := range d.LiveItems {

				content := items.Content
				if items.ContentType == model.CONTENT_TYPE_IMAGE {

					imageUrl, err := downLoadImageUrlToUploadOss(content)
					if err == nil {
						content = imageUrl
					} else {
						//logrus.Errorf("**** 平台lives信息 Image 上传失败，平台名称 %s ，downLoadImageUrlToUploadOss error : %s", u.Nickname, err)
						continue
					}
				}
				pContent := &model.PostContent{
					UserID:  u.ID,
					Content: content,
					Type:    items.ContentType,
					Sort:    int64(100 + k),
					Model: &model.Model{
						CreatedOn:  d.CreatedOn,
						ModifiedOn: d.CreatedOn,
					},
				}
				postContent = append(postContent, pContent)

			}
			post, err = ds.CreateCrawlerPost(post, postContent)
			if err != nil {
				if post != nil && post.Model != nil {
					postLivesId = append(postLivesId, d.LiveId)
				}
				logrus.Errorf("**** 平台lives信息同步失败，平台名称 %s ，CreateCrawlerPost error : %s", u.Nickname, err)
				continue
			}

			// 创建标签
			for _, ts := range tags {
				tag := &model.Tag{
					UserID: u.ID,
					Tag:    ts,
					Model: &model.Model{
						CreatedOn:  d.CreatedOn,
						ModifiedOn: d.CreatedOn,
					},
				}
				_, err := ds.CreateTag(tag)
				if err != nil {
					logrus.Errorf("**** 平台lives信息Tags同步失败，平台名称 %s ，CreateTag error : %s", u.Nickname, err)
				}
			}
			// 发布完成
			// 推送Search
			PushPostToSearch(post)

			logrus.Infof("同步平台lives，平台名称 %s ,获取到lives %d 条，正在同步推文，正在处理第 %d 条", u.Nickname, len(listData), c+1)

			// 记录已发布live
			postLivesId = append(postLivesId, d.LiveId)

		}
		// 修改 is_tweet 标签已发布数据
		err = ds.UpdateLivesIsTweet(u.Username, postLivesId)
		if err != nil {
			logrus.Errorf("**** 平台lives 修改标签is_tweet 失败 ，平台名称 %s ，处理的id %v , UpdateLivesIsTweet error : %s", u.Nickname, postLivesId, err)
			continue
		}
		logrus.Infof("平台名称 %s , 同步平台lives全部同步完成，获取到lives %d 条，已处理 %d 条", u.Nickname, len(listData), len(postLivesId))
		// 发布数据
		u.PostUpdatedAt = time.Now().Unix()
		err = ds.UpdateUser(u)
		if err != nil {
			logrus.Errorf("**** 平台 %s 修改推文更新时间标签 PostUpdatedAt 失败 ， PostUpdatedAt error : %s", u.Nickname, err)
			continue
		}
	}
	logrus.Infof("本次平台lives推文同步已完成")

	return nil
}
