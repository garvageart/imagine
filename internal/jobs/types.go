package jobs

// JobCommand represents a command for a worker job. Commands may be shared
// (e.g. "all", "missing", "single") or worker-specific strings.
type JobCommand string

const (
	JobCommandAll     JobCommand = "all"
	JobCommandMissing JobCommand = "missing"
	JobCommandSingle  JobCommand = "single"
)
