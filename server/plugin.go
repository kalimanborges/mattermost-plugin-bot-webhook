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

func (c *Configuration) activeBots() []botMapping {
	all := []botMapping{
		{c.BotUserID, c.WebhookURL, c.BearerToken},
		{c.BotUserID2, c.WebhookURL2, c.BearerToken2},
		{c.BotUserID3, c.WebhookURL3, c.BearerToken3},
		{c.BotUserID4, c.WebhookURL4, c.BearerToken4},
		{c.BotUserID5, c.WebhookURL5, c.BearerToken5},
		{c.BotUserID6, c.WebhookURL6, c.BearerToken6},
		{c.BotUserID7, c.WebhookURL7, c.BearerToken7},
		{c.BotUserID8, c.WebhookURL8, c.BearerToken8},
		{c.BotUserID9, c.WebhookURL9, c.BearerToken9},
		{c.BotUserID10, c.WebhookURL10, c.BearerToken10},
	}
	active := make([]botMapping, 0, len(all))
	for _, b := range all {
		if b.BotUserID != "" {
			active = append(active, b)
		}
	}
	return active
}

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type BotWebhookPlugin struct {
	plugin.MattermostPlugin
	configuration *Configuration
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
			p.API.LogError("Failed to resolve username", "username", u, "error", appErr.Error())
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

	if updated {
		configBytes, err := json.Marshal(configuration)
		if err != nil {
			p.API.LogError("Failed to marshal configuration", "error", err.Error())
		} else {
			var configMap map[string]interface{}
			if err := json.Unmarshal(configBytes, &configMap); err != nil {
				p.API.LogError("Failed to unmarshal configuration", "error", err.Error())
			} else if appErr := p.API.SavePluginConfig(configMap); appErr != nil {
				p.API.LogError("Failed to save resolved configuration", "error", appErr.Error())
			}
		}
	}

	p.configuration = &configuration
	return nil
}

func (p *BotWebhookPlugin) MessageHasBeenPosted(c *plugin.Context, post *model.Post) {
	p.API.LogDebug("MessageHasBeenPosted")

	channel, err := p.API.GetChannel(post.ChannelId)
	if err != nil {
		p.API.LogError("Failed to get channel", "error", err.Error())
		return
	}

	for _, bot := range p.configuration.activeBots() {
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
