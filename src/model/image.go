package model

// Image 对应于数据库中的 'images' 表
type Image struct {
	ID        int    `json:"id" gorm:"column:id;primaryKey"`
	Name      string `json:"name" gorm:"column:name"`
	URL       string `json:"url" gorm:"column:url"`
	LocalPath string `json:"local_path" gorm:"column:local_path"`
	IsLocal   int    `json:"is_local" gorm:"column:is_local"`
	IsOss     int    `json:"is_oss" gorm:"column:is_oss"`
	Status    string `json:"status" gorm:"column:status"`
}

func (Image) TableName() string {
	return "images"
}

// PublicImageMetadata 是可匿名公开访问的图片元数据，刻意不包含
// LocalPath 等服务器内部信息。
type PublicImageMetadata struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	URL    string `json:"url"`
	Status string `json:"status"`
}

// ToPublicMetadata 转换为公开接口可安全返回的元数据。
func (i *Image) ToPublicMetadata() PublicImageMetadata {
	return PublicImageMetadata{
		ID:     i.ID,
		Name:   i.Name,
		URL:    i.URL,
		Status: i.Status,
	}
}

type QueryImageResponse struct {
	Images []Image
	Total  int64
}
