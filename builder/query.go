package builder

type QueryBuilder interface {
	PricingGridQuery() string
	PricingGridCountQuery() string
}
