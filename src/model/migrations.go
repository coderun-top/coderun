package model

type Migration struct{
	Name string `gorm:"type:varchar(255);unique"`
}
