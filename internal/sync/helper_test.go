package sync

// this test helper facilitates unit testing of private functions.

var (
	CalculateIndexPatternDiff       = calculateIndexPatternDiff
	CalculateRoleDiff               = calculateRoleDiff
	GenerateIndexPatterns           = generateIndexPatterns
	GenerateIndexPatternsForGroup   = generateIndexPatternsForGroup
	GenerateIndexPermissionPatterns = generateIndexPermissionPatterns
	GenerateProjectRole             = generateProjectRole
	GenerateRegularGroupRole        = generateRegularGroupRole
	GenerateRoles                   = generateRoles
	HashPrefix                      = hashPrefix
)
