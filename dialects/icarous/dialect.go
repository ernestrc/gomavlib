// autogenerated with dialgen. do not edit.

package icarous

import (
	"github.com/gswly/gomavlib"
)

// icarous.xml

type MessageIcarousHeartbeat struct {
	Status uint8
}

func (*MessageIcarousHeartbeat) GetId() uint32 {
	return 42000
}

type MessageIcarousKinematicBands struct {
	Numbands int8 `mavname:"numBands"`
	Type1    uint8
	Min1     float32
	Max1     float32
	Type2    uint8
	Min2     float32
	Max2     float32
	Type3    uint8
	Min3     float32
	Max3     float32
	Type4    uint8
	Min4     float32
	Max4     float32
	Type5    uint8
	Min5     float32
	Max5     float32
}

func (*MessageIcarousKinematicBands) GetId() uint32 {
	return 42001
}

var Dialect = []gomavlib.Message{
	// icarous.xml
	&MessageIcarousHeartbeat{},
	&MessageIcarousKinematicBands{},
}

type ICAROUS_FMS_STATE int

const (
	ICAROUS_FMS_STATE_IDLE     ICAROUS_FMS_STATE = 0
	ICAROUS_FMS_STATE_TAKEOFF  ICAROUS_FMS_STATE = 1
	ICAROUS_FMS_STATE_CLIMB    ICAROUS_FMS_STATE = 2
	ICAROUS_FMS_STATE_CRUISE   ICAROUS_FMS_STATE = 3
	ICAROUS_FMS_STATE_APPROACH ICAROUS_FMS_STATE = 4
	ICAROUS_FMS_STATE_LAND     ICAROUS_FMS_STATE = 5
)

type ICAROUS_TRACK_BAND_TYPES int

const (
	ICAROUS_TRACK_BAND_TYPE_NONE     ICAROUS_TRACK_BAND_TYPES = 0
	ICAROUS_TRACK_BAND_TYPE_NEAR     ICAROUS_TRACK_BAND_TYPES = 1
	ICAROUS_TRACK_BAND_TYPE_RECOVERY ICAROUS_TRACK_BAND_TYPES = 2
)
