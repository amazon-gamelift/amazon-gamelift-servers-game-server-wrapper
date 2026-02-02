package orchestration

type EventType int

const (
	PlacementStarting EventType = iota
	PlacementActive
	PlacementDormant
	PlacementTerminating
	PlacementTerminated
)

var eventTypeString = map[EventType]string{
	PlacementStarting:    "PlacementStarting",
	PlacementActive:      "PlacementActive",
	PlacementDormant:     "PlacementDormant",
	PlacementTerminating: "PlacementTerminating",
	PlacementTerminated:  "PlacementTerminated",
}

func (t EventType) String() string {
	return eventTypeString[t]
}
