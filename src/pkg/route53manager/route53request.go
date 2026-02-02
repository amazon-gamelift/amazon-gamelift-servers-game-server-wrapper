package route53manager

import (
	"context"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/pkg/errors"
)

func route53Request(ctx context.Context, data route53RequestData, client *route53.Client, logger *slog.Logger) error {
	logger.DebugContext(ctx, "requesting route53 upsert",
		"name", data.recordName,
		"value", data.recordValue,
		"ttl", data.recordTtl,
		"type", data.recordType,
		"comment", data.recordComment,
		"hostedZoneId", data.hostedZoneId)

	input := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &types.ChangeBatch{
			Changes: []types.Change{
				{
					Action: types.ChangeActionUpsert,
					ResourceRecordSet: &types.ResourceRecordSet{
						Name: aws.String(data.recordName),
						ResourceRecords: []types.ResourceRecord{
							{
								Value: aws.String(data.recordValue),
							},
						},
						TTL:  aws.Int64(data.recordTtl),
						Type: types.RRType(data.recordType),
					},
				},
			},
			Comment: aws.String(data.recordComment),
		},
		HostedZoneId: aws.String(data.hostedZoneId),
	}

	_, err := client.ChangeResourceRecordSets(ctx, input)
	if err != nil {
		logger.ErrorContext(ctx, "error performing route53 upsert Request", "err", err)
		return errors.Wrap(err, "error performing route53 upsert Request")
	}

	return nil
}

type route53RequestData struct {
	recordName    string
	recordValue   string
	recordTtl     int64
	recordType    string
	recordComment string
	hostedZoneId  string
}
