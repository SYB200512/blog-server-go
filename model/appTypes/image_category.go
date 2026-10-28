package appTypes

import "encoding/json"

// ImageCategory 图片用途分类
type ImageCategory int

const (
	Unused       ImageCategory = iota // 未使用
	Cover                             // 封面
	Illustration                      // 正文插图
	Avatar                            // 头像
	System                            // 系统
	Photo                             // 图片墙
)

// MarshalJSON 序列化为字符串，便于前端直接展示
func (c ImageCategory) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

// UnmarshalJSON 反序列化字符串为 ImageCategory
func (c *ImageCategory) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	*c = ToImageCategory(str)
	return nil
}

// String 返回分类的中文名称
func (c ImageCategory) String() string {
	switch c {
	case Unused:
		return "未使用"
	case Cover:
		return "封面"
	case Illustration:
		return "插图"
	case Avatar:
		return "头像"
	case System:
		return "系统"
	case Photo:
		return "图片墙"
	default:
		return "未知分类"
	}
}

// ToImageCategory 将字符串转换为 ImageCategory
func ToImageCategory(str string) ImageCategory {
	switch str {
	case "未使用":
		return Unused
	case "封面":
		return Cover
	case "插图":
		return Illustration
	case "头像":
		return Avatar
	case "系统":
		return System
	case "图片墙":
		return Photo
	default:
		return Unused
	}
}
