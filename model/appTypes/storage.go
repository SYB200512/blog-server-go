package appTypes

import "encoding/json"

// Storage 图片存储方式
type Storage int

const (
	Local Storage = iota // 本地磁盘
	Cos                  // 腾讯云 COS
)

// MarshalJSON 序列化为字符串，便于前端直接展示
func (s Storage) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// UnmarshalJSON 反序列化字符串为 Storage
func (s *Storage) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	*s = ToStorage(str)
	return nil
}

// String 返回存储方式的中文名称
func (s Storage) String() string {
	switch s {
	case Local:
		return "本地"
	case Cos:
		return "腾讯云COS"
	default:
		return "未知存储"
	}
}

// ToStorage 将字符串转换为 Storage
func ToStorage(str string) Storage {
	switch str {
	case "本地", "local", "Local":
		return Local
	case "腾讯云COS", "cos", "COS", "Cos":
		return Cos
	default:
		return Local
	}
}
