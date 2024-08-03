package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/liuminhaw/mist-miner/shared"
	"github.com/liuminhaw/mm-plugins/utils"
)

type requestPaymentMiner struct {
	propertyType  string
	serviceClient *s3Client
	configuration *s3.GetBucketRequestPaymentOutput
}

func newRequestPaymentMiner(
	serviceClient utils.Client,
	property string,
) (*requestPaymentMiner, error) {
	client, err := assertS3Client(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newRequestPaymentMiner: %w", err)
	}

	return &requestPaymentMiner{propertyType: property, serviceClient: client}, nil
}

func (rp *requestPaymentMiner) PropertyType() string { return rp.propertyType }

func (rp *requestPaymentMiner) FetchConf(input any) error {
	requestPaymentInput, ok := input.(*s3.GetBucketRequestPaymentInput)
	if !ok {
		return fmt.Errorf("fetchConf: GetBucketRequestPaymentInput type assertion failed")
	}

	var err error
	rp.configuration, err = rp.serviceClient.client.GetBucketRequestPayment(
		context.Background(),
		requestPaymentInput,
	)
	if err != nil {
		return fmt.Errorf("fetchConf bucket requestPayment: %w", err)
	}

	return nil
}

func (rp *requestPaymentMiner) Generate(dummy utils.CacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := rp.FetchConf(&s3.GetBucketRequestPaymentInput{Bucket: rp.serviceClient.bucket.Name}); err != nil {
		return nil, fmt.Errorf("generate bucket requestPaymentProp: %w", err)
	}

	property := shared.MinerProperty{
		Type: requestPayment,
		Label: shared.MinerPropertyLabel{
			Name:   "RequestPayment",
			Unique: true,
		},
		Content: shared.MinerPropertyContent{
			Format: shared.FormatText,
		},
	}
	if err := property.FormatContentValue(rp.configuration.Payer); err != nil {
		return nil, fmt.Errorf("generate requestPaymentProp: %w", err)
	}

	properties = append(properties, property)
	return properties, nil
}
