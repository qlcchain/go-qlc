package apis

func toStringPoint(s string) *string {
	if s != "" {
		return &s
	}
	return nil
}
