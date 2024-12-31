package flags

type PruneVolumeFlags struct {
	Name   *string
	Status *string
	Type   *string
	All    *bool
	Yes    *bool
	Marker *string
	Limit  *uint
}
