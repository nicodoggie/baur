package cfg

type AppVarResolver struct {
	StrReplacementResolver
}

func NewAppVarResolver(appName string) *AppVarResolver {
	return &AppVarResolver{
		StrReplacementResolver{
			Old: "$" + "APP",
			New: appName,
		},
	}
}
