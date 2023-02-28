package util

import (
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/scshark/Hato/internal/conf"

	"github.com/sirupsen/logrus"
)

func DownloaderSave(url string, fileUri string) (string, error) {

	fileURI := conf.LocalOSSSetting.SavePath + "/media_temp/" + time.Now().Format("2006-01-02") + fileUri

	fi, err := os.Stat(fileURI)

	if err == nil && !fi.IsDir() {
		logrus.Errorf("object exist so do nothing objectKey: %s", fileURI)
		return fileURI, nil
	}

	saveDir := filepath.Dir(fileURI)
	if err = os.MkdirAll(saveDir, 0750); err != nil && !os.IsExist(err) {
		return "", err
	}

	return writeFile(url, fileURI)
}

func writeFile(URL string, fileURI string) (string, error) {

	client := &http.Client{
		Timeout: time.Duration(15) * time.Second,
	}

	// Supply http request with headers to ensure a higher possibility of success
	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("content-type", "text/html")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("Proxy-Connection", "keep-alive")

	//for k, v := range headers {
	//	req.Header.Set(k, v)
	//}

	res, err := client.Do(req)
	if err != nil {
		logrus.Errorf("------err : %v---------", err)
		return "", err
	}
	//fmt.Printf("Url: %s, Status: %s, Size: %d", URL, res.Status, res.ContentLength)
	if res.StatusCode != http.StatusOK {
		time.Sleep(1 * time.Second)
		res, err = client.Get(URL)
		if err != nil {
			return "", err
		}
	}
	defer res.Body.Close()

	if path.Ext(fileURI) == "" {
		switch res.Header.Get("content-type") {
		case "image/jpeg":
			if strings.Index(fileURI, ".jp") < 0 {
				fileURI = fileURI + ".jpg"
			}
		case "image/png":
			if strings.Index(fileURI, ".png") < 0 {
				fileURI = fileURI + ".png"
			}
		case "image/gif":
			if strings.Index(fileURI, ".gif") < 0 {
				fileURI = fileURI + ".gif"
			}
		case "image/vnd.wap.wbmp":
			if strings.Index(fileURI, ".wbmp") < 0 {
				fileURI = fileURI + ".wbmp"
			}
		case "video/mp4":
			if strings.Index(fileURI, ".mp4") < 0 {
				fileURI = fileURI + ".mp4"
			}
		}

	}

	openOpts := os.O_RDWR | os.O_CREATE

	file, err := os.OpenFile(fileURI, openOpts, 0666)
	if err != nil {
		logrus.Errorf("OpenFile error %s", err)
		return "", err
	}
	defer file.Close()

	var writer io.Writer
	writer = file
	//some sites do not return "content-type" or "content-length" in http header

	// Note that io.Copy reads 32kb(maximum) from input and writes them to output, then repeats.
	// So don't worry about memory.
	_, copyErr := io.Copy(writer, res.Body)
	if copyErr != nil && copyErr != io.EOF {
		logrus.Errorf("Copy error %s", copyErr)
		return "", copyErr
	}
	return fileURI, nil
}
