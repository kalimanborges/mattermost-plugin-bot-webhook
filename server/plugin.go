package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"
)

type Configuration struct {
	BotUsername string
	BotUserID   string
	WebhookURL  string
	BearerToken string

	BotUsername2 string
	BotUserID2   string
	WebhookURL2  string
	BearerToken2 string

	BotUsername3 string
	BotUserID3   string
	WebhookURL3  string
	BearerToken3 string

	BotUsername4 string
	BotUserID4   string
	WebhookURL4  string
	BearerToken4 string

	BotUsername5 string
	BotUserID5   string
	WebhookURL5  string
	BearerToken5 string

	BotUsername6 string
	BotUserID6   string
	WebhookURL6  string
	BearerToken6 string

	BotUsername7 string
	BotUserID7   string
	WebhookURL7  string
	BearerToken7 string

	BotUsername8 string
	BotUserID8   string
	WebhookURL8  string
	BearerToken8 string

	BotUsername9 string
	BotUserID9   string
	WebhookURL9  string
	BearerToken9 string

	BotUsername10 string
	BotUserID10   string
	WebhookURL10  string
	BearerToken10 string
}

type botMapping struct {
	BotUserID   string
	WebhookURL  string
	BearerToken string
}

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type BotWebhookPlugin struct {
	plugin.MattermostPlugin
	configuration *Configuration
}

// activeBots returns the list of configured bot mappings with non-empty user IDs.
// If a bot has a username but no user ID, it attempts to resolve the username via the API.
func (p *BotWebhookPlugin) activeBots() []botMapping {
	cfg := p.configuration
	all := []struct {
		username    string
		userID      string
		webhookURL  string
		bearerToken string
	}{
		{cfg.BotUsername, cfg.BotUserID, cfg.WebhookURL, cfg.BearerToken},
		{cfg.BotUsername2, cfg.BotUserID2, cfg.WebhookURL2, cfg.BearerToken2},
		{cfg.BotUsername3, cfg.BotUserID3, cfg.WebhookURL3, cfg.BearerToken3},
		{cfg.BotUsername4, cfg.BotUserID4, cfg.WebhookURL4, cfg.BearerToken4},
		{cfg.BotUsername5, cfg.BotUserID5, cfg.WebhookURL5, cfg.BearerToken5},
		{cfg.BotUsername6, cfg.BotUserID6, cfg.WebhookURL6, cfg.BearerToken6},
		{cfg.BotUsername7, cfg.BotUserID7, cfg.WebhookURL7, cfg.BearerToken7},
		{cfg.BotUsername8, cfg.BotUserID8, cfg.WebhookURL8, cfg.BearerToken8},
		{cfg.BotUsername9, cfg.BotUserID9, cfg.WebhookURL9, cfg.BearerToken9},
		{cfg.BotUsername10, cfg.BotUserID10, cfg.WebhookURL10, cfg.BearerToken10},
	}

	active := make([]botMapping, 0, len(all))
	for _, b := range all {
		id := b.userID
		// Fallback: resolve username if user ID is not yet populated
		if id == "" && b.username != "" {
			u := strings.TrimPrefix(b.username, "@")
			if user, appErr := p.API.GetUserByUsername(u); appErr == nil {
				id = user.Id
			} else {
				p.API.LogWarn("Could not resolve bot username", "username", u, "error", appErr.Error())
			}
		}
		if id != "" {
			active = append(active, botMapping{id, b.webhookURL, b.bearerToken})
		}
	}
	return active
}

func (p *BotWebhookPlugin) OnConfigurationChange() error {
	var configuration Configuration
	if err := p.API.LoadPluginConfiguration(&configuration); err != nil {
		p.API.LogError("Failed to load configuration", "error", err.Error())
		return err
	}

	// Resolve @username fields to user IDs
	updated := false
	resolve := func(username *string, userID *string) {
		if *username == "" {
			return
		}
		u := strings.TrimPrefix(*username, "@")
		user, appErr := p.API.GetUserByUsername(u)
		if appErr != nil {
			p.API.LogWarn("Failed to resolve username", "username", u, "error", appErr.Error())
			return
		}
		if *userID != user.Id {
			*userID = user.Id
			updated = true
		}
	}

	resolve(&configuration.BotUsername, &configuration.BotUserID)
	resolve(&configuration.BotUsername2, &configuration.BotUserID2)
	resolve(&configuration.BotUsername3, &configuration.BotUserID3)
	resolve(&configuration.BotUsername4, &configuration.BotUserID4)
	resolve(&configuration.BotUsername5, &configuration.BotUserID5)
	resolve(&configuration.BotUsername6, &configuration.BotUserID6)
	resolve(&configuration.BotUsername7, &configuration.BotUserID7)
	resolve(&configuration.BotUsername8, &configuration.BotUserID8)
	resolve(&configuration.BotUsername9, &configuration.BotUserID9)
	resolve(&configuration.BotUsername10, &configuration.BotUserID10)

	// Set configuration in memory first so the plugin is immediately operational
	// with the resolved IDs, regardless of when SavePluginConfig completes.
	p.configuration = &configuration

	if updated {
		// Write resolved IDs back to the config store in a goroutine to avoid
		// blocking and to prevent re-entrant OnConfigurationChange ordering issues.
		configSnapshot := configuration
		go func() {
			configBytes, err := json.Marshal(configSnapshot)
			if err != nil {
				p.API.LogError("Failed to marshal configuration for save", "error", err.Error())
				return
			}
			var configMap map[string]interface{}
			if err := json.Unmarshal(configBytes, &configMap); err != nil {
				p.API.LogError("Failed to unmarshal configuration for save", "error", err.Error())
				return
			}
			if appErr := p.API.SavePluginConfig(configMap); appErr != nil {
				p.API.LogError("Failed to save resolved configuration", "error", appErr.Error())
			}
		}()
	}

	return nil
}

func (p *BotWebhookPlugin) MessageHasBeenPosted(c *plugin.Context, post *model.Post) {
	p.API.LogDebug("MessageHasBeenPosted")

	channel, err := p.API.GetChannel(post.ChannelId)
	if err != nil {
		p.API.LogError("Failed to get channel", "error", err.Error())
		return
	}

	for _, bot := range p.activeBots() {
		if post.UserId == bot.BotUserID {
			continue
		}

		if !strings.Contains(channel.Name, bot.BotUserID) {
			continue
		}

		p.API.LogDebug("Message to bot detected", "channel", channel.Name, "user", post.UserId, "message", post.Message, "bot", bot.BotUserID)

		jsonPayload, err := json.Marshal(post)
		if err != nil {
			p.API.LogError("Failed to marshal JSON payload", "error", err.Error())
			continue
		}

		req, err := http.NewRequest("POST", bot.WebhookURL, bytes.NewBuffer(jsonPayload))
		if err != nil {
			p.API.LogError("Failed to create HTTP request", "error", err.Error())
			continue
		}
		req.Header.Set("Authorization", "Bearer "+bot.BearerToken)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			p.API.LogError("Failed to make an HTTP request", "error", err.Error())
			continue
		}
		resp.Body.Close()
		break
	}
}

func (p *BotWebhookPlugin) OnActivate() error {
	return p.OnConfigurationChange()
}

func main() {
	plugin.ClientMain(&BotWebhookPlugin{})
}
