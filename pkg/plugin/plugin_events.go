package plugin

import (
	"encoding/json"
	"log"
	"strconv"

	"go.acpr.dev/touchportal-golang-sdk/pkg/client"
)

type PluginEvent string

const (
	closePluginEvent = PluginEvent("closePlugin")
	infoEvent        = PluginEvent("info")
	settingsEvent    = PluginEvent("settings")
)

func (p *Plugin) on(event PluginEvent, handler func(event interface{})) {
	p.client.AddMessageHandler(client.ClientMessageType(event), handler)
}

// OnClosePlugin allows the registration of an event handler to the "closePlugin" TouchPortal
// message. A default handler is already in place to close down the plugin itself but you
// may wish to add an additional hook so you can carry out other shutdown tasks.
func (p *Plugin) OnClosePlugin(handler func(event client.ClosePluginMessage)) {
	p.on(closePluginEvent, func(e interface{}) {
		handler(e.(client.ClosePluginMessage))
	})
}

// OnInfo allows the registration of an event handler to the "info" TouchPortal message.
// As the "info" message is only sent as a part of the registration process it is necessary
// to register any handlers before plugin.Register function is called.
func (p *Plugin) OnInfo(handler func(event client.InfoMessage)) {
	p.on(infoEvent, func(e interface{}) {
		handler(e.(client.InfoMessage))
	})
}

// OnInfo allows the registration of an event handler
func (p *Plugin) OnSettings(handler func(settings interface{})) {

}

// onSettings sets up the necessary processing to turn a message containing settings
// into a data structure that can be packed into a user supplied struct.
func (p *Plugin) onSettings(handler func(event client.SettingsMessage)) {
	p.on(settingsEvent, func(e interface{}) {
		msg, ok := e.(client.SettingsMessage)
		if !ok {
			log.Printf("failed to assert event is SettingsMessage: %+v", msg)
			return
		}

		settings := new([]map[string]interface{})
		err := json.Unmarshal(msg.RawValues, settings)
		if err != nil {
			log.Printf("failed to unmarshal settings from raw data: %v", err)
			return
		}

		flat := make(map[string]interface{}, len(*settings))

		// the touchportal message format is aware of numbers but still sends them as
		// text over json so we do some clunky recasting of that info here.
		for _, setting := range *settings {
			for k, v := range setting {
				i, err := strconv.ParseInt(v.(string), 10, 64)
				if err == nil {
					flat[k] = i
				} else {
					flat[k] = v
				}
			}
		}

		msg.Values = flat

		handler(msg)
	})
}