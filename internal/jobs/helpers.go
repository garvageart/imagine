package jobs

// small helpers used by workers
// Truncate returns the first n characters of s if longer, otherwise s.
func Truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
