package model

type Auth struct {
	DeviceID       *string `json:"device_id,omitempty"`
	TokenID        string  `json:"token_id,omitempty" validate:"required"`
	FCMToken       *string `json:"fcm_token,omitempty"`
	DeviceProvider *string `json:"device_provider,omitempty"`
	AuthType       *string `json:"auth_type,omitempty"`
	PushKey        *string `json:"push_key,omitempty"`
	Status         string  `json:"status,omitempty"`
	OTP            string  `json:"otp,omitempty"`
	UserID         int64   `json:"user_id,omitempty"`
	AccessKey      string  `json:"access_key,omitempty"`
}
