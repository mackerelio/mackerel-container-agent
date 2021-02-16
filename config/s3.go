package config

import (
	"context"
	"fmt"
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type downloader interface {
	download(context.Context, *url.URL) ([]byte, error)
}

type s3Downloader struct {
	regionHint string
}

func (d s3Downloader) download(ctx context.Context, u *url.URL) ([]byte, error) {
	sess := session.Must(session.NewSession())

	r, err := s3manager.GetBucketRegion(ctx, sess, u.Host, d.regionHint)
	if err != nil {
		return nil, fmt.Errorf("failed to get bucket region for %s: %w", u.Host, err)
	}
	sess.Config.Region = aws.String(r)

	downloader := s3manager.NewDownloader(sess)

	buf := &aws.WriteAtBuffer{}
	_, err = downloader.DownloadWithContext(ctx, buf, &s3.GetObjectInput{
		Bucket: aws.String(u.Host),
		Key:    aws.String(u.Path),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download config from %s: %w", u, err)
	}

	return buf.Bytes(), nil
}

var s3downloader downloader = s3Downloader{
	regionHint: "ap-northeast-1",
}

func fetchS3(ctx context.Context, u *url.URL) ([]byte, error) {
	return s3downloader.download(ctx, u)
}
