package rank

type Priority int

const (
	PriorityDefault = iota * 5
	PriorityVip
	PriorityOG
	PrioritySaint
	PriorityFamous
	PriorityJuniorHelper
	PriorityHelper
	PrioritySeniorModerator
	PriorityAdmin
	PriorityManager
	PriorityOwner = iota * 50000
)
