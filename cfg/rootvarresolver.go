package cfg

type RootVarResolver struct {
	StrReplacementResolver
}

func NewRootVarResolver(rootPath string) *RootVarResolver {
	return &RootVarResolver{
		StrReplacementResolver{
			Old: "$" + "ROOT",
			New: rootPath,
		},
	}
}
