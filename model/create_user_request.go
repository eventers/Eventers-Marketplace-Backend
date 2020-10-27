package model

type CreateUser struct {
	Data Data `json:"data"`
}

type Data struct {
	User *User `json:"user,omitempty" validate:"required"`
	Auth *Auth `json:"auth,omitempty" validate:"required"`
}

type CreateMarketPlaceUser struct {
	Data struct {
		User *MarketplaceUser `json:"user_marketplace,omitempty" validate:"required"`
		Auth *Auth            `json:"auth,omitempty" validate:"required"`
	} `json:"data"`
}
