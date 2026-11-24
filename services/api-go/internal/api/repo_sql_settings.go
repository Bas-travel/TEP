package api

import "time"

func (t *sqlTx) PlatformSettings() (map[string]PlatformSetting, error) {
	rows, err := t.query(`SELECT setting_key, setting_value, updated_at FROM platform_settings`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]PlatformSetting{}
	for rows.Next() {
		var p PlatformSetting
		var at time.Time
		if err := rows.Scan(&p.Key, &p.Value, &at); err != nil {
			return nil, err
		}
		p.UpdatedAt = &at
		out[p.Key] = p
	}
	return out, rows.Err()
}

func (t *sqlTx) SetPlatformSetting(key, value string, now time.Time) error {
	_, err := t.exec(`INSERT INTO platform_settings (setting_key, setting_value, updated_at) VALUES (?,?,?)
		ON DUPLICATE KEY UPDATE setting_value=VALUES(setting_value), updated_at=VALUES(updated_at)`, key, value, utc(now))
	return err
}
