package jobs

import (
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

// TranslateUsernameArgs are the arguments for a translate_username job.
type TranslateUsernameArgs struct {
	Username string `json:"username"`
	Region   string `json:"region"`
}

func (TranslateUsernameArgs) Kind() string { return "translate_username" }

func (args TranslateUsernameArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
			ByState: []rivertype.JobState{
				rivertype.JobStateAvailable,
				rivertype.JobStatePending,
				rivertype.JobStateRunning,
				rivertype.JobStateRetryable,
				rivertype.JobStateScheduled,
			},
		},
		MaxAttempts: 3,
	}
}
