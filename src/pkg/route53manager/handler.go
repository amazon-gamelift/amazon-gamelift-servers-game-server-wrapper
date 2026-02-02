package route53manager

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	wrapperConfig "github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/internal/config"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/helpers"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/types/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/pkg/errors"
)

func SetupRoute53Mappings(ctx context.Context, logger *slog.Logger, zoneId string, cfg *wrapperConfig.Config, creds *events.AwsCredentials, requestHandler *helpers.HttpRequestHandler) error {

	token, err := requestHandler.Request(ctx, helpers.HttpRequestDetails{
		Method:  "PUT",
		Url:     cfg.Route53.TokenUrl,
		Headers: map[string]string{cfg.Route53.TokenHeaderKey: cfg.Route53.TokenHeaderValue},
	})
	if err != nil {
		logger.ErrorContext(ctx, "error getting token", "err", err)
		return errors.Wrap(err, "error getting token")
	}

	metaData, err := requestHandler.Request(ctx, helpers.HttpRequestDetails{
		Method:  "GET",
		Url:     cfg.Route53.MetaDataUrl,
		Headers: map[string]string{cfg.Route53.MetaDataHeaderKey: token},
	})
	if err != nil {
		logger.ErrorContext(ctx, "error getting metaData", "err", err)
		return errors.Wrap(err, "error getting metadata")
	}

	publicIp, err := requestHandler.Request(ctx, helpers.HttpRequestDetails{
		Method:  "GET",
		Url:     strings.Replace(cfg.Route53.PublicIpUrl, "{{metaData}}", metaData, -1),
		Headers: map[string]string{cfg.Route53.MetaDataHeaderKey: token},
	})
	if err != nil {
		logger.ErrorContext(ctx, "error getting publicIp", "err", err)
		return errors.Wrap(err, "error getting public ip")
	}

	privateIp, err := requestHandler.Request(ctx, helpers.HttpRequestDetails{
		Method:  "GET",
		Url:     strings.Replace(cfg.Route53.PrivateIpUrl, "{{metaData}}", metaData, -1),
		Headers: map[string]string{cfg.Route53.MetaDataHeaderKey: token},
	})
	if err != nil {
		logger.ErrorContext(ctx, "error getting privateIp", "err", err)
		return errors.Wrap(err, "error getting private ip")
	}

	r53Config, err := config.LoadDefaultConfig(ctx, config.WithRegion(cfg.Route53.Region), config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
		Value: aws.Credentials{
			AccessKeyID: creds.AccessKeyId, SecretAccessKey: creds.SecretAccessKey, SessionToken: creds.SessionToken, Source: "Fleet Instance Role",
		},
	}))
	if err != nil {
		logger.ErrorContext(ctx, "failed to retrieve given r53 config", "err", err)
		r53Config, err = config.LoadDefaultConfig(ctx)
		if err != nil {
			logger.ErrorContext(ctx, "error getting default r53 config", "err", err)
			return errors.Wrap(err, "error getting default r53 config")
		}
	}

	client := route53.NewFromConfig(r53Config)
	recordName := fmt.Sprintf("%s.worlds.%s", zoneId, cfg.Route53.HostDomain) // TODO: Check format

	if cfg.Route53.PublicHostedZoneId != "" && publicIp != "" {
		err = route53Request(ctx, route53RequestData{
			recordName:    recordName,
			recordValue:   publicIp,
			recordTtl:     cfg.Route53.Ttl,
			recordType:    cfg.Route53.Type,
			recordComment: cfg.Route53.Comment,
			hostedZoneId:  cfg.Route53.PublicHostedZoneId,
		}, client, logger)
		if err != nil {
			logger.ErrorContext(ctx, "error performing route53 Request for public hosted zone", "err", err)
			return errors.Wrap(err, "error performing route53 Request for public hosted zone")
		}
	} else {
		logger.Debug("skipping route53 setup for public ip", "publicIp", publicIp, "public hosted zoneId", cfg.Route53.PublicHostedZoneId)
	}

	if cfg.Route53.PrivateHostedZoneId != "" && privateIp != "" {
		err = route53Request(ctx, route53RequestData{
			recordName:    recordName,
			recordValue:   privateIp,
			recordTtl:     cfg.Route53.Ttl,
			recordType:    cfg.Route53.Type,
			recordComment: cfg.Route53.Comment,
			hostedZoneId:  cfg.Route53.PrivateHostedZoneId,
		}, client, logger)
		if err != nil {
			logger.ErrorContext(ctx, "error performing route53 Request for private hosted zone", "err", err)
			return errors.Wrap(err, "error performing route53 Request for private hosted zone")
		}
	} else {
		logger.DebugContext(ctx, "skipping route53 setup for private ip", "privateIp", privateIp, "private hosted zoneId", cfg.Route53.PrivateHostedZoneId)
	}

	return nil
}
