package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

const pluginID = "bot-webhook-plugin"

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
// If a bot has a username but no user ID, it attempts to resolve the username via the API as fallback.
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
		if id == "" && b.username != "" {
			u := strings.TrimPrefix(b.username, "@")
			if user, appErr := p.API.GetUserByUsername(u); appErr == nil {
				id = user.Id
			}
		}
		if id != "" {
			active = append(active, botMapping{id, b.webhookURL, b.bearerToken})
		}
	}
	return active
}

// getSettingString retrieves a string value from the plugin settings map using
// case-insensitive key lookup — Mattermost may lowercase keys during storage.
func getSettingString(settings map[string]interface{}, key string) string {
	// Try exact key first
	if v, ok := settings[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	// Try lowercase key
	lower := strings.ToLower(key)
	if v, ok := settings[lower]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// ConfigurationWillBeSaved intercepts the config save, resolves @usernames to user IDs,
// and injects the resolved IDs into the config before it is persisted.
// This ensures the BotUserID fields are always populated in the System Console UI.
func (p *BotWebhookPlugin) ConfigurationWillBeSaved(newCfg *model.Config) (*model.Config, error) {
	if newCfg.PluginSettings.Plugins == nil {
		return nil, nil
	}
	settings, ok := newCfg.PluginSettings.Plugins[pluginID]
	if !ok {
		return nil, nil
	}

	// Log all keys present to diagnose case normalization
	keys := make([]string, 0, len(settings))
	for k := range settings {
		keys = append(keys, fmt.Sprintf("%s=%v", k, settings[k]))
	}
	p.API.LogInfo("[BotWebhook] ConfigurationWillBeSaved called", "entries", strings.Join(keys, " | "))

	// setID writes the resolved user ID under both CamelCase and lowercase key variants.
	setID := func(userIDKey, id string) {
		settings[userIDKey] = id
		settings[strings.ToLower(userIDKey)] = id
	}

	resolve := func(usernameKey, userIDKey string) {
		// Always clear the existing ID first.
		// This ensures that if the username changes or is removed, the stale ID
		// from a previous save does not persist.
		setID(userIDKey, "")

		usernameVal := getSettingString(settings, usernameKey)
		if usernameVal == "" {
			return
		}
		u := strings.TrimPrefix(usernameVal, "@")
		user, appErr := p.API.GetUserByUsername(u)
		if appErr != nil {
			p.API.LogError("[BotWebhook] Failed to resolve username", "username", u, "error", appErr.Error())
			return
		}
		p.API.LogInfo("[BotWebhook] Resolved username", "username", u, "userID", user.Id)
		setID(userIDKey, user.Id)
	}

	resolve("BotUsername", "BotUserID")
	resolve("BotUsername2", "BotUserID2")
	resolve("BotUsername3", "BotUserID3")
	resolve("BotUsername4", "BotUserID4")
	resolve("BotUsername5", "BotUserID5")
	resolve("BotUsername6", "BotUserID6")
	resolve("BotUsername7", "BotUserID7")
	resolve("BotUsername8", "BotUserID8")
	resolve("BotUsername9", "BotUserID9")
	resolve("BotUsername10", "BotUserID10")

	newCfg.PluginSettings.Plugins[pluginID] = settings
	return newCfg, nil
}

func (p *BotWebhookPlugin) OnConfigurationChange() error {
	var configuration Configuration
	if err := p.API.LoadPluginConfiguration(&configuration); err != nil {
		p.API.LogError("[BotWebhook] Failed to load configuration", "error", err.Error())
		return err
	}
	p.configuration = &configuration
	return nil
}

func (p *BotWebhookPlugin) MessageHasBeenPosted(c *plugin.Context, post *model.Post) {
	channel, err := p.API.GetChannel(post.ChannelId)
	if err != nil {
		p.API.LogError("[BotWebhook] Failed to get channel", "error", err.Error())
		return
	}

	for _, bot := range p.activeBots() {
		if post.UserId == bot.BotUserID {
			continue
		}
		if !strings.Contains(channel.Name, bot.BotUserID) {
			continue
		}

		p.API.LogInfo("[BotWebhook] Message to bot detected", "channel", channel.Name, "user", post.UserId, "bot", bot.BotUserID)

		jsonPayload, err := json.Marshal(post)
		if err != nil {
			p.API.LogError("[BotWebhook] Failed to marshal JSON payload", "error", err.Error())
			continue
		}

		req, err := http.NewRequest("POST", bot.WebhookURL, bytes.NewBuffer(jsonPayload))
		if err != nil {
			p.API.LogError("[BotWebhook] Failed to create HTTP request", "error", err.Error())
			continue
		}
		req.Header.Set("Authorization", "Bearer "+bot.BearerToken)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			p.API.LogError("[BotWebhook] Failed to make HTTP request", "error", err.Error())
			continue
		}
		resp.Body.Close()
		p.API.LogInfo("[BotWebhook] Webhook dispatched", "bot", bot.BotUserID, "url", bot.WebhookURL)
		break
	}
}

func (p *BotWebhookPlugin) OnActivate() error {
	return p.OnConfigurationChange()
}

func main() {
	plugin.ClientMain(&BotWebhookPlugin{})
}
