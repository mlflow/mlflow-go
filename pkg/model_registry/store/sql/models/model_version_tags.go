package models

// ModelVersionTag mapped from table <model_version_tags>.
//
//revive:disable:exported
type ModelVersionTag struct {
	Key     string `gorm:"column:key;primaryKey"`
	Value   string `gorm:"column:value"`
	Name    string `gorm:"column:name;primaryKey"`
	Version int32  `gorm:"column:version;primaryKey"`
}
