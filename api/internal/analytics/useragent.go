package analytics

import "github.com/mileusna/useragent"

type uaResult struct {
	Browser        string
	BrowserVersion string
	OS             string
	DeviceType     string
	IsBot          bool
}

// parseUA extrai browser/SO/tipo de dispositivo do cabeçalho User-Agent.
func parseUA(raw string) uaResult {
	ua := useragent.Parse(raw)

	deviceType := "desktop"
	switch {
	case ua.Bot:
		deviceType = "bot"
	case ua.Tablet:
		deviceType = "tablet"
	case ua.Mobile:
		deviceType = "mobile"
	}

	return uaResult{
		Browser:        ua.Name,
		BrowserVersion: ua.Version,
		OS:             ua.OS,
		DeviceType:     deviceType,
		IsBot:          ua.Bot,
	}
}
