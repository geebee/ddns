/*
   Copyright 2022 https://github.com/geebee

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package ddns

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/rs/zerolog"
)

const (
	defaultRefreshInterval = time.Hour * 24
)

type DynamicDNS struct {
	ctx       context.Context
	cancelCtx context.CancelFunc

	startOnce sync.Once

	l zerolog.Logger

	API *cloudflare.API

	Record      cloudflare.DNSRecord
	IPLookupURL string

	refreshTicker *time.Ticker
}

func (ddns DynamicDNS) Context() context.Context {
	return ddns.ctx
}

func (ddns DynamicDNS) Refresh() error {
	currentIP, err := externalIP(ddns.IPLookupURL)
	if err != nil {
		return fmt.Errorf("failed to lookup external IP at: %s: %w", ddns.IPLookupURL, err)
	}

	ddns.l.Info().
		Str("fqdn", ddns.Record.Name).
		Str("previous", ddns.Record.Content).
		Str("current", currentIP).
		Bool("update_required", currentIP != ddns.Record.Content).
		Msg("refresh")

	if currentIP != ddns.Record.Content {
		if err := ddns.API.UpdateDNSRecord(ddns.Context(), ddns.Record.ZoneID, ddns.Record.ID, ddns.Record); err != nil {
			return fmt.Errorf("failed to update dns record: %s: %w", ddns.Record.Name, err)
		}
	}

	return nil
}

func (ddns DynamicDNS) Start() {
	ddns.startOnce.Do(func() {
		ddns.l.Info().Msg("starting")
		ddns.Refresh()

		go func() {
			for {
				select {
				case <-ddns.ctx.Done():
					ddns.l.Debug().Msg("context completed")
					return
				case <-ddns.refreshTicker.C:
					if err := ddns.Refresh(); err != nil {
						ddns.l.Error().Err(err).Msg("failed to refresh dynamic DNS")
					}
				}
			}
		}()
	})
}

func (ddns DynamicDNS) Stop() {
	ddns.l.Info().Msg("stopping")

	ddns.refreshTicker.Stop()
	ddns.cancelCtx()
	time.Sleep(time.Millisecond)

	ddns.l.Info().Msg("stopped")
}

func NewDynamicDNSFromEnv() *DynamicDNS {
	ddns, err := NewDynamicDNS(
		os.Getenv("CLOUDFLARE_API_KEY"),
		os.Getenv("DDNS_HOST"),
		os.Getenv("DDNS_DOMAIN"),
		os.Getenv("IP_LOOKUP_URL"),
		os.Getenv("REFRESH_INTERVAL"),
	)
	if err != nil {
		panic(fmt.Errorf("failed to create new dynamic DNS instance from environment: %w", err))
	}

	return ddns
}

func NewDynamicDNS(apiKey, hostname, domain, ipLookupURL, refreshInterval string) (*DynamicDNS, error) {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).Level(zerolog.DebugLevel).With().Timestamp().Caller().Logger()

	parsedRefreshInterval, err := time.ParseDuration(refreshInterval)
	if err != nil {
		logger.Warn().
			Str("requested_interval", refreshInterval).
			Str("default_interval", defaultRefreshInterval.String()).
			Msg("failed to parse requested interval duration; using default")

		parsedRefreshInterval = defaultRefreshInterval
	}

	api, err := cloudflare.NewWithAPIToken(apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cloudflare API handle from API key: %w", err)
	}

	zoneID, err := api.ZoneIDByName(domain)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve zone ID from name: %s: %w", domain, err)
	}

	fqdn := fmt.Sprintf("%s.%s", hostname, domain)

	ctx, cancelCtx := context.WithCancel(context.Background())
	setupCtx, cancelSetupCtx := context.WithTimeout(ctx, time.Second*5)
	defer cancelSetupCtx()

	records, err := api.DNSRecords(setupCtx, zoneID, cloudflare.DNSRecord{Name: fqdn})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve zone DNS records: %w", err)
	}

	var ddnsRecord cloudflare.DNSRecord
	for _, r := range records {
		if r.Name == fqdn {
			ddnsRecord = r
			break
		}
	}

	logger.Info().
		Str("fqdn", fqdn).
		Str("ip", ddnsRecord.Content).
		Str("refresh_interval", parsedRefreshInterval.String()).
		Bool("exists", !ddnsRecord.CreatedOn.IsZero()).
		Msg("initial state")

	if ddnsRecord.Content == "" {
		currentIP, err := externalIP(ipLookupURL)
		if err != nil {
			return nil, fmt.Errorf("failed to lookup external IP at: %s: %w", ipLookupURL, err)
		}

		logger.Info().
			Str("fqdn", fqdn).
			Str("ip", currentIP).
			Msg("creating missing DNS record")

		createCtx, cancelCreateCtx := context.WithTimeout(ctx, time.Second*5)
		defer cancelCreateCtx()
		response, err := api.CreateDNSRecord(createCtx, zoneID, cloudflare.DNSRecord{
			Type:    "A",
			Name:    fqdn,
			Content: currentIP,
			TTL:     1,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create new dynamic DNS record: %s: %w", fqdn, err)
		}

		ddnsRecord = response.Result
	}

	return &DynamicDNS{
		ctx:       ctx,
		cancelCtx: cancelCtx,

		l: logger,

		API: api,

		Record:      ddnsRecord,
		IPLookupURL: ipLookupURL,

		refreshTicker: time.NewTicker(parsedRefreshInterval),
	}, nil
}
