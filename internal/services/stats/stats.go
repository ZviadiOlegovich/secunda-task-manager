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

type TaskWithInvalidAssignee struct {
	TaskID     int64
	TaskTitle  string
	TeamID     int64
	AssigneeID int64
}
