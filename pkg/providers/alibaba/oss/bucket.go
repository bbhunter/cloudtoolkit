package oss

import (
	"context"
	"strings"

	"github.com/404tk/cloudtoolkit/pkg/schema"
	"github.com/404tk/cloudtoolkit/utils/logger"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type Driver struct {
	Cred   *credentials.StsTokenCredential
	Region string
}

func (d *Driver) NewClient() *oss.Client {
	region := d.Region
	if region == "all" {
		region = "cn-hangzhou"
	}
	client, _ := oss.New(
		"https://oss-"+region+".aliyuncs.com",
		d.Cred.AccessKeyId,
		d.Cred.AccessKeySecret,
		oss.SecurityToken(d.Cred.AccessKeyStsToken))
	return client
}

func (d *Driver) GetBuckets(ctx context.Context) ([]schema.Storage, error) {
	list := []schema.Storage{}
	select {
	case <-ctx.Done():
		return list, nil
	default:
		logger.Info("List OSS buckets ...")
	}
	client := d.NewClient()
	response, err := client.ListBuckets(oss.MaxKeys(1000))
	if err != nil {
		logger.Error("List buckets failed.")
		return list, err
	}

	for _, bucket := range response.Buckets {
		/*
			if !strings.Contains(d.Client.Config.Endpoint, bucket.Location) {
				continue
			}
		*/
		_bucket := schema.Storage{
			BucketName: bucket.Name,
			Region:     strings.TrimPrefix(bucket.Location, "oss-"),
		}
		list = append(list, _bucket)
	}

	return list, nil
}
