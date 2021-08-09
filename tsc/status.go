package tsc

const CmdStatus = "\x1B!?" // 1

const (
	StatusOpened Status = 1 << iota
	StatusPaperJam
	StatusNoPaper
	StatusNoRibbon
	StatusPaused
	StatusPrinting
	_
	StatusOther
	StatusReady Status = 0
)

type Status uint8

func (s Status) String() string {
	if s == 0 {
		return "ready"
	}
	res := ""
	if s&StatusOpened != 0 {
		res += "opened, "
	}
	if s&StatusPaperJam != 0 {
		res += "paper jam, "
	}
	if s&StatusNoPaper != 0 {
		res += "no paper, "
	}
	if s&StatusNoRibbon != 0 {
		res += "no ribbon, "
	}
	if s&StatusPaused != 0 {
		res += "paused, "
	}
	if s&StatusPrinting != 0 {
		res += "printing, "
	}
	if len(res) > 2 {
		return res[:len(res)-2]
	}
	return "unknown error"
}
