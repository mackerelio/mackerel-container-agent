package config

import (
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type downloader interface {
	download(u *url.URL) ([]byte, error)
}

type s3Downloader struct{}

func (s3Downloader) download(u *url.URL) ([]byte, error) {
	sess := session.Must(session.NewSession())
	downloader := s3manager.NewDownloader(sess)

	buf := &aws.WriteAtBuffer{}
	_, err := downloader.Download(buf, &s3.GetObjectInput{
		Bucket: aws.String(u.Host),
		Key:    aws.String(u.Path),
	})
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

var s3downloader downloader = s3Downloader{}

func fetchS3(u *url.URL) ([]byte, error) {
	return s3downloader.download(u)
}
