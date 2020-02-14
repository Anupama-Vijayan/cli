package v2action

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/SermoDigital/jose/jws"

	"code.cloudfoundry.org/cli/actor/sharedaction"
)

func (actor Actor) GetStreamingLogs(appGUID string, client sharedaction.LogCacheClient) (<-chan sharedaction.LogMessage, <-chan error, context.CancelFunc) {
	return sharedaction.GetStreamingLogs(appGUID, client)
}

func (actor Actor) GetRecentLogsForApplicationByNameAndSpace(appName string, spaceGUID string, client sharedaction.LogCacheClient) ([]sharedaction.LogMessage, Warnings, error) {
	app, allWarnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return nil, allWarnings, err
	}

	logCacheMessages, err := sharedaction.GetRecentLogs(app.GUID, client)
	if err != nil {
		return nil, allWarnings, err
	}

	var logMessages []sharedaction.LogMessage

	for _, message := range logCacheMessages {
		logMessages = append(logMessages, *sharedaction.NewLogMessage(
			message.Message(),
			message.Type(),
			message.Timestamp(),
			message.SourceType(),
			message.SourceInstance(),
		))
	}

	return logMessages, allWarnings, nil
}

func (actor Actor) GetStreamingLogsForApplicationByNameAndSpace(appName string, spaceGUID string, client sharedaction.LogCacheClient) (<-chan sharedaction.LogMessage, <-chan error, context.CancelFunc, Warnings, error) {
	app, allWarnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return nil, nil, func() {}, allWarnings, err
	}

	messages, logErrs, stopStreaming := actor.GetStreamingLogs(app.GUID, client)

	return messages, logErrs, stopStreaming, allWarnings, err
}

func (actor Actor) ScheduleTokenRefresh(tickerChan <-chan time.Time) (chan bool, error) {
	accessTokenString, err := actor.RefreshAccessToken(actor.Config.RefreshToken())
	if err != nil {
		return nil, err
	}

	accessTokenString = strings.TrimPrefix(accessTokenString, "bearer ")
	token, err := jws.ParseJWT([]byte(accessTokenString))
	if err != nil {
		return nil, err
	}

	var timeToRefresh time.Duration
	expiration, ok := token.Claims().Expiration()
	if ok {
		expiresIn := time.Until(expiration)
		timeToRefresh = expiresIn * 9 / 10
	} else {
		return nil, errors.New("Failed to get an expiry time from the current access token")
	}
	quitNowChannel := make(chan bool, 1)

	go func() {
		if tickerChan == nil {
			ticker := time.NewTicker(timeToRefresh)
			defer ticker.Stop()
			tickerChan = ticker.C
		}
		for {
			select {
			case <-tickerChan:
				_, err := actor.RefreshAccessToken(actor.Config.RefreshToken())
				if err != nil {
					panic(err)
				}
			case <-quitNowChannel:
				return
			}
		}
	}()

	return quitNowChannel, nil
}
