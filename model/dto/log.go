package dto

import "dash/consts"

type Log struct {
	ID         int64          `json:"id"`
	LogKey     string         `json:"logKey"`
	LogType    consts.LogType `json:"type"`
	Content    string         `json:"content"`
	IPAddress  string         `json:"ipAddress"`
	CreateTime int64          `json:"createTime"`
}
