package orchestration

import (
	"time"
)

type Event struct {
	Version   string    `mapstructure:"version,omitempty" yaml:"version,omitempty"`
	Id        string    `mapstructure:"id,omitempty" yaml:"id,omitempty"`
	Type      string    `mapstructure:"type,omitempty" yaml:"type,omitempty" json:"detail-type"`
	Source    string    `mapstructure:"source,omitempty" yaml:"source,omitempty"`
	Account   string    `mapstructure:"account,omitempty" yaml:"account,omitempty"`
	Time      time.Time `mapstructure:"time,omitempty" yaml:"time,omitempty"`
	Region    string    `mapstructure:"region,omitempty" yaml:"region,omitempty"`
	Resources []string  `mapstructure:"resources,omitempty" yaml:"resources,omitempty"`
	Detail    Detail    `mapstructure:"detail,omitempty" yaml:"detail,omitempty"`
}

type Detail struct {
	Type                 string    `mapstructure:"type,omitempty" yaml:"type,omitempty"`
	PlacementId          string    `mapstructure:"placementId,omitempty" yaml:"placementId,omitempty"`
	Port                 int       `mapstructure:"port,omitempty" yaml:"port,omitempty"`
	GameSessionArn       string    `mapstructure:"gameSessionArn,omitempty" yaml:"gameSessionArn,omitempty"`
	IpAddress            string    `mapstructure:"ipAddress,omitempty" yaml:"ipAddress,omitempty"`
	DnsName              string    `mapstructure:"dnsName,omitempty" yaml:"dnsName,omitempty"`
	StartTime            time.Time `mapstructure:"startTime,omitempty" yaml:"startTime,omitempty"`
	EndTime              time.Time `mapstructure:"endTime,omitempty" yaml:"endTime,omitempty"`
	GameSessionRegion    string    `mapstructure:"gameSessionRegion,omitempty" yaml:"gameSessionRegion,omitempty"`
	PlacedPlayerSessions []string  `mapstructure:"placedPlayerSessions,omitempty" yaml:"placedPlayerSessions,omitempty"`
}
