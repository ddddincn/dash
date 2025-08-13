package dto

import "dash/consts"

type User struct {
	ID          int32          `json:"id"`
	Username    string         `json:"username"`
	Nickname    string         `json:"nickname"`
	Email       string         `json:"email"`
	Avatar      string         `json:"avatar"`
	Description string         `json:"description"`
	MFAType     consts.MFAType `json:"mfa_type"`
	CreateTime  int64          `json:"create_time"`
	UpdateTime  int64          `json:"update_time"`
}
