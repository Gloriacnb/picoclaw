package channels

import (
	"context"
	"time"

	"github.com/sipeed/picoclaw/pkg/bus"
	runtimeevents "github.com/sipeed/picoclaw/pkg/events"
)

const channelEventPublishTimeout = 100 * time.Millisecond

func channelTypeForEvent(m *Manager, channelName string) string {
	if m == nil || m.config == nil {
		return channelName
	}
	if bc := m.config.Channels.Get(channelName); bc != nil && bc.Type != "" {
		return bc.Type
	}
	return channelName
}

func (m *Manager) publishChannelEvent(
	kind runtimeevents.Kind,
	channelName string,
	scope runtimeevents.Scope,
	severity runtimeevents.Severity,
	payload any,
) {
	if m == nil || m.runtimeEvents == nil {
		return
	}
	if scope.Channel == "" {
		scope.Channel = channelName
	}
	ctx, cancel := context.WithTimeout(context.Background(), channelEventPublishTimeout)
	defer cancel()
	m.runtimeEvents.Publish(ctx, runtimeevents.Event{
		Kind:     kind,
		Source:   runtimeevents.Source{Component: "channel", Name: channelName},
		Scope:    scope,
		Severity: severity,
		Payload:  payload,
	})
}

func (m *Manager) publishOutboundSent(
	channelName string,
	msg bus.OutboundMessage,
	messageIDs []string,
) {
	m.publishChannelEvent(
		runtimeevents.KindChannelMessageOutboundSent,
		channelName,
		scopeFromOutboundContext(msg.Context),
		runtimeevents.SeverityInfo,
		ChannelOutboundPayload{
			ContentLen:       len([]rune(msg.Content)),
			MessageIDs:       append([]string(nil), messageIDs...),
			ReplyToMessageID: msg.ReplyToMessageID,
		},
	)
}

func (m *Manager) publishOutboundQueued(
	channelName string,
	msg bus.OutboundMessage,
) {
	m.publishChannelEvent(
		runtimeevents.KindChannelMessageOutboundQueued,
		channelName,
		scopeFromOutboundContext(msg.Context),
		runtimeevents.SeverityInfo,
		ChannelOutboundPayload{
			ContentLen:       len([]rune(msg.Content)),
			ReplyToMessageID: msg.ReplyToMessageID,
		},
	)
}

func (m *Manager) publishOutboundFailed(
	channelName string,
	msg bus.OutboundMessage,
	err error,
	media bool,
) {
	payload := ChannelOutboundPayload{
		Media:            media,
		ContentLen:       len([]rune(msg.Content)),
		ReplyToMessageID: msg.ReplyToMessageID,
		Retries:          maxRetries,
	}
	if err != nil {
		payload.Error = err.Error()
	}
	m.publishChannelEvent(
		runtimeevents.KindChannelMessageOutboundFailed,
		channelName,
		scopeFromOutboundContext(msg.Context),
		runtimeevents.SeverityError,
		payload,
	)
}

func (m *Manager) publishOutboundMediaSent(
	channelName string,
	msg bus.OutboundMediaMessage,
	messageIDs []string,
) {
	m.publishChannelEvent(
		runtimeevents.KindChannelMessageOutboundSent,
		channelName,
		scopeFromOutboundContext(msg.Context),
		runtimeevents.SeverityInfo,
		ChannelOutboundPayload{
			Media:      true,
			MessageIDs: append([]string(nil), messageIDs...),
		},
	)
}

func (m *Manager) publishOutboundMediaQueued(
	channelName string,
	msg bus.OutboundMediaMessage,
) {
	m.publishChannelEvent(
		runtimeevents.KindChannelMessageOutboundQueued,
		channelName,
		scopeFromOutboundContext(msg.Context),
		runtimeevents.SeverityInfo,
		ChannelOutboundPayload{Media: true},
	)
}

func (m *Manager) publishOutboundMediaFailed(
	channelName string,
	msg bus.OutboundMediaMessage,
	err error,
) {
	payload := ChannelOutboundPayload{
		Media:   true,
		Retries: maxRetries,
	}
	if err != nil {
		payload.Error = err.Error()
	}
	m.publishChannelEvent(
		runtimeevents.KindChannelMessageOutboundFailed,
		channelName,
		scopeFromOutboundContext(msg.Context),
		runtimeevents.SeverityError,
		payload,
	)
}

func scopeFromOutboundContext(ctx bus.InboundContext) runtimeevents.Scope {
	return runtimeevents.Scope{
		Channel:   ctx.Channel,
		Account:   ctx.Account,
		ChatID:    ctx.ChatID,
		TopicID:   ctx.TopicID,
		SpaceID:   ctx.SpaceID,
		SpaceType: ctx.SpaceType,
		ChatType:  ctx.ChatType,
		SenderID:  ctx.SenderID,
		MessageID: ctx.MessageID,
	}
}
