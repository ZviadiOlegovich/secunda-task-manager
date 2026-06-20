package stats

type TeamStat struct {
	TeamID       int64
	TeamName     string
	MemberCount  int
	DoneLastWeek int
}

type TopUser struct {
	TeamID    int64
	TeamName  string
	UserID    int64
	UserName  string
	TaskCount int
	Rank      int
}
