package dto

type AccessToken struct {
	AccessToken string `json:"access_token"`
	ExpiredIn   int    `json:"expired_in"`
}
