package model

import (
	"time"
)

type User struct {
	UserID int64 `json:"user_id,omitempty"`

	FirstName    *string `json:"first_name,omitempty"`
	LastName     *string `json:"last_name,omitempty"`
	DisplayName  *string `json:"display_name,omitempty"`
	EmailAddress *string `json:"email_address,omitempty"`

	City       *string `json:"city,omitempty"`
	State      *string `json:"state,omitempty"`
	Country    *string `json:"country,omitempty"`
	Address    *string `json:"address,omitempty"`
	Pincode    *string `json:"pincode,omitempty"`
	ProfilePic *string `json:"profile_pic,omitempty"`

	Provider *string `json:"provider,omitempty"`

	AFirebaseID *string `json:"a_firebase_id,omitempty"`
	AEmail      *string `json:"a_email,omitempty"`
	AName       *string `json:"a_name,omitempty"`
	AImageURL   *string `json:"a_image_url,omitempty"`

	FBFirebaseID *string `json:"fb_firebase_id,omitempty"`
	FBEmail      *string `json:"fb_email,omitempty"`
	FBName       *string `json:"fb_name,omitempty"`
	FBImageURL   *string `json:"fb_image_url,omitempty"`

	GFirebaseID *string `json:"g_firebase_id,omitempty"`
	GEmail      *string `json:"g_email,omitempty"`
	GName       *string `json:"g_name,omitempty"`
	GImageURL   *string `json:"g_image_url,omitempty"`

	PhoneFirebaseID  *string `json:"phone_firebase_id,omitempty"`
	PhoneCountryCode *string `json:"phone_country_code,omitempty"`
	PhoneNumber      *string `json:"phone_number,omitempty"`

	GCMID string `json:"gcm_id,omitempty"`

	IsRegistered bool `json:"is_registered,omitempty"`
	IsActive     bool `json:"is_active,omitempty"`

	CreatedDate *time.Time `json:"created_date,omitempty"`
	UpdatedDate *time.Time `json:"updated_date,omitempty"`

	AccountAddress string `json:"account_address,omitempty"`
}

type UserDevice struct {
	UserID         int64   `json:"user_id,omitempty"`
	IsValid        *bool   `json:"is_valid,omitempty"`
	DeviceID       *string `json:"device_id,omitempty"`
	FCMToken       *string `json:"fcm_token,omitempty"`
	DeviceProvider *string `json:"device_provider,omitempty"`
}

type UserAddress struct {
	UserAddressID      int64   `json:"user_address_id,omitempty"`
	UserID             int64   `json:"user_id,omitempty"`
	AccountAddress     *string `json:"account_address,omitempty"`
	SecurityParaphrase *string `json:"security_paraphrase,omitempty"`
	AccountType        *string `json:"account_type,omitempty"`
	AlgorandHolding    int64   `json:"algorand_holding,omitempty"`
}

type UserConsent struct {
	UserConsentID        int64   `json:"user_consent_id,omitempty"`
	UserID               int64   `json:"user_id,omitempty"`
	FirebaseID           *string `json:"firebase_id,omitempty"`
	PhoneCountryCode     *string `json:"phone_country_code,omitempty"`
	PhoneNumber          *string `json:"phone_number,omitempty"`
	TermsAndConditions   *bool   `json:"terms_and_conditions,omitempty"`
	PrivacyPolicy        *bool   `json:"privacy_policy,omitempty"`
	WhatsappNotification *bool   `json:"whatsapp_notification,omitempty"`
}
