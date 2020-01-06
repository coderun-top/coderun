package model

import (
	"database/sql/driver"
	"encoding/json"	
)

type MapType map[string]string
type StringArrType []string 

func (u *MapType) Scan(src interface{}) error { 
	var dat map[string]string
	json.Unmarshal(src.([]uint8), &dat);
	*u = dat
	return nil 
}

func (u MapType) Value() (driver.Value, error)  { 
	mjson,_ :=json.Marshal(u)
	return mjson,nil
}


func (u *StringArrType) Scan(src interface{}) error { 
	var dat []string
	json.Unmarshal(src.([]uint8), &dat);
	*u = dat
	return nil 
}

func (u StringArrType) Value() (driver.Value, error)  { 
	// return string(u), nil 
	mjson,_ :=json.Marshal(u)
	// mString :=string(mjson)
	return mjson,nil
}
