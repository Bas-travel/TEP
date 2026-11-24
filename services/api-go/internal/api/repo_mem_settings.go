package api

import "time"

func (m *memRepo) PlatformSettings() (map[string]PlatformSetting, error) {
	out := make(map[string]PlatformSetting, len(m.platform))
	for k, v := range m.platform {
		out[k] = v
	}
	return out, nil
}

func (m *memRepo) SetPlatformSetting(key, value string, now time.Time) error {
	if m.platform == nil {
		m.platform = map[string]PlatformSetting{}
	}
	at := now
	m.platform[key] = PlatformSetting{Key: key, Value: value, UpdatedAt: &at}
	return nil
}
