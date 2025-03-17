package config

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type downloader interface {
	download(context.Context, *url.URL) ([]byte, error)
}

type s3Downloader struct {
	regionHint string
}

func (d s3Downloader) download(ctx context.Context, u *url.URL) ([]byte, error) {
	var (
		bucket = u.Host
		key    = strings.TrimPrefix(u.Path, "/")
	)

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(d.regionHint))
	if err != nil {
		return nil, err
	}

	region, err := manager.GetBucketRegion(ctx, s3.NewFromConfig(cfg), bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to get bucket region for %s: %w", bucket, err)
	}
	cfg.Region = region

	downloader := manager.NewDownloader(s3.NewFromConfig(cfg))

	buf := manager.NewWriteAtBuffer([]byte{})
	_, err = downloader.Download(ctx, buf, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
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
