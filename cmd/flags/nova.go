package flags

type ServerListFlags struct {
	Name          *string
	Host          *string
	Status        *[]string
	Flavor        *string
	Project       *string
	All           *bool
	AZ            *string
	Fields        *string
	Verbose       *bool
	Dsc           *bool
	Search        *string
	Watch         *bool
	WatchInterval *uint
	Long          *bool
	// Marker *string
	// Limit  *uint
}

type ServerCreateFlags struct {
	Flavor     *string
	Image      *string
	Nic        *[]string
	VolumeBoot *bool
	VolumeSize *uint16
	VolumeType *string
	AZ         *string
	Min        *uint16
	Max        *uint16
	UserData   *string
	KeyName    *string
	AdminPass  *string
	Wait       *bool
}
type ServerSetFlags struct {
	Name           *string
	Password       *string
	PasswordPrompt *bool
	User           *string
	Status         *string
	Description    *string
}

type ServerDeleteFlags struct {
	Wait *bool
}

type ServerRebootFlags struct {
	Hard *bool
	Wait *bool
}

type ServerResizeFlags struct {
	Confirm *bool
	Revert  *bool
	Flavor  *string
	Wait    *bool
}
type ServerMigrateFlags struct {
	Live         *bool
	BlockMigrate *bool
	Host         *string
	Wait         *bool
}
type ServerRebuildFlags struct {
	Image         *string
	Password      *string
	Name          *string
	UserData      *string
	UserDataUnset *bool
}
type ServerEvacuateFlags struct {
	Password *string
	Host     *string
	Force    *bool
}
type ServerRegionMigrateFlags struct {
	Live   *bool
	Host   *string
	DryRun *bool
}
type ServerMigrationListFlags struct {
	Status        *string
	Type          *string
	Latest        *bool
	Long          *bool
	Watch         *bool
	WatchInterval *uint16
}

type AggregateListFlags struct {
	Long *bool
	Name *string
}
type AggregateCreateFlags struct {
	AZ *string
}

type AZListFlags struct {
	Tree *bool
}

type ComputeServiceListFlags struct {
	Binary *string
	Host   *string
	Zone   *string
	State  *[]string
	Long   *bool
}
type ComputeServiceDisableFlags struct {
	Reason *string
}

type ConsoleLogFlags struct {
	Lines *uint
}

type ServerActionFlags struct {
	Name  *string
	Spend *bool
	Last  *int
	Long  *bool
}

type FlavorListFlags struct {
	Public  *bool
	Name    *string
	MinVcpu *uint64
	MinRam  *uint64
	MinDisk *uint64
	Long    *bool
	Human   *bool
}
type FlavorCreateFlags struct {
	Id   *string
	Name *string

	Disk       *uint
	Swap       *uint
	Private    *bool
	RxtxFactor *float32
	Ephemeral  *uint
	Properties *[]string
	Long       *bool
}
type FlavorSetFlags struct {
	Properties   *[]string
	NoProperties *[]string
}

type HypervisorListFlags struct {
	Name *string

	Type        *string
	WithServers *bool
	Long        *bool
}

type HypervisorShowFlags struct {
	Bar *bool
}
type MigrationListFlags struct {
	Host     *string
	Status   *string
	Instance *string
	Type     *string
	Long     *bool
}

type GroupListFlags struct {
	Long *bool
}
type ServerCreateImageFlags struct {
	Metadata *[]string
}
