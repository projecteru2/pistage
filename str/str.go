package str

// Merge .
func Merge(a, b []string) []string {
	var dic = make(map[string]struct{})
	var strs = make([]string, 0, len(a)+len(b))

	var append = func(ar []string) {
		for _, k := range ar {
			if _, exists := dic[k]; exists {
				continue
			}
			dic[k] = struct{}{}
			strs = append(strs, k)
		}
	}

	append(a)
	append(b)

	return strs
}
